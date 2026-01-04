//go:build windows

package command

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func ExecCommand(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP,
	}
	return cmd, nil
}

func KillProcess(p *os.Process) error {
	if err := p.Kill(); err != nil {
		slog.Error(err.Error())
		return windows.GenerateConsoleCtrlEvent(windows.CTRL_C_EVENT, uint32(p.Pid))
	}
	return nil
}
