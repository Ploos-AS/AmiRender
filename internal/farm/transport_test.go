package farm

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestTCPWorkerClientRoundTrip(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

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
		_ = json.NewEncoder(c).Encode(WorkerResult{
			Type: "COMPLETE", JobID: msg.Job.ID, Engine: msg.Job.Engine, Output: msg.Job.Output,
		})
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := (TCPWorkerClient{}).Render(ctx, ln.Addr().String(), RenderJob{
		ID: "frame-0042", Engine: "povray", Scene: "scene.pov", Output: "frame0042.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "COMPLETE" || result.JobID != "frame-0042" || result.Engine != "povray" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
