package idle

import (
	"context"
	"errors"
	"time"
)

var ErrDisabled = errors.New("idle workload disabled")

type Config struct {
	Enabled    bool
	IdleDelay  time.Duration
	CPUPercent int
}

type Workload interface {
	Name() string
	Start(context.Context) error
	Stop() error
}

type Manager struct {
	Config   Config
	Workload Workload
}

func (m *Manager) StartIfIdle(ctx context.Context) error {
	if !m.Config.Enabled {
		return ErrDisabled
	}
	return m.Workload.Start(ctx)
}

// Preempt is called before render work is dispatched.
// Idle work must never block rendering.
func (m *Manager) Preempt() error {
	if m.Workload == nil {
		return nil
	}
	return m.Workload.Stop()
}
