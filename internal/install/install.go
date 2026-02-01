package install

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	pb "github.com/schollz/progressbar/v3"
)

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type ResolvedInstall struct {
	Executable string
	Version    string
	FromCache  bool
	Downloaded bool
}

func fetchLatestRelease(githubOwner, githubRepo string) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubOwner, githubRepo)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "go-ytdlp-fetcher")
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var rel githubRelease
	if err := json.NewDecoder(res.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func download(url string, outputPath string, shouldShowProgress bool) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if shouldShowProgress {
		bar := pb.DefaultBytes(
			resp.ContentLength,
			"downloading",
		)
		_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	} else {
		_, err = io.Copy(out, resp.Body)
	}

	if err != nil {
		return err
	}

	err = os.Chmod(outputPath, 0755)
	return err
}

func detectPlatform() (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	if goos == "linux" && goarch == "arm" {
		goarch = "arm64"
	}

	key := fmt.Sprintf("%s_%s", goos, goarch)
	if _, ok := assetMap[key]; !ok {
		return "", fmt.Errorf("unsupported platform: %s", key)
	}
	return key, nil
}

func install(url, checksumURL, pathToFile, checksumDir, githubReleaseFileName string) (*ResolvedInstall, error) {
	if err := os.MkdirAll(filepath.Dir(pathToFile), 0755); err != nil {
		panic(err)
	}

	err := download(url, pathToFile, true)

	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(filepath.Dir(checksumDir), 0755); err != nil {
		panic(err)
	}

	err = download(checksumURL, checksumDir, false)

	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(checksumDir)

	if err != nil {
		panic(err)
	}

	expectedHash, err := getExpectedHash(string(data), githubReleaseFileName)

	if err != nil {
		panic(err)
	}

	calculatedHash, err := calculateFileCheckSum(pathToFile)

	if err != nil {
		panic(err)
	}

	isValid := validateChecksum(calculatedHash, expectedHash)

	if !isValid {
		slog.Error("check sum validation gone wrong")
		fmt.Fprintln(os.Stderr, "Oops! Something went wrong while checking the downloaded ffmpeg binary. Please try downloading it manually or let us know if you need help.")
		os.Exit(1)
	}

	return &ResolvedInstall{
		Executable: pathToFile,
		FromCache:  false,
		Downloaded: true,
	}, nil
}

func calculateFileCheckSum(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func validateChecksum(calculatedChecksum, expectedChecksum []byte) bool {
	return subtle.ConstantTimeCompare(calculatedChecksum, expectedChecksum) == 1
}

func extractBinaries(targetDir, archive string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	cmd := exec.Command(
		"tar",
		"-xJf", archive,
		"-C", targetDir,
		"--strip-components=2",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	binaries := []string{"ffmpeg", "ffprobe"}
	for _, bin := range binaries {
		path := filepath.Join(targetDir, bin)
		if err := os.Chmod(path, 0755); err != nil {
			return err
		}
	}

	fmt.Println("✅ ffmpeg and ffprobe extracted to ./ffmpeg/")
	return nil
}
