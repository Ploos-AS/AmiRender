package farm

import (
	"errors"
	"testing"
)

func TestDispatcherRoutesPOVRayToRegisteredWorker(t *testing.T) {
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

	a, err := (Dispatcher{Registry: r}).Assign("frame-0042", "povray")
	if err != nil {
		t.Fatal(err)
	}
	if a.Node.ID != "rock5-01" || a.Engine != "povray" || a.JobID != "frame-0042" {
		t.Fatalf("unexpected assignment: %#v", a)
	}
}

func TestDispatcherRejectsMissingCapability(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(Node{
		ID: "rock5-01", Arch: "arm64",
		Capabilities: []Capability{{Engine: "povray"}},
	}); err != nil {
		t.Fatal(err)
	}
	_, err := (Dispatcher{Registry: r}).Assign("frame-0001", "blender-cycles")
	if !errors.Is(err, ErrNoCapableNode) {
		t.Fatalf("want ErrNoCapableNode, got %v", err)
	}
}
