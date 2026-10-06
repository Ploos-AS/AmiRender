package node

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Ploos-AS/AmiRender/internal/farm"
)

func TestProductionHandlerNullRoundTrip(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() { _ = Serve(ln) }()

	r := farm.NewRegistry()
	if err := r.Register(farm.Node{
		ID: "node-01", Arch: "amd64",
		Capabilities: []farm.Capability{{Engine: "null"}},
	}); err != nil {
		t.Fatal(err)
	}
	c := farm.Coordinator{
		Registry: r,
		Endpoints: map[string]string{"node-01": ln.Addr().String()},
		Client: farm.TCPWorkerClient{},
	}
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
