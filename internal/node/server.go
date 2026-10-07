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

	"github.com/Ploos-AS/AmiRender/internal/engine"
	"github.com/Ploos-AS/AmiRender/internal/engine/povray"
	"github.com/Ploos-AS/AmiRender/internal/farm"
)

const maxUploadBytes = 1024 * 1024
const maxMessageBytes = 2 * 1024 * 1024

type message struct {
	Type  string          `json:"type"`
	Job   json.RawMessage `json:"job"`
	Name  string          `json:"name,omitempty"`
	Data  string          `json:"data,omitempty"`
	Asset string          `json:"asset,omitempty"`
}

type stagedResult struct {
	Type  string `json:"type"`
	Asset string `json:"asset,omitempty"`
	Error string `json:"error,omitempty"`
	Data  string `json:"data,omitempty"`
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
			renderPOVRay(c, j)
		default:
			writeResult(c, farm.WorkerResult{
				Type: "FAILED", JobID: j.ID, Engine: j.Engine, Error: "unsupported engine",
			})
		}
	}
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
	e := povray.New("")
	result, err := e.Render(context.Background(), engine.Job{
		ID: j.ID, Scene: j.Scene, Output: j.Output, Frame: j.Frame, Width: j.Width, Height: j.Height,
	})
	if err != nil {
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
