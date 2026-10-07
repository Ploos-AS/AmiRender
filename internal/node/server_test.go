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
	if result.Type != "COMPLETE" || result.Output != output {
		t.Fatalf("unexpected result: %#v", result)
	}
	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("render output missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("render output is empty")
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
