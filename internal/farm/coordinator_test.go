package farm

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestCoordinatorRoutesJobToCapableWorker(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	received := make(chan RenderJob, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		var msg struct {
			Type string    `json:"type"`
			Job  RenderJob `json:"job"`
		}
		if json.NewDecoder(c).Decode(&msg) != nil {
			return
		}
		received <- msg.Job
		_ = json.NewEncoder(c).Encode(WorkerResult{
			Type: "COMPLETE", JobID: msg.Job.ID, Engine: msg.Job.Engine, Output: msg.Job.Output,
		})
	}()

	r := NewRegistry()
	if err := r.Register(Node{
		ID: "rock5-01", Arch: "arm64",
		Capabilities: []Capability{{Engine: "povray"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(Node{
		ID: "gpu-01", Arch: "amd64",
		Capabilities: []Capability{{Engine: "blender-cycles"}},
	}); err != nil {
		t.Fatal(err)
	}

	coordinator := Coordinator{
		Registry: r,
		Endpoints: map[string]string{
			"rock5-01": ln.Addr().String(),
		},
		Client: TCPWorkerClient{},
	}
	job := RenderJob{
		ID: "frame-0042", Engine: "povray", Scene: "m1-smoke.pov",
		Output: "frame0042.png", Frame: 42, Width: 160, Height: 120,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := coordinator.Render(ctx, job)
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "COMPLETE" || result.JobID != job.ID {
		t.Fatalf("unexpected result: %#v", result)
	}

	select {
	case got := <-received:
		if got.Engine != "povray" || got.ID != job.ID {
			t.Fatalf("wrong job reached worker: %#v", got)
		}
	case <-ctx.Done():
		t.Fatal("worker did not receive job")
	}
}
