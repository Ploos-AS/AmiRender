package idle

import (
	"context"
	"errors"
	"testing"
)

type fakeWorkload struct{ started, stopped bool }

func (f *fakeWorkload) Name() string                { return "fake" }
func (f *fakeWorkload) Start(context.Context) error { f.started = true; return nil }
func (f *fakeWorkload) Stop() error                 { f.stopped = true; return nil }

func TestDisabledByDefault(t *testing.T) {
	f := &fakeWorkload{}
	m := Manager{Workload: f}
	if !errors.Is(m.StartIfIdle(context.Background()), ErrDisabled) {
		t.Fatal("idle workload must be disabled by default")
	}
	if f.started {
		t.Fatal("disabled workload started")
	}
}

func TestRenderCanPreemptIdleWork(t *testing.T) {
	f := &fakeWorkload{}
	m := Manager{Config: Config{Enabled: true}, Workload: f}
	if err := m.StartIfIdle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := m.Preempt(); err != nil {
		t.Fatal(err)
	}
	if !f.stopped {
		t.Fatal("idle workload was not stopped")
	}
}
