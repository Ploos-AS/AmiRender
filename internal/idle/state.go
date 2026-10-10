package idle

import "fmt"

type State string

const (
	StateIdle       State = "IDLE"
	StateDelay      State = "IDLE_DELAY"
	StateWorkload   State = "IDLE_WORKLOAD"
	StatePreempting State = "PREEMPTING"
	StateRendering  State = "RENDERING"
)

type Machine struct{ State State }

func NewMachine() *Machine { return &Machine{State: StateIdle} }

func (m *Machine) IdleTimerStarted() error {
	if m.State != StateIdle {
		return fmt.Errorf("cannot start idle delay from %s", m.State)
	}
	m.State = StateDelay
	return nil
}

func (m *Machine) IdleTimerExpired() error {
	if m.State != StateDelay {
		return fmt.Errorf("cannot start workload from %s", m.State)
	}
	m.State = StateWorkload
	return nil
}

func (m *Machine) RenderArrived() error {
	switch m.State {
	case StateIdle, StateDelay:
		m.State = StateRendering
	case StateWorkload:
		m.State = StatePreempting
	default:
		return fmt.Errorf("cannot accept render from %s", m.State)
	}
	return nil
}

func (m *Machine) Preempted() error {
	if m.State != StatePreempting {
		return fmt.Errorf("cannot finish preemption from %s", m.State)
	}
	m.State = StateRendering
	return nil
}

func (m *Machine) RenderQueueDrained() error {
	if m.State != StateRendering {
		return fmt.Errorf("cannot drain render queue from %s", m.State)
	}
	m.State = StateIdle
	return nil
}
