package install

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kumneger0/clispot/internal/config"
)

type ffmpegBinConfig struct {
	ffmpegURL  string
	ffprobeURL string
	ffmpeg     string
	ffprobe    string
	isArchive  bool
}

var (
	checksumURL string = "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/checksums.sha256"

	ffmpegBinConfigs = map[string]*ffmpegBinConfig{
		"linux_amd64": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz",
			ffmpeg:    "ffmpeg",
			ffprobe:   "ffprobe",
			isArchive: true,
		},
		"windows_amd64": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
			ffmpeg:    "ffmpeg.exe",
			ffprobe:   "ffprobe.exe",
			isArchive: true,
		},
		"windows_arm": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-winarm64-gpl.zip",
			ffmpeg:    "ffmpeg.exe",
			ffprobe:   "ffprobe.exe",
			isArchive: true,
		},
	}
)

func FFmpeg(ctx context.Context) (*ResolvedInstall, error) {
	if runtime.GOOS == "darwin" {
		fmt.Println("install command is not supported on this platform please install manually from https://www.ffmpeg.org/download.html")
		return nil, nil
	}
	plat, err := detectPlatform()

	if err != nil {
		panic(err)
	}

	ffmpegBinConfig := ffmpegBinConfigs[plat]

	if ffmpegBinConfig == nil {
		panic("Failed to found the binary url for your OS")
	}

	splitted := strings.Split(ffmpegBinConfig.ffmpegURL, "/")

	filename := splitted[len(splitted)-1]

	ffmpegDirectory := filepath.Join(config.GetCacheDir(runtime.GOOS), filename)

	checksumDir := filepath.Join(config.GetCacheDir(runtime.GOOS), "ffmpeg-checksum")

	result, err := install(ffmpegBinConfig.ffmpegURL, checksumURL, ffmpegDirectory, checksumDir, filename)

	if err != nil {
		panic(err)
	}

	targetDir := filepath.Join(config.GetCacheDir(runtime.GOOS), "ffmpeg")

	err = extractBinaries(targetDir, ffmpegDirectory)

	if err != nil {
		panic(err)
	}

	return result, nil
}
