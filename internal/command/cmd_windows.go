//go:build windows

package command

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func ExecCommand(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	return cmd, nil
}

func KillProcess(p *os.Process) error {
	return windows.GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, uint32(p.Pid))
}
