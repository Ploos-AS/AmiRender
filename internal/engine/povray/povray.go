package povray

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/Ploos-AS/AmiRender/internal/engine"
)

type Engine struct {
	Binary string
}

func New(binary string) *Engine {
	if binary == "" {
		binary = "povray"
	}
	return &Engine{Binary: binary}
}

func (e *Engine) Name() string { return "povray" }

func (e *Engine) Probe(ctx context.Context) error {
	_, err := exec.LookPath(e.Binary)
	if err != nil {
		return fmt.Errorf("povray executable %q not found: %w", e.Binary, err)
	}
	return nil
}

func (e *Engine) Render(ctx context.Context, job engine.Job) (engine.Result, error) {
	if job.Scene == "" || job.Output == "" {
		return engine.Result{}, fmt.Errorf("scene and output are required")
	}

	args := []string{"+I" + job.Scene, "+O" + job.Output, "-D"}
	if job.Width > 0 {
		args = append(args, "+W"+strconv.Itoa(job.Width))
	}
	if job.Height > 0 {
		args = append(args, "+H"+strconv.Itoa(job.Height))
	}

	cmd := exec.CommandContext(ctx, e.Binary, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return engine.Result{}, fmt.Errorf("povray render failed: %w: %s", err, string(out))
	}
	return engine.Result{Output: job.Output}, nil
}
