package youtube

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ebitengine/oto/v3"
	"github.com/kumneger0/clispot/internal/config"
	"github.com/kumneger0/clispot/internal/notification"
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
	ytDlpErrWriter *io.PipeWriter,
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

	args := []string{
		searchQuery,
		"--no-playlist",
		"-f", "bestaudio",
		"-o", "-",
	}

	if appConfig.YtDlpArgs.CookiesFromBrowser != nil {
		args = append(args, "--cookies-from-browser", *appConfig.YtDlpArgs.CookiesFromBrowser)
	}
	if appConfig.YtDlpArgs.Cookies != nil {
		args = append(args, "--cookies", *appConfig.YtDlpArgs.Cookies)
	}

	logPathName := appConfig.DebugDir

	ytStderr, _ := os.Create(filepath.Join(*logPathName, "ytstderr.log"))
	ffStderr, _ := os.Create(filepath.Join(*logPathName, "ffstderr.log"))

	var ytdlpWriter io.Writer

	if ytDlpErrWriter != nil {
		ytdlpWriter = io.MultiWriter(ytStderr, ytDlpErrWriter)
	} else {
		ytdlpWriter = ytStderr
	}

	if _, err := os.Stat(musicPath); err == nil {
		player, isPlayable, err := playExistingMusic(musicPath, shouldWait, ffStderr, ytStderr)
		if err != nil {
			slog.Error(err.Error())
		}
		if isPlayable {
			return player, nil
		}
		slog.Error("cached audio is not playable trying to play from youtube")
	}

	yt := exec.Command("yt-dlp", args...)
	yt.Stderr = ytdlpWriter

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

func playExistingMusic(musicPath string, shouldWait bool, ffStderr, ytStderr *os.File) (*Player, bool, error) {
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		notificationTitle := "ffprobe is missing"
		notificationMessage := "ffprobe is missing, please install it helps us to check the status of cached audio"
		notification.Notify(notificationTitle, notificationMessage)
		return nil, false, err
	}

	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "json",
		musicPath,
	)

	err = cmd.Run()
	playable := err == nil

	if !playable {
		return nil, false, errors.New("audio is not playable")
	}

	f, err := os.Open(musicPath)
	if err != nil {
		slog.Error(err.Error())
		notificationTitle := "Error opening music file"
		notificationMessage := err.Error()
		notification.Notify(notificationTitle, notificationMessage)
		return nil, false, err
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
		slog.Error(err.Error())
		notificationTitle := "Audio Processing Failed"
		notificationMessage := err.Error()
		notification.Notify(notificationTitle, notificationMessage)
		return nil, false, err
	}

	ctx, ready, err := getOtoContext()
	if err != nil {
		slog.Error(err.Error())
		return nil, false, err
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
	}, true, nil
}

type YtDlpLogs string

const (
	WARNING  YtDlpLogs = "warning"
	INFO     YtDlpLogs = "info"
	ERROR    YtDlpLogs = "error"
	DOWNLOAD YtDlpLogs = "download"
	YOUTUBE  YtDlpLogs = "youtube"
)

type ScanFuncArgs struct {
	Line    string
	LogType YtDlpLogs
}

func ReadYtDlpErrReader(reader *io.PipeReader, scanFunc func(args ScanFuncArgs)) {
	if reader == nil {
		return
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), "error") {
			scanFunc(ScanFuncArgs{
				Line:    line,
				LogType: ERROR,
			})
		} else if strings.Contains(strings.ToLower(line), "warning") {
			scanFunc(ScanFuncArgs{
				Line:    line,
				LogType: WARNING,
			})
		} else if strings.Contains(strings.ToLower(line), "download") {
			scanFunc(ScanFuncArgs{
				Line:    line,
				LogType: DOWNLOAD,
			})
		} else if strings.Contains(strings.ToLower(line), "youtube") {
			scanFunc(ScanFuncArgs{
				Line:    line,
				LogType: YOUTUBE,
			})
		} else {
			scanFunc(ScanFuncArgs{
				Line:    line,
				LogType: INFO,
			})
		}
	}
}
