//go:build !windows

package awake

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// Start impide la suspension hasta que ctx termine. Usa caffeinate (macOS) o
// systemd-inhibit (Linux); devuelve error si la herramienta no existe.
func Start(ctx context.Context) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "caffeinate", "-d", "-i")
	case "linux":
		cmd = exec.CommandContext(ctx, "systemd-inhibit", "--what=idle:sleep", "--who=move-mouse",
			"--why=move-mouse activo", "sleep", "infinity")
	default:
		return fmt.Errorf("keep-awake no soportado en %s", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("keep-awake: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
