package farm

import (
	"fmt"
	"sort"
	"sync"
)

type Registry struct {
	mu    sync.RWMutex
	nodes map[string]Node
}

func NewRegistry() *Registry {
	return &Registry{nodes: make(map[string]Node)}
}

func (r *Registry) Register(node Node) error {
	if node.ID == "" {
		return fmt.Errorf("node id is required")
	}
	if len(node.Capabilities) == 0 {
		return fmt.Errorf("node %q has no capabilities", node.ID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[node.ID] = node
	return nil
}

func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, id)
}

func (r *Registry) NodesFor(engine string) []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Node
	for _, node := range r.nodes {
		if node.Supports(engine) {
			out = append(out, node)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
