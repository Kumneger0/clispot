package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/gofrs/flock"
	"github.com/kumneger0/clispot/cmd"
)

var (
	version = ""
)

func main() {
	lockFilePath := filepath.Join(os.TempDir(), "clispot.lock")
	fileLock := flock.New(lockFilePath)
	locked, err := fileLock.TryLock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error trying to acquire lock: %v\n", err)
		os.Exit(1)
	}
	if !locked {
		showAnotherProcessIsRunning(lockFilePath)
		os.Exit(1)
	}
	defer func() {
		if err := fileLock.Unlock(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not unlock file: %v\n", err)
		}
		_ = os.Remove(lockFilePath)
	}()

	pid := os.Getpid()
	if err := os.WriteFile(lockFilePath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write PID to lock file: %v\n", err)
	}
	err = cmd.Execute(version)
	if err != nil {
		slog.Error(err.Error())
	}
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func showAnotherProcessIsRunning(lockFilePath string) {
	pidBytes, readErr := os.ReadFile(lockFilePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading lock file: %v\n", readErr)
		os.Exit(1)
	}
	pid, parseErr := strconv.Atoi(string(pidBytes))

	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error parsing PID from lock file: %v\n", parseErr)
		os.Exit(1)
	}

	if !isProcessRunning(pid) {
		fmt.Fprintf(os.Stderr, "Another instance of clispot is not running (stale lock file for PID %d).\n", pid)
		fmt.Fprintf(os.Stderr, "Please try removing %s and running again if this persists.\n", lockFilePath)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Another instance of clispot is already running (PID: %d).\n", pid)
}
