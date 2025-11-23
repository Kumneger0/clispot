package youtube

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/ebitengine/oto/v3"
	"github.com/kumneger0/clispot/internal/config"
)

type Player struct {
	OtoPlayer         *oto.Player
	Close             func(isSkip bool) error
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
	if n > 0 {
		b.total += n
	}
	return n, err
}

func (b *byteCounterReader) CurrentSeconds() float64 {
	return float64(b.total) / 176400.0
}

func SearchAndDownloadMusic(
	trackName string,
	albumName string,
	artistNames []string,
	spotifyID string,
	shouldWait bool,
) (*Player, error) {
	searchQuery := "ytsearch:" + trackName
	if len(artistNames) > 0 {
		searchQuery += " " + artistNames[0]
	}

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache", "clispot", "yt-audio")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}

	musicPath := filepath.Join(cacheDir, spotifyID+".m4a")

	appConfig := config.GetConfig()

	logPathName := appConfig.DebugDir

	ytStderr, _ := os.Create(filepath.Join(logPathName, "ytstderr.log"))
	ffStderr, _ := os.Create(filepath.Join(logPathName, "ffstderr.log"))

	if _, err := os.Stat(musicPath); err == nil {
		return playExistingMusic(musicPath, shouldWait, ffStderr, ytStderr)
	}

	yt := exec.Command("yt-dlp",
		searchQuery,
		"--no-playlist",
		"-f", "bestaudio[ext=m4a]/bestaudio",
		"-o", "-",
	)
	yt.Stderr = ytStderr

	ytOut, err := yt.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := yt.Start(); err != nil {
		return nil, err
	}

	isCacheDisabled := appConfig.CacheDisabled

	var cacheFile *os.File

	var reader io.Reader
	if !isCacheDisabled {
		cacheFile, err = os.Create(musicPath)
		if err != nil {
			return nil, err
		}
		reader = io.TeeReader(ytOut, cacheFile)
	} else {
		reader = ytOut
	}

	ff := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ac", "2",
		"-ar", "44100",
		"pipe:1",
	)

	ff.Stdin = reader
	ff.Stderr = ffStderr

	pr, pw := io.Pipe()
	ff.Stdout = pw

	if err := ff.Start(); err != nil {
		return nil, err
	}

	ctx, ready, err := getOtoContext()
	if err != nil {
		return nil, err
	}
	if shouldWait {
		<-ready
	}

	counter := &byteCounterReader{
		r: pr,
	}

	player := ctx.NewPlayer(counter)
	player.Play()

	return &Player{
		OtoPlayer:         player,
		ByteCounterReader: counter,
		Close: func(isSkip bool) error {
			var firstErr error

			if player != nil {
				player.Close()
			}

			if ff.Process != nil {
				if err := ff.Process.Kill(); err != nil {
					slog.Error(err.Error())
					firstErr = err
				}
			}
			if yt.Process != nil {
				if err := yt.Process.Kill(); err != nil && firstErr == nil {
					slog.Error(err.Error())
					firstErr = err
				}
			}

			_ = pw.Close()
			_ = pr.Close()
			if ytStderr != nil {
				_ = ytStderr.Close()
			}
			if ffStderr != nil {
				_ = ffStderr.Close()
			}

			if cacheFile != nil {
				_ = cacheFile.Close()
			}

			if isSkip {
				if err := os.Remove(musicPath); err != nil && !os.IsNotExist(err) && firstErr == nil {
					slog.Error(err.Error())
					firstErr = err
				}
			}

			return firstErr
		},
	}, nil
}

func playExistingMusic(musicPath string, shouldWait bool, ffStderr, ytStderr *os.File) (*Player, error) {
	f, err := os.Open(musicPath)
	if err != nil {
		return nil, err
	}

	ff := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ac", "2",
		"-ar", "44100",
		"pipe:1",
	)

	ff.Stdin = f
	ff.Stderr = ffStderr

	pr, pw := io.Pipe()
	ff.Stdout = pw

	if err := ff.Start(); err != nil {
		return nil, err
	}

	ctx, ready, err := getOtoContext()
	if err != nil {
		return nil, err
	}
	if shouldWait {
		<-ready
	}

	counter := &byteCounterReader{
		r: pr,
	}

	player := ctx.NewPlayer(counter)
	player.Play()

	return &Player{
		OtoPlayer:         player,
		ByteCounterReader: counter,
		Close: func(isSkip bool) error {
			var firstErr error

			if player != nil {
				player.Close()
			}
			_ = f.Close()
			_ = pw.Close()
			_ = pr.Close()

			if ff.Process != nil {
				if err := ff.Process.Kill(); err != nil {
					slog.Error(err.Error())
					firstErr = err
				}
			}

			if ytStderr != nil {
				_ = ytStderr.Close()
			}
			if ffStderr != nil {
				_ = ffStderr.Close()
			}

			return firstErr
		},
	}, nil
}
