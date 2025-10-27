package youtube

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/ebitengine/oto/v3"
)

type Player struct {
	OtoPlayer         *oto.Player
	Close             func() error
	ByteCounterReader *byteCounterReader
}

var otoContext *oto.Context
var once sync.Once

func getOtoContext() (*oto.Context, chan struct{}, error) {
	var readyChan chan struct{}
	var ctxErr error
	once.Do(func() {
		ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
			SampleRate:   44100,
			ChannelCount: 2,
			Format:       oto.FormatSignedInt16LE,
		})
		readyChan = ready
		ctxErr = err
		otoContext = ctx
	})

	return otoContext, readyChan, ctxErr
}

type byteCounterReader struct {
	r     io.Reader
	total int
}

func (b *byteCounterReader) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	b.total += n
	return n, err
}

func (b *byteCounterReader) CurrentSeconds() float64 {
	bytesPerSecond := 176400.0
	return float64(b.total) / bytesPerSecond
}

func SearchAndDownloadMusic(trackName, albumName string, artistNames []string, shouldWait bool) (*Player, error) {
	searchQuery := "ytsearch:" + trackName

	if len(artistNames) > 0 {
		searchQuery = searchQuery + " " + artistNames[0]
	}

	yt := exec.Command("yt-dlp",
		searchQuery,
		"--no-playlist",
		"-f", "bestaudio[ext=m4a]/bestaudio",
		"--downloader", "aria2c",
		"--downloader-args", "aria2c:-x 16 -s 16 -k 1M --file-allocation=none",
		"--no-part",
		"-o", "-",
	)

	ff := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ac", "2",
		"-ar", "44100",
		"pipe:1",
	)

	ytStderr, _ := os.Create("ytStdErr.debug.txt")
	ffStderr, _ := os.Create("ffStdErr.debug.txt")

	ytOut, _ := yt.StdoutPipe()

	ff.Stdin = ytOut
	ff.Stderr = ffStderr
	yt.Stderr = ytStderr

	pr, pw := io.Pipe()
	ff.Stdout = pw

	if err := yt.Start(); err != nil {
		return nil, err
	}
	if err := ff.Start(); err != nil {
		return nil, err
	}
	ctx, ready, err := getOtoContext()
	if err != nil {
		log.Fatal(err)
	}

	if shouldWait {
		<-ready
	}

	type Reader struct {
		Read func(p byte) (n int, err error)
	}

	counter := &byteCounterReader{r: pr}
	player := ctx.NewPlayer(counter)
	player.Play()

	return &Player{
		OtoPlayer:         player,
		ByteCounterReader: counter,
		Close: func() error {
			var err error

			if player != nil {
				player.Close()
			}

			if ff.Process != nil {
				_ = ff.Process.Kill()
			}

			if yt.Process != nil {
				_ = yt.Process.Kill()
			}

			if ytOut != nil {
				_ = ytOut.Close()
			}

			_ = pw.Close()
			_ = pr.Close()

			return err
		},
	}, nil
}

func CheckMusicSimilarity() {
	//TODO: implement music similarity
}
