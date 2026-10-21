package mining

import "fmt"

// Config describes an owner-supplied external idle workload.
// AmiRender stores policy and configuration only; execution is delegated
// outside the render node.
type Config struct {
	Enabled    bool
	Kind       string
	Executable string
	Pool       string
	Wallet     string
	CPUPercent int
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	switch c.Kind {
	case "monero", "verus":
	default:
		return fmt.Errorf("unsupported mining workload %q", c.Kind)
	}
	if c.Executable == "" || c.Pool == "" || c.Wallet == "" {
		return fmt.Errorf("%s requires owner-supplied executable, pool and wallet", c.Kind)
	}
	if c.CPUPercent < 1 || c.CPUPercent > 100 {
		return fmt.Errorf("cpu percent must be between 1 and 100")
	}
	return nil
}
