package node

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Ploos-AS/AmiRender/internal/engine"
	"github.com/Ploos-AS/AmiRender/internal/engine/povray"
	"github.com/Ploos-AS/AmiRender/internal/farm"
)

type uploadSession struct {
	size int64
	created time.Time
}

const uploadSessionTTL = 30 * time.Minute

var uploadSessions = struct {
	sync.Mutex
	expected map[string]uploadSession
}{expected: make(map[string]uploadSession)}

const maxUploadBytes = 1024 * 1024
const maxMessageBytes = 2 * 1024 * 1024

type message struct {
	Type   string          `json:"type"`
	Job    json.RawMessage `json:"job"`
	Name   string          `json:"name,omitempty"`
	Data   string          `json:"data,omitempty"`
	Asset  string          `json:"asset,omitempty"`
	Offset int64           `json:"offset,omitempty"`
	Size   int64           `json:"size,omitempty"`
}

type stagedResult struct {
	Type   string `json:"type"`
	Asset  string `json:"asset,omitempty"`
	Error  string `json:"error,omitempty"`
	Data   string `json:"data,omitempty"`
	Offset int64  `json:"offset,omitempty"`
	Size   int64  `json:"size,omitempty"`
	EOF    bool   `json:"eof,omitempty"`
}

func Handle(c net.Conn) {
	defer c.Close()
	s := bufio.NewScanner(c)
	s.Buffer(make([]byte, 64*1024), maxMessageBytes)
	for s.Scan() {
		var m message
		if json.Unmarshal(s.Bytes(), &m) != nil {
			writeResult(c, farm.WorkerResult{Type: "FAILED", Error: "bad request"})
			continue
		}
		if m.Type == "UPLOAD" {
			handleUpload(c, m)
			continue
		}
		if m.Type == "UPLOAD_BEGIN" {
			handleUploadBegin(c, m)
			continue
		}
		if m.Type == "UPLOAD_CHUNK" {
			handleUploadChunk(c, m)
			continue
		}
		if m.Type == "UPLOAD_END" {
			handleUploadEnd(c, m)
			continue
		}
		if m.Type == "DOWNLOAD_CHUNK" {
			handleDownloadChunk(c, m)
			continue
		}
		if m.Type == "DOWNLOAD" {
			handleDownload(c, m)
			continue
		}
		if m.Type != "SUBMIT" {
			writeResult(c, farm.WorkerResult{Type: "FAILED", Error: "bad request"})
			continue
		}
		var j farm.RenderJob
		if json.Unmarshal(m.Job, &j) != nil {
			writeResult(c, farm.WorkerResult{Type: "FAILED", Error: "bad job"})
			continue
		}
		switch j.Engine {
		case "null":
			writeResult(c, farm.WorkerResult{
				Type: "COMPLETE", JobID: j.ID, Engine: "null", Output: j.Output,
			})
		case "povray":
			uploadSessions.Lock()
			_, incomplete := uploadSessions.expected[filepath.Clean(j.Scene)]
			uploadSessions.Unlock()
			if incomplete {
				writeResult(c, farm.WorkerResult{
					Type: "FAILED", JobID: j.ID, Engine: "povray", Error: "scene upload incomplete",
				})
				continue
			}
			renderPOVRay(c, j)
		default:
			writeResult(c, farm.WorkerResult{
				Type: "FAILED", JobID: j.ID, Engine: j.Engine, Error: "unsupported engine",
			})
		}
	}
}

func validStagedAsset(asset string) (string, bool) {
	clean := filepath.Clean(asset)
	dir := filepath.Dir(clean)
	return clean, asset != "" && strings.HasPrefix(filepath.Base(dir), "amirender-asset-")
}

func cleanupExpiredUploads(now time.Time) {
	var expired []string
	uploadSessions.Lock()
	for path, session := range uploadSessions.expected {
		if now.Sub(session.created) >= uploadSessionTTL {
			delete(uploadSessions.expected, path)
			expired = append(expired, path)
		}
	}
	uploadSessions.Unlock()
	for _, path := range expired {
		_ = os.RemoveAll(filepath.Dir(path))
	}
}

func handleUploadBegin(c net.Conn, m message) {
	cleanupExpiredUploads(time.Now())
	if filepath.Base(m.Name) != m.Name || m.Name == "." || m.Name == "" || m.Size <= 0 || m.Size > maxUploadBytes {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "invalid upload"})
		return
	}
	dir, err := os.MkdirTemp("", "amirender-asset-")
	if err != nil {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "staging failed"})
		return
	}
	path := filepath.Join(dir, m.Name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		_ = os.RemoveAll(dir)
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "staging failed"})
		return
	}
	_ = file.Close()
	uploadSessions.Lock()
	uploadSessions.expected[path] = uploadSession{size: m.Size, created: time.Now()}
	uploadSessions.Unlock()
	_ = json.NewEncoder(c).Encode(stagedResult{Type: "STAGING", Asset: path, Size: m.Size})
}

