package install

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

	ytDlpDirectory := filepath.Join(config.GetCacheDir(runtime.GOOS), "yt-dlp")

	_, err = os.Create(ytDlpDirectory)

	if err != nil {
		panic(err)
	}

	err = download(url, ytDlpDirectory)

	if err != nil {
		panic(err)
	}

	checksumDownloadURL, err := findDownloadURL(rel, "SHA2-256SUMS")

	if err != nil {
		panic(err)
	}

	checksumDir := filepath.Join(config.GetCacheDir(runtime.GOOS), "yt-dlp-checksum")

	err = download(checksumDownloadURL, checksumDir)

	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(checksumDir)

	if err != nil {
		panic(err)
	}

	expectedHash, err := getExpectedHash(string(data), want)

	if err != nil {
		panic(err)
	}

	calculatedHash, err := calculateFileCheckSum(ytDlpDirectory)

	if err != nil {
		panic(err)
	}

	isValid := validateChecksum(calculatedHash, []byte(expectedHash))

	if isValid {
		slog.Error("check sum validation gone wrong")
		fmt.Fprintln(os.Stderr, "there is something wrong while validating the downloaded binary could download the binary by ur sel")
		os.Exit(1)
	}

	return &ResolvedInstall{
		Executable: ytDlpDirectory,
		Version:    "fuck it i don't care but im sure it is the latest version",
		FromCache:  false,
		Downloaded: true,
	}, nil
}

func getExpectedHash(checksumData string, filename string) (string, error) {
	lines := strings.SplitSeq(checksumData, "\n")

	for line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}

		hash := fields[0]
		name := fields[1]

		if name == filename {
			return hash, nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", filename)
}
