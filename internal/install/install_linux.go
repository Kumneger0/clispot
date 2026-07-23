//go:build linux

package install

import (
	"errors"
	"os/exec"

	"github.com/kumneger0/clispot/internal/types"
)

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func FFmpegInstallCommand() ([]types.InstallStep, error) {
	switch {
	case commandExists("apt"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"apt", "update"},
			},
			{
				Command: "sudo",
				Args:    []string{"apt", "install", "-y", "ffmpeg"},
			},
		}, nil

	case commandExists("dnf"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"dnf", "install", "-y", "ffmpeg"},
			},
		}, nil

	case commandExists("yum"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"yum", "install", "-y", "ffmpeg"},
			},
		}, nil

	case commandExists("pacman"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"pacman", "-S", "--noconfirm", "ffmpeg"},
			},
		}, nil

	case commandExists("zypper"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"zypper", "--non-interactive", "install", "ffmpeg"},
			},
		}, nil

	case commandExists("apk"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"apk", "add", "ffmpeg"},
			},
		}, nil

	case commandExists("xbps-install"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"xbps-install", "-Sy", "ffmpeg"},
			},
		}, nil

	case commandExists("emerge"):
		return []types.InstallStep{
			{
				Command: "sudo",
				Args:    []string{"emerge", "media-video/ffmpeg"},
			},
		}, nil

	case commandExists("nix"):
		return []types.InstallStep{
			{
				Command: "nix",
				Args:    []string{"profile", "install", "nixpkgs#ffmpeg"},
			},
		}, nil

	default:
		return nil, errors.New("unsupported Linux distribution")
	}
}
