package youtube

import (
	"io"
	"os"
	"os/exec"
)

func SearchAndDownloadMusic(trackName, albumName string, artistNames []string) (*os.Process, error) {
	searchQuery := "ytsearch:" + trackName

	if len(artistNames) > 0 {
		searchQuery = searchQuery + artistNames[0]
	}

	yt := exec.Command("yt-dlp",
		searchQuery,
		"-f", "bestaudio[ext=m4a]/bestaudio",
		"--downloader", "aria2c",
		"--downloader-args", "aria2c:-x 16 -s 16 -k 1M --file-allocation=none",
		"--no-part",
		"-o", "-",
	)

	ff := exec.Command("ffplay", "-nodisp", "-autoexit", "-i", "-")

	ytStderr, _ := os.Create("ytStdErr.debug.txt")
	ffStderr, _ := os.Create("ffStdErr.debug.txt")

	r, w := io.Pipe()
	yt.Stdout = w
	ff.Stdin = r
	yt.Stderr = ytStderr
	ff.Stderr = ffStderr

	if err := ff.Start(); err != nil {
		return nil, err
	}

	go (func() {
		_ = yt.Run()
		w.Close()
	})()

	return ff.Process, nil
}

func CheckMusicSimilarity() {
	//TODO: implement music similarity
}
