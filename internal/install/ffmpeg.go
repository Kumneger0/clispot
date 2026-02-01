package install

import (
	"context"
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

	ffmpegBinConfigs = map[string]ffmpegBinConfig{
		"darwin_amd64": {
			ffmpegURL:  "https://evermeet.cx/ffmpeg/getrelease/ffmpeg",
			ffprobeURL: "https://evermeet.cx/ffmpeg/getrelease/ffprobe",
			ffmpeg:     "ffmpeg",
			ffprobe:    "ffprobe",
			isArchive:  false,
		},
		"linux_amd64": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz",
			ffmpeg:    "ffmpeg",
			ffprobe:   "ffprobe",
			isArchive: true,
		},
		"linux_arm64": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz",
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
	plat, err := detectPlatform()

	if err != nil {
		panic(err)
	}

	splitted := strings.Split(ffmpegBinConfigs[plat].ffmpegURL, "/")

	filename := splitted[len(splitted)-1]

	ffmpegDirectory := filepath.Join(config.GetCacheDir(runtime.GOOS), filename)

	checksumDir := filepath.Join(config.GetCacheDir(runtime.GOOS), "ffmpeg-checksum")

	result, err := install(ffmpegBinConfigs[plat].ffmpegURL, checksumURL, ffmpegDirectory, checksumDir, filename)

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