func handleUploadChunk(c net.Conn, m message) {
	clean, ok := validStagedAsset(m.Asset)
	if !ok || m.Offset < 0 {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "unsafe asset reference"})
		return
	}
	data, err := base64.StdEncoding.DecodeString(m.Data)
	if err != nil || len(data) == 0 || len(data) > 4096 || m.Offset+int64(len(data)) > maxUploadBytes {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "invalid chunk"})
		return
	}
	file, err := os.OpenFile(clean, os.O_WRONLY, 0600)
	if err != nil {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "asset unavailable"})
		return
	}
	info, statErr := file.Stat()
	if statErr != nil || info.Size() != m.Offset {
		_ = file.Close()
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "unexpected chunk offset"})
		return
	}
	_, err = file.WriteAt(data, m.Offset)
	_ = file.Close()
	if err != nil {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "staging failed"})
		return
	}
	_ = json.NewEncoder(c).Encode(stagedResult{Type: "CHUNKED", Asset: clean, Offset: m.Offset + int64(len(data))})
}

func handleUploadEnd(c net.Conn, m message) {
	clean, ok := validStagedAsset(m.Asset)
	if !ok || m.Size <= 0 || m.Size > maxUploadBytes {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "invalid upload completion"})
		return
	}
	uploadSessions.Lock()
	expected, exists := uploadSessions.expected[clean]
	uploadSessions.Unlock()
	if !exists || expected.size != m.Size {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "upload size mismatch"})
		return
	}
	info, err := os.Stat(clean)
	if err != nil {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "asset unavailable"})
		return
	}
	if info.Size() != m.Size {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "incomplete upload", Size: info.Size()})
		return
	}
	uploadSessions.Lock()
	delete(uploadSessions.expected, clean)
	uploadSessions.Unlock()
	_ = json.NewEncoder(c).Encode(stagedResult{Type: "STAGED", Asset: clean, Size: info.Size()})
}

func handleDownloadChunk(c net.Conn, m message) {
	clean, ok := validStagedAsset(m.Asset)
	if !ok || m.Offset < 0 || m.Size <= 0 || m.Size > 4096 {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "invalid chunk request"})
		return
	}
	file, err := os.Open(clean)
	if err != nil {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "asset unavailable"})
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || m.Offset > info.Size() {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "invalid chunk offset"})
		return
	}
	buf := make([]byte, m.Size)
	n, err := file.ReadAt(buf, m.Offset)
	if err != nil && n == 0 {
		if m.Offset == info.Size() {
			_ = json.NewEncoder(c).Encode(stagedResult{Type: "DATA", Asset: clean, Offset: m.Offset, Size: 0, EOF: true})
			return
		}
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "asset unavailable"})
		return
	}
	_ = json.NewEncoder(c).Encode(stagedResult{Type: "DATA", Asset: clean, Offset: m.Offset, Size: int64(n), Data: base64.StdEncoding.EncodeToString(buf[:n]), EOF: m.Offset+int64(n) >= info.Size()})
}

func handleUpload(c net.Conn, m message) {
	if filepath.Base(m.Name) != m.Name || m.Name == "." || m.Name == "" {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "unsafe asset name"})
		return
	}
	data, err := base64.StdEncoding.DecodeString(m.Data)
	if err != nil || len(data) == 0 || len(data) > maxUploadBytes {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "invalid asset data"})
		return
	}
	dir, err := os.MkdirTemp("", "amirender-asset-")
	if err != nil {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "staging failed"})
		return
	}
	path := filepath.Join(dir, m.Name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		_ = os.RemoveAll(dir)
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "staging failed"})
		return
	}
	_ = json.NewEncoder(c).Encode(stagedResult{Type: "STAGED", Asset: path})
}

func handleDownload(c net.Conn, m message) {
	clean := filepath.Clean(m.Asset)
	dir := filepath.Dir(clean)
	if m.Asset == "" || !strings.HasPrefix(filepath.Base(dir), "amirender-asset-") {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "unsafe asset reference"})
		return
	}
	data, err := os.ReadFile(clean)
	if err != nil || len(data) == 0 || len(data) > maxUploadBytes {
		_ = json.NewEncoder(c).Encode(stagedResult{Type: "FAILED", Error: "asset unavailable"})
		return
	}
	_ = json.NewEncoder(c).Encode(stagedResult{
		Type: "DATA", Asset: clean, Data: base64.StdEncoding.EncodeToString(data),
	})
}

func renderPOVRay(c net.Conn, j farm.RenderJob) {
	dir, err := os.MkdirTemp("", "amirender-asset-")
	if err != nil {
		writeResult(c, farm.WorkerResult{
			Type: "FAILED", JobID: j.ID, Engine: "povray", Error: "output staging failed",
		})
		return
	}
	output := filepath.Join(dir, "frame.png")
	e := povray.New("")
	result, err := e.Render(context.Background(), engine.Job{
		ID: j.ID, Scene: j.Scene, Output: output, Frame: j.Frame, Width: j.Width, Height: j.Height,
	})
	if err != nil {
		_ = os.RemoveAll(dir)
		writeResult(c, farm.WorkerResult{
			Type: "FAILED", JobID: j.ID, Engine: "povray", Error: err.Error(),
		})
		return
	}
	writeResult(c, farm.WorkerResult{
		Type: "COMPLETE", JobID: j.ID, Engine: "povray", Output: result.Output,
	})
}

func writeResult(c net.Conn, result farm.WorkerResult) {
	_ = json.NewEncoder(c).Encode(result)
}

func Serve(ln net.Listener) error {
	for {
		c, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}
		go Handle(c)
	}
}
