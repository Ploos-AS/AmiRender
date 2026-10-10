package engine

import "context"

type Job struct {
	ID      string
	Scene   string
	Output  string
	Frame   int
	Width   int
	Height  int
	Options map[string]string
}

type Result struct {
	Output string
}

type Engine interface {
	Name() string
	Probe(context.Context) error
	Render(context.Context, Job) (Result, error)
}
