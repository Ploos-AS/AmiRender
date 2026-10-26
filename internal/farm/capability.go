package farm

type Capability struct {
	Engine string   `json:"engine"`
	Modes  []string `json:"modes,omitempty"`
}

type Node struct {
	ID           string       `json:"id"`
	Arch         string       `json:"arch"`
	Capabilities []Capability `json:"capabilities"`
}

func (n Node) Supports(engine string) bool {
	for _, c := range n.Capabilities {
		if c.Engine == engine {
			return true
		}
	}
	return false
}
