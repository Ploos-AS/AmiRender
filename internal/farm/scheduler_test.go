package farm

import "testing"

func TestSelectNodeByEngineCapability(t *testing.T) {
	nodes := []Node{
		{ID: "arm-01", Arch: "arm64", Capabilities: []Capability{{Engine: "povray"}}},
		{ID: "gpu-01", Arch: "amd64", Capabilities: []Capability{{Engine: "povray"}, {Engine: "blender-cycles"}}},
	}
	n, err := SelectNode(nodes, "povray")
	if err != nil || n.ID != "arm-01" {
		t.Fatalf("unexpected selection: %#v %v", n, err)
	}
	n, err = SelectNode(nodes, "blender-cycles")
	if err != nil || n.ID != "gpu-01" {
		t.Fatalf("unexpected Blender selection: %#v %v", n, err)
	}
}
