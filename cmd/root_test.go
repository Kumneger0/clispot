package cmd

//test for autoInstallDeps

import (
	"context"
	"testing"

	"github.com/lrstanley/go-ytdlp"
	"github.com/stretchr/testify/assert"
)

func TestAutoInstallDeps_NoAutoInstallable(t *testing.T) {
	deps := []DebsCheckResult{
		{ToolName: YtDlp, Installed: false},
		{ToolName: Ffmpeg, Installed: false},
	}

	err := autoInstallDeps(deps, "plan9")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no missing dependencies can be installed automatically")
}

func TestAutoInstallDeps_UserDeclines(t *testing.T) {
	origConfirm := confirmInstall
	defer func() { confirmInstall = origConfirm }()

	confirmInstall = func() (bool, error) {
		return false, nil
	}

	deps := []DebsCheckResult{
		{ToolName: YtDlp, Installed: false},
		{ToolName: Ffmpeg, Installed: false},
	}

	err := autoInstallDeps(deps, "linux_amd64")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user declined")
}

func TestAutoInstallDeps_InstallYtDlp(t *testing.T) {
	origConfirm := confirmInstall
	origInstall := installYtDlp
	defer func() {
		confirmInstall = origConfirm
		installYtDlp = origInstall
	}()

	confirmInstall = func() (bool, error) { return true, nil }

	called := false

	installAllDeps = func(ctx context.Context) ([]*ytdlp.ResolvedInstall, error) {
		called = true
		return []*ytdlp.ResolvedInstall{
			{Executable: "yt-dlp", Version: "2025.01"},
		}, nil
	}

	installYtDlp = func(ctx context.Context) (*ytdlp.ResolvedInstall, error) {
		called = true
		return &ytdlp.ResolvedInstall{
			Executable: "yt-dlp",
			Version:    "2025.01",
		}, nil
	}

	deps := []DebsCheckResult{
		{ToolName: YtDlp, Installed: false},
	}

	err := autoInstallDeps(deps, "linux_amd64")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestAutoInstallDeps_InstallFFmpeg(t *testing.T) {
	origConfirm := confirmInstall
	origInstall := installFFmpeg

	defer func() {
		confirmInstall = origConfirm
		installFFmpeg = origInstall
	}()

	confirmInstall = func() (bool, error) { return true, nil }

	called := false

	installAllDeps = func(ctx context.Context) ([]*ytdlp.ResolvedInstall, error) {
		called = true
		return []*ytdlp.ResolvedInstall{
			{Executable: "yt-dlp", Version: "2025.01"},
		}, nil
	}

	installFFmpeg = func(ctx context.Context) (*ytdlp.ResolvedInstall, error) {
		called = true
		return &ytdlp.ResolvedInstall{
			Executable: "ffmpeg",
			Version:    "6.1",
		}, nil
	}

	deps := []DebsCheckResult{
		{ToolName: Ffmpeg, Installed: false},
	}

	err := autoInstallDeps(deps, "linux_amd64")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestAutoInstallDeps_InstallAll(t *testing.T) {
	origConfirm := confirmInstall
	origInstallAll := installAllDeps
	defer func() {
		confirmInstall = origConfirm
		installAllDeps = origInstallAll
	}()

	confirmInstall = func() (bool, error) { return true, nil }

	called := false
	installAllDeps = func(ctx context.Context) ([]*ytdlp.ResolvedInstall, error) {
		called = true
		return []*ytdlp.ResolvedInstall{
			{Executable: "yt-dlp", Version: "1"},
			{Executable: "ffmpeg", Version: "6"},
		}, nil
	}

	deps := []DebsCheckResult{
		{ToolName: YtDlp, Installed: false},
		{ToolName: Ffmpeg, Installed: false},
	}

	err := autoInstallDeps(deps, "linux_amd64")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestAutoInstallDeps_WindowsARM_MixedSupport(t *testing.T) {
	origConfirm := confirmInstall
	origInstallFFmpeg := installFFmpeg
	origInstallYtDlp := installYtDlp
	defer func() {
		confirmInstall = origConfirm
		installFFmpeg = origInstallFFmpeg
		installYtDlp = origInstallYtDlp
	}()

	confirmInstall = func() (bool, error) { return true, nil }

	ffmpegCalled := false
	ytCalled := false

	installFFmpeg = func(ctx context.Context) (*ytdlp.ResolvedInstall, error) {
		ffmpegCalled = true
		return &ytdlp.ResolvedInstall{
			Executable: "ffmpeg",
			Version:    "6.1",
		}, nil
	}

	installYtDlp = func(ctx context.Context) (*ytdlp.ResolvedInstall, error) {
		ytCalled = true
		return &ytdlp.ResolvedInstall{
			Executable: "yt-dlp",
			Version:    "2025.01",
		}, nil
	}

	deps := []DebsCheckResult{
		{ToolName: YtDlp, Installed: false},
		{ToolName: Ffmpeg, Installed: false},
	}

	err := autoInstallDeps(deps, "windows_arm")

	assert.NoError(t, err)
	assert.True(t, ffmpegCalled, "ffmpeg should be installed on windows_arm")
	assert.False(t, ytCalled, "yt-dlp should NOT be auto-installed on windows_arm")
}

func TestAutoInstallDeps_DarwinARM_MixedSupport(t *testing.T) {
	origConfirm := confirmInstall
	origInstallFFmpeg := installFFmpeg
	origInstallYtDlp := installYtDlp
	defer func() {
		confirmInstall = origConfirm
		installFFmpeg = origInstallFFmpeg
		installYtDlp = origInstallYtDlp
	}()

	confirmInstall = func() (bool, error) { return true, nil }

	ffmpegCalled := false
	ytCalled := false

	installYtDlp = func(ctx context.Context) (*ytdlp.ResolvedInstall, error) {
		ytCalled = true
		return &ytdlp.ResolvedInstall{
			Executable: "yt-dlp",
			Version:    "2025.01",
		}, nil
	}

	installFFmpeg = func(ctx context.Context) (*ytdlp.ResolvedInstall, error) {
		ffmpegCalled = true
		return &ytdlp.ResolvedInstall{
			Executable: "ffmpeg",
			Version:    "6.1",
		}, nil
	}

	deps := []DebsCheckResult{
		{ToolName: YtDlp, Installed: false},
		{ToolName: Ffmpeg, Installed: false},
	}

	err := autoInstallDeps(deps, "darwin_arm64")

	assert.NoError(t, err)
	assert.True(t, ytCalled, "yt-dlp should be installed on darwin_arm64")
	assert.False(t, ffmpegCalled, "ffmpeg should NOT be auto-installed on darwin_arm64")
}
