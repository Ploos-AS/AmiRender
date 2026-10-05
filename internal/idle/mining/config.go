package mining

import (
	"fmt"

	"github.com/Ploos-AS/AmiRender/internal/idle"
)

type Config struct {
	Executable string
	Pool       string
	Wallet     string
	ExtraArgs  []string
}

func Monero(c Config) (idle.Workload, error) {
	if c.Executable == "" || c.Pool == "" || c.Wallet == "" {
		return nil, fmt.Errorf("monero requires executable, pool and wallet")
	}
	args := []string{"-o", c.Pool, "-u", c.Wallet, "--coin", "monero"}
	args = append(args, c.ExtraArgs...)
	return idle.NewProcessWorkload("monero", c.Executable, args), nil
}

func Verus(c Config) (idle.Workload, error) {
	if c.Executable == "" || c.Pool == "" || c.Wallet == "" {
		return nil, fmt.Errorf("verus requires executable, pool and wallet")
	}
	args := []string{"-o", c.Pool, "-u", c.Wallet}
	args = append(args, c.ExtraArgs...)
	return idle.NewProcessWorkload("verus", c.Executable, args), nil
}
