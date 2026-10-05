package farm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
)

type RenderJob struct {
	ID     string `json:"id"`
	Engine string `json:"engine"`
	Scene  string `json:"scene"`
	Output string `json:"output"`
	Frame  int    `json:"frame,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type WorkerResult struct {
	Type   string `json:"type"`
	JobID  string `json:"job_id"`
	Engine string `json:"engine,omitempty"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

type WorkerClient interface {
	Render(context.Context, string, RenderJob) (WorkerResult, error)
}

type TCPWorkerClient struct{}

func (TCPWorkerClient) Render(ctx context.Context, address string, job RenderJob) (WorkerResult, error) {
	var result WorkerResult
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return result, err
	}
	defer conn.Close()

	msg := struct {
		Type string    `json:"type"`
		Job  RenderJob `json:"job"`
	}{Type: "SUBMIT", Job: job}
	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		return result, err
	}
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&result); err != nil {
		return result, err
	}
	switch result.Type {
	case "COMPLETE":
		return result, nil
	case "FAILED":
		return result, fmt.Errorf("worker failed: %s", result.Error)
	default:
		return result, fmt.Errorf("unexpected worker response %q", result.Type)
	}
}
