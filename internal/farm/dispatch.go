package farm

import "fmt"

type Assignment struct {
	JobID  string
	Engine string
	Node   Node
}

type Dispatcher struct {
	Registry *Registry
}

func (d Dispatcher) Assign(jobID, engine string) (Assignment, error) {
	if d.Registry == nil {
		return Assignment{}, fmt.Errorf("registry is required")
	}
	if jobID == "" || engine == "" {
		return Assignment{}, fmt.Errorf("job id and engine are required")
	}
	candidates := d.Registry.NodesFor(engine)
	node, err := SelectNode(candidates, engine)
	if err != nil {
		return Assignment{}, err
	}
	return Assignment{JobID: jobID, Engine: engine, Node: node}, nil
}
