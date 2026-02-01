package install

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kumneger0/clispot/internal/config"
)

var assetMap = map[string]string{
	"linux_amd64":   "yt-dlp_linux",
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
	if runtime.GOOS == "darwin" {
		fmt.Println("install command is not supported on this platform please install manually from https://github.com/yt-dlp/yt-dlp/releases")
		return nil, nil
	}

	ytDlpDirectory := filepath.Join(config.GetCacheDir(runtime.GOOS), "yt-dlp")
	checksumDir := filepath.Join(config.GetCacheDir(runtime.GOOS), "yt-dlp-checksum")
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

	isAlreadyLatestVersion, err := isAlreadyInstalledAndLatestVersion(ytDlpDirectory, rel.TagName)

	if err == nil && isAlreadyLatestVersion {
		return &ResolvedInstall{
			Executable: ytDlpDirectory,
			Version:    rel.TagName,
			FromCache:  true,
			Downloaded: true,
		}, nil
	}

	checksumDownloadURL, err := findDownloadURL(rel, "SHA2-256SUMS")

	if err != nil {
		panic(err)
	}

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

func isAlreadyInstalledAndLatestVersion(path, newVersion string) (bool, error) {
	fileStat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fileStat.IsDir() {
		return false, errors.New("oops this is not  a file")
	}

	cmd := exec.Command(path, "--version")

	version, err := cmd.Output()
	if err != nil {
		return false, err
	}

	oldVersionSlice := strings.Split(string(version), ".")
	newVersionSlice := strings.Split(string(newVersion), ".")
	newDate := sliceToTime(newVersionSlice)
	oldDate := sliceToTime(oldVersionSlice)

	return !newDate.After(oldDate), nil
}

func sliceToTime(d []string) time.Time {
	year, err := strconv.Atoi(d[0])

	if err != nil {
		panic(err)
	}
	month, err := strconv.Atoi(d[1])
	if err != nil {
		panic(err)
	}
	date, err := strconv.Atoi(strings.ReplaceAll(d[2], "\n", ""))

	if err != nil {
		panic(err)
	}

	return time.Date(year, time.Month(month), date, 0, 0, 0, 0, time.UTC)
}
