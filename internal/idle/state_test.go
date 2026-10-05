package idle

import "testing"

func TestIdleWorkloadPreemptLifecycle(t *testing.T) {
	m := NewMachine()
	steps := []struct{ want State; fn func() error }{
		{StateDelay, m.IdleTimerStarted},
		{StateWorkload, m.IdleTimerExpired},
		{StatePreempting, m.RenderArrived},
		{StateRendering, m.Preempted},
		{StateIdle, m.RenderQueueDrained},
	}
	for _, s := range steps {
		if err := s.fn(); err != nil { t.Fatal(err) }
		if m.State != s.want { t.Fatalf("want %s, got %s", s.want, m.State) }
	}
}

func TestRenderSkipsIdleWorkloadDuringDelay(t *testing.T) {
	m := NewMachine()
	if err := m.IdleTimerStarted(); err != nil { t.Fatal(err) }
	if err := m.RenderArrived(); err != nil { t.Fatal(err) }
	if m.State != StateRendering { t.Fatalf("want rendering, got %s", m.State) }
}
