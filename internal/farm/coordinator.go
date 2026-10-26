package farm

import (
	"context"
	"fmt"
)

type EndpointNode struct {
	Node
	Address string
}

type Coordinator struct {
	Registry  *Registry
	Endpoints map[string]string
	Client    WorkerClient
}

func (c Coordinator) Render(ctx context.Context, job RenderJob) (WorkerResult, error) {
	if c.Registry == nil || c.Client == nil {
		return WorkerResult{}, fmt.Errorf("registry and worker client are required")
	}
	assignment, err := (Dispatcher{Registry: c.Registry}).Assign(job.ID, job.Engine)
	if err != nil {
		return WorkerResult{}, err
	}
	address := c.Endpoints[assignment.Node.ID]
	if address == "" {
		return WorkerResult{}, fmt.Errorf("node %q has no endpoint", assignment.Node.ID)
	}
	return c.Client.Render(ctx, address, job)
}
