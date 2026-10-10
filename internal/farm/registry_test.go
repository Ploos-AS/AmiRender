package farm

import "testing"

func TestRegistryFiltersByCapability(t *testing.T) {
	r := NewRegistry()
	for _, n := range []Node{
		{ID: "rock5-01", Arch: "arm64", Capabilities: []Capability{{Engine: "povray"}}},
		{ID: "gpu-01", Arch: "amd64", Capabilities: []Capability{{Engine: "povray"}, {Engine: "blender-cycles"}}},
	} {
		if err := r.Register(n); err != nil {
			t.Fatal(err)
		}
	}
	pov := r.NodesFor("povray")
	if len(pov) != 2 {
		t.Fatalf("want 2 POV-Ray nodes, got %d", len(pov))
	}
	blender := r.NodesFor("blender-cycles")
	if len(blender) != 1 || blender[0].ID != "gpu-01" {
		t.Fatalf("unexpected Blender nodes: %#v", blender)
	}
}

func TestReregistrationUpdatesCapabilities(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(Node{ID: "node-01", Arch: "arm64", Capabilities: []Capability{{Engine: "povray"}}}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(Node{ID: "node-01", Arch: "arm64", Capabilities: []Capability{{Engine: "blender-cycles"}}}); err != nil {
		t.Fatal(err)
	}
	if len(r.NodesFor("povray")) != 0 || len(r.NodesFor("blender-cycles")) != 1 {
		t.Fatal("re-registration did not replace capabilities")
	}
}
