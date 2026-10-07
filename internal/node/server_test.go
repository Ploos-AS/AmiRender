package node

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Ploos-AS/AmiRender/internal/farm"
)

func testCoordinator(t *testing.T, engine string) (farm.Coordinator, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = Serve(ln) }()

	r := farm.NewRegistry()
	if err := r.Register(farm.Node{
		ID: "node-01", Arch: "amd64",
		Capabilities: []farm.Capability{{Engine: engine}},
	}); err != nil {
		ln.Close()
		t.Fatal(err)
	}
	return farm.Coordinator{
		Registry:  r,
		Endpoints: map[string]string{"node-01": ln.Addr().String()},
		Client:    farm.TCPWorkerClient{},
	}, func() { _ = ln.Close() }
}

func TestProductionHandlerNullRoundTrip(t *testing.T) {
	c, closeServer := testCoordinator(t, "null")
	defer closeServer()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := c.Render(ctx, farm.RenderJob{
		ID: "m1-e2e", Engine: "null", Output: "result.dat",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "COMPLETE" || result.JobID != "m1-e2e" || result.Output != "result.dat" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestProductionHandlerPOVRayCreatesPNG(t *testing.T) {
	if _, err := exec.LookPath("povray"); err != nil {
		t.Skip("povray not installed")
	}

	dir := t.TempDir()
	scene := filepath.Join(dir, "scene.pov")
	output := filepath.Join(dir, "frame.png")
	source := `camera { location <0, 0, -3> look_at <0, 0, 0> }
light_source { <-2, 3, -4> color rgb <1, 1, 1> }
sphere { <0, 0, 0>, 1 pigment { color rgb <0.7, 0.7, 0.7> } }
`
	if err := os.WriteFile(scene, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}

	c, closeServer := testCoordinator(t, "povray")
	defer closeServer()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := c.Render(ctx, farm.RenderJob{
		ID: "pov-e2e", Engine: "povray", Scene: scene, Output: output, Width: 160, Height: 120,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "COMPLETE" || result.Output == "" || result.Output == output {
		t.Fatalf("unexpected result: %#v", result)
	}
	if !strings.HasPrefix(filepath.Base(filepath.Dir(result.Output)), "amirender-asset-") {
		t.Fatalf("output is not worker-staged: %q", result.Output)
	}
	defer os.RemoveAll(filepath.Dir(result.Output))
	info, err := os.Stat(result.Output)
	if err != nil {
		t.Fatalf("render output missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("render output is empty")
	}
}

func TestPOVRayUploadRenderDownloadRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("povray"); err != nil {
		t.Skip("povray not installed")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() { _ = Serve(ln) }()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(bufio.NewReader(conn))

	source := `camera { location <0, 0, -3> look_at <0, 0, 0> }
light_source { <-2, 3, -4> color rgb <1, 1, 1> }
sphere { <0, 0, 0>, 1 pigment { color rgb <0.7, 0.7, 0.7> } }
`
	if err := enc.Encode(map[string]string{
		"type": "UPLOAD", "name": "scene.pov",
		"data": base64.StdEncoding.EncodeToString([]byte(source)),
	}); err != nil {
		t.Fatal(err)
	}
	var staged stagedResult
	if err := dec.Decode(&staged); err != nil {
		t.Fatal(err)
	}
	if staged.Type != "STAGED" || staged.Asset == "" {
		t.Fatalf("unexpected staging result: %#v", staged)
	}
	defer os.RemoveAll(filepath.Dir(staged.Asset))

	if err := enc.Encode(map[string]any{
		"type": "SUBMIT",
		"job": map[string]any{
			"id": "pov-wire-e2e", "engine": "povray", "scene": staged.Asset,
			"output": "RAM:frame.png", "width": 64, "height": 48,
		},
	}); err != nil {
		t.Fatal(err)
	}
	var rendered farm.WorkerResult
	if err := dec.Decode(&rendered); err != nil {
		t.Fatal(err)
	}
	if rendered.Type != "COMPLETE" || rendered.Output == "" || rendered.Output == "RAM:frame.png" {
		t.Fatalf("unexpected render result: %#v", rendered)
	}
	defer os.RemoveAll(filepath.Dir(rendered.Output))

	if err := enc.Encode(map[string]string{"type": "DOWNLOAD", "asset": rendered.Output}); err != nil {
		t.Fatal(err)
	}
	var downloaded stagedResult
	if err := dec.Decode(&downloaded); err != nil {
		t.Fatal(err)
	}
	if downloaded.Type != "DATA" {
		t.Fatalf("unexpected download result: %#v", downloaded)
	}
	png, err := base64.StdEncoding.DecodeString(downloaded.Data)
	if err != nil {
		t.Fatal(err)
	}
	signature := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	if len(png) <= len(signature) || string(png[:len(signature)]) != string(signature) {
		t.Fatalf("download is not a PNG: %x", png[:min(len(png), len(signature))])
	}
}

func TestDownloadReturnsStagedAsset(t *testing.T) {
	dir, err := os.MkdirTemp("", "amirender-asset-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "frame.png")
	want := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a}
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatal(err)
	}

	server, client := net.Pipe()
	defer client.Close()
	go Handle(server)

	enc := json.NewEncoder(client)
	dec := json.NewDecoder(bufio.NewReader(client))
	if err := enc.Encode(map[string]string{"type": "DOWNLOAD", "asset": path}); err != nil {
		t.Fatal(err)
	}
	var got stagedResult
	if err := dec.Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "DATA" || got.Asset != path {
		t.Fatalf("unexpected download result: %#v", got)
	}
	decoded, err := base64.StdEncoding.DecodeString(got.Data)
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != string(want) {
		t.Fatalf("download mismatch: %v", decoded)
	}

	if err := enc.Encode(map[string]string{"type": "DOWNLOAD", "asset": "/etc/passwd"}); err != nil {
		t.Fatal(err)
	}
	var rejected stagedResult
	if err := dec.Decode(&rejected); err != nil {
		t.Fatal(err)
	}
	if rejected.Type != "FAILED" {
		t.Fatalf("unsafe download was not rejected: %#v", rejected)
	}
}

func TestUploadStagesAssetAndRejectsTraversal(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() { _ = Serve(ln) }()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(bufio.NewReader(conn))

	data := base64.StdEncoding.EncodeToString([]byte("camera {}\\n"))
	if err := enc.Encode(map[string]string{"type": "UPLOAD", "name": "scene.pov", "data": data}); err != nil {
		t.Fatal(err)
	}
	var staged stagedResult
	if err := dec.Decode(&staged); err != nil {
		t.Fatal(err)
	}
	if staged.Type != "STAGED" || staged.Asset == "" {
		t.Fatalf("unexpected staging result: %#v", staged)
	}
	defer os.RemoveAll(filepath.Dir(staged.Asset))
	got, err := os.ReadFile(staged.Asset)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "camera {}\\n" {
		t.Fatalf("unexpected staged data %q", got)
	}

	if err := enc.Encode(map[string]string{"type": "UPLOAD", "name": "../escape.pov", "data": data}); err != nil {
		t.Fatal(err)
	}
	var rejected stagedResult
	if err := dec.Decode(&rejected); err != nil {
		t.Fatal(err)
	}
	if rejected.Type != "FAILED" {
		t.Fatalf("traversal upload was not rejected: %#v", rejected)
	}
}

func TestChunkedAssetRoundTripAndOffsetValidation(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	go Handle(server)

	enc := json.NewEncoder(client)
	dec := json.NewDecoder(bufio.NewReader(client))
	want := make([]byte, 9000)
	for i := range want {
		want[i] = byte(i % 251)
	}

	if err := enc.Encode(map[string]any{
		"type": "UPLOAD_BEGIN", "name": "large.bin", "size": len(want),
	}); err != nil {
		t.Fatal(err)
	}
	var begun stagedResult
	if err := dec.Decode(&begun); err != nil {
		t.Fatal(err)
	}
	if begun.Type != "STAGING" || begun.Asset == "" || begun.Size != int64(len(want)) {
		t.Fatalf("unexpected begin result: %#v", begun)
	}
	defer os.RemoveAll(filepath.Dir(begun.Asset))

	offset := 0
	for offset < len(want) {
		end := offset + 4096
		if end > len(want) {
			end = len(want)
		}
		if err := enc.Encode(map[string]any{
			"type": "UPLOAD_CHUNK", "asset": begun.Asset, "offset": offset,
			"data": base64.StdEncoding.EncodeToString(want[offset:end]),
		}); err != nil {
			t.Fatal(err)
		}
		var ack stagedResult
		if err := dec.Decode(&ack); err != nil {
			t.Fatal(err)
		}
		if ack.Type != "CHUNKED" || ack.Offset != int64(end) {
			t.Fatalf("unexpected chunk acknowledgement: %#v", ack)
		}
		offset = end
	}

	if err := enc.Encode(map[string]any{
		"type": "UPLOAD_CHUNK", "asset": begun.Asset, "offset": 1,
		"data": base64.StdEncoding.EncodeToString([]byte("bad")),
	}); err != nil {
		t.Fatal(err)
	}
	var rejected stagedResult
	if err := dec.Decode(&rejected); err != nil {
		t.Fatal(err)
	}
	if rejected.Type != "FAILED" || rejected.Error != "unexpected chunk offset" {
		t.Fatalf("bad offset was not rejected: %#v", rejected)
	}

	var got []byte
	for offset = 0; ; {
		if err := enc.Encode(map[string]any{
			"type": "DOWNLOAD_CHUNK", "asset": begun.Asset, "offset": offset, "size": 4096,
		}); err != nil {
			t.Fatal(err)
		}
		var part stagedResult
		if err := dec.Decode(&part); err != nil {
			t.Fatal(err)
		}
		if part.Type != "DATA" || part.Offset != int64(offset) {
			t.Fatalf("unexpected download chunk: %#v", part)
		}
		decoded, err := base64.StdEncoding.DecodeString(part.Data)
		if err != nil {
			t.Fatal(err)
		}
		if int64(len(decoded)) != part.Size {
			t.Fatalf("chunk size mismatch: decoded=%d declared=%d", len(decoded), part.Size)
		}
		got = append(got, decoded...)
		offset += len(decoded)
		if part.EOF {
			break
		}
	}
	if string(got) != string(want) {
		t.Fatalf("chunked round trip mismatch: got=%d want=%d", len(got), len(want))
	}

	if err := enc.Encode(map[string]any{
		"type": "DOWNLOAD_CHUNK", "asset": begun.Asset, "offset": 0, "size": 4097,
	}); err != nil {
		t.Fatal(err)
	}
	var oversized stagedResult
	if err := dec.Decode(&oversized); err != nil {
		t.Fatal(err)
	}
	if oversized.Type != "FAILED" {
		t.Fatalf("oversized download chunk was not rejected: %#v", oversized)
	}
}
func TestUploadEndRequiresDeclaredSizeAndSingleCompletion(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	go Handle(server)
	enc := json.NewEncoder(client)
	dec := json.NewDecoder(bufio.NewReader(client))
	send := func(m map[string]any) stagedResult {
		t.Helper()
		if err := enc.Encode(m); err != nil {
			t.Fatal(err)
		}
		var result stagedResult
		if err := dec.Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	begun := send(map[string]any{"type": "UPLOAD_BEGIN", "name": "partial.bin", "size": 9000})
	if begun.Type != "STAGING" {
		t.Fatalf("begin: %#v", begun)
	}
	defer os.RemoveAll(filepath.Dir(begun.Asset))
	payload := base64.StdEncoding.EncodeToString(make([]byte, 4096))
	for _, offset := range []int{0, 4096} {
		ack := send(map[string]any{"type": "UPLOAD_CHUNK", "asset": begun.Asset, "offset": offset, "data": payload})
		if ack.Type != "CHUNKED" {
			t.Fatalf("chunk: %#v", ack)
		}
	}
	if result := send(map[string]any{"type": "UPLOAD_END", "asset": begun.Asset, "size": 8192}); result.Type != "FAILED" || result.Error != "upload size mismatch" {
		t.Fatalf("changed declared size accepted: %#v", result)
	}
	if result := send(map[string]any{"type": "UPLOAD_END", "asset": begun.Asset, "size": 9000}); result.Type != "FAILED" || result.Error != "incomplete upload" {
		t.Fatalf("partial upload accepted: %#v", result)
	}
	last := base64.StdEncoding.EncodeToString(make([]byte, 808))
	if result := send(map[string]any{"type": "UPLOAD_CHUNK", "asset": begun.Asset, "offset": 8192, "data": last}); result.Type != "CHUNKED" {
		t.Fatalf("final chunk: %#v", result)
	}
	if result := send(map[string]any{"type": "UPLOAD_END", "asset": begun.Asset, "size": 9000}); result.Type != "STAGED" {
		t.Fatalf("completion: %#v", result)
	}
	if result := send(map[string]any{"type": "UPLOAD_END", "asset": begun.Asset, "size": 9000}); result.Type != "FAILED" {
		t.Fatalf("duplicate completion accepted: %#v", result)
	}
}
