package install

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func download(url string, outputPath string) error {
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

	bar := pb.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)

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
