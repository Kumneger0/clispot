package install

import (
	"context"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kumneger0/clispot/internal/config"
)

var assetMap = map[string]string{
	"linux_amd64":   "yt-dlp_linux",
	"linux_arm64":   "yt-dlp_linux",
	"darwin_amd64":  "yt-dlp",
	"darwin_arm64":  "yt-dlp",
	"windows_amd64": "yt-dlp.exe",
	"windows_arm64": "yt-dlp_arm64.exe",
}

func findDownloadURL(rel *githubRelease, want string) (string, error) {
	for _, a := range rel.Assets {
		if a.Name == want {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no asset found for %s", want)
}

var YtDlp = func(ctx context.Context) (*ResolvedInstall, error) {
	plat, err := detectPlatform()
	if err != nil {
		panic(err)
	}

	githubOwner := "yt-dlp"
	githubRepo := "yt-dlp"

	rel, err := fetchLatestRelease(githubOwner, githubRepo)
	if err != nil {
		panic(err)
	}

	want := assetMap[plat]

	url, err := findDownloadURL(rel, want)
	if err != nil {
		panic(err)
	}

	checksumDownloadURL, err := findDownloadURL(rel, "SHA2-256SUMS")

	if err != nil {
		panic(err)
	}

	ytDlpDirectory := filepath.Join(config.GetCacheDir(runtime.GOOS), "yt-dlp")
	checksumDir := filepath.Join(config.GetCacheDir(runtime.GOOS), "yt-dlp-checksum")
	return install(url, checksumDownloadURL, ytDlpDirectory, checksumDir, want)
}

func getExpectedHash(checksumData string, filename string) ([]byte, error) {
	lines := strings.SplitSeq(checksumData, "\n")

	for line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}

		hash := fields[0]
		name := fields[1]

		if name == filename {
			return hex.DecodeString(hash)
		}
	}
	return nil, fmt.Errorf("checksum not found for %s", filename)
}
