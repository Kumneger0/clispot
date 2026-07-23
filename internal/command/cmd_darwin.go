//go:build darwin

package command

import (
	"context"
	"os"
	"os/exec"
)

func ExecCommand(ctx context.Context, command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	return cmd, nil
}

func KillProcess(p *os.Process) error {
	return p.Kill()
}
