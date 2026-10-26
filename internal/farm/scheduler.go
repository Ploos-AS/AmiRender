package farm

import "errors"

var ErrNoCapableNode = errors.New("no capable node")

// SelectNode is deliberately simple in M1: first capable node wins.
// Later schedulers can add load, benchmark weights, locality and GPU preference
// without changing the job or engine contracts.
func SelectNode(nodes []Node, engine string) (Node, error) {
	for _, n := range nodes {
		if n.Supports(engine) {
			return n, nil
		}
	}
	return Node{}, ErrNoCapableNode
}
