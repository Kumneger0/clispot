//go:build darwin

package command

import (
	"os"
	"os/exec"
)

func ExecCommand(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	return cmd, nil
}

func KillProcess(p *os.Process) error {
	return p.Kill()
}
