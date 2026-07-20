package backend

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func StartBackend(fs embed.FS) (*exec.Cmd, error) {
	data, err := fs.ReadFile("main")
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(data)
	actualHash := fmt.Sprintf("%x", hash)

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(cacheDir, "yt-music-tui")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	binaryPath := filepath.Join(appDir, "backend")

	file, err := os.ReadFile(binaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(appDir, 0755)
			if err != nil {
				return nil, err
			}
			return writeBinaryToCacheFolderAndRun(data, binaryPath)
		}
	}

	expectedHash := fmt.Sprintf("%x", sha256.Sum256(file))
	if actualHash != expectedHash {
		if err := os.Remove(binaryPath); err != nil {
			return nil, fmt.Errorf("backend binary integrity check failed")
		}
		return writeBinaryToCacheFolderAndRun(data, binaryPath)
	}

	cmd := exec.Command(binaryPath)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

func writeBinaryToCacheFolderAndRun(data []byte, binaryPath string) (*exec.Cmd, error) {
	if err := os.WriteFile(binaryPath, data, 0755); err != nil {
		return nil, err
	}

	if err := os.Chmod(binaryPath, 0755); err != nil {
		return nil, err
	}

	cmd := exec.Command(binaryPath)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}
