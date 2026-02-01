package youtube

import (
	"bufio"
	"errors"
	"fmt"

	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/ebitengine/oto/v3"
	"github.com/kumneger0/clispot/internal/command"
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

type CoreDepsPath struct {
	FFmpeg  string
	YtDlp   string
	FFprobe string
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
	durationSec int,
	coreDepsPath *CoreDepsPath,
) (*Player, error) {
	if coreDepsPath == nil {
		return nil, errors.New("Filed to fund necessary dependencies")
	}
	searchQuery := "ytsearch5:" + trackName
	if len(artistNames) > 0 {
		searchQuery += " " + artistNames[0]
	}

	appConfig := config.GetConfig()
	cacheDir := filepath.Join(*appConfig.CacheDir, "yt-audio")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}

	junkWords := []string{"hour", "loop", "slowed", "reverb", "mix", "playlist", "album"}
	JunkWordsToFilterOut := make([]string, 0)
	for _, word := range junkWords {
		if !strings.Contains(strings.ToLower(trackName), word) {
			JunkWordsToFilterOut = append(JunkWordsToFilterOut, word)
		}
	}

	var rejectTitleConditions []string

	for _, word := range JunkWordsToFilterOut {
		rejectTitleConditions = append(rejectTitleConditions, "--reject-title", fmt.Sprintf("(?i)%s", word))
	}

	var conditions []string
	conditions = append(conditions, "!is_live")
	conditions = append(conditions, fmt.Sprintf("duration >= %d", durationSec-60))
	conditions = append(conditions, fmt.Sprintf("duration <= %d", durationSec+60))
	conditions = append(conditions, `title ~= "(?i)`+strings.ReplaceAll(regexp.QuoteMeta(trackName), " ", ".*")+`"`)
	matchFilter := strings.Join(conditions, " & ")

	musicPath := filepath.Join(cacheDir, spotifyID+".m4a")

	args := []string{
		searchQuery,
		"--no-playlist",
		"-f", "bestaudio",
		"--match-filter", strings.TrimSpace(matchFilter),
	}

	args = append(slices.Concat(args, rejectTitleConditions), "--max-downloads", "1", "-o", "-")

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
		player, isPlayable, err := playExistingMusic(musicPath, shouldWait, ffStderr, ytStderr, *coreDepsPath)
		if err != nil {
			slog.Error(err.Error())
		}
		if isPlayable {
			return player, nil
		}
		slog.Error("cached audio is not playable trying to play from youtube")
	}

	yt, _ := command.ExecCommand(coreDepsPath.YtDlp, args...)

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

	ff, _ := command.ExecCommand(coreDepsPath.FFmpeg,
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
	go func() {
		player.Play()
	}()

	return &Player{
		OtoPlayer:         player,
		ByteCounterReader: counter,
		Close: func(isSkip bool) error {
			var firstErr error
			if player != nil {
				player.Close()
			}
			if ff.Process != nil {
				err := command.KillProcess(ff.Process)
				if err != nil {
					slog.Error(err.Error())
					firstErr = err
				}
			}

			if yt.Process != nil {
				err := command.KillProcess(yt.Process)
				if err != nil && firstErr == nil {
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

func playExistingMusic(musicPath string, shouldWait bool, ffStderr, ytStderr *os.File, coreCoreDepsPath CoreDepsPath) (*Player, bool, error) {
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		notificationTitle := "ffprobe is missing"
		notificationMessage := "ffprobe is missing, please install it to enable cached audio validation"
		notification.Notify(notificationTitle, notificationMessage)
		return nil, false, err
	}

	cmd, _ := command.ExecCommand(
		coreCoreDepsPath.FFprobe,
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

	ff, _ := command.ExecCommand(coreCoreDepsPath.FFmpeg,
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
		if closeErr := f.Close(); closeErr != nil {
			slog.Error(closeErr.Error())
		}
		notification.Notify(notificationTitle, notificationMessage)
		return nil, false, err
	}

	ctx, ready, err := getOtoContext()
	if err != nil {
		slog.Error(err.Error())
		if closeErr := f.Close(); closeErr != nil {
			slog.Error(closeErr.Error())
		}
		if closeErr := pw.Close(); closeErr != nil {
			slog.Error(closeErr.Error())
		}
		if closeErr := pr.Close(); closeErr != nil {
			slog.Error(closeErr.Error())
		}
		if ff.Process != nil {
			err := command.KillProcess(ff.Process)
			if err != nil {
				slog.Error(err.Error())
			}
		}
		return nil, false, err
	}
	if shouldWait {
		<-ready
	}

	counter := &byteCounterReader{
		r: pr,
	}

	player := ctx.NewPlayer(counter)

	go func() {
		player.Play()
	}()

	return &Player{
		OtoPlayer:         player,
		ByteCounterReader: counter,
		Close: func(isSkip bool) error {
			var firstErr error

			if player != nil {
				err := player.Close()
				if err != nil {
					slog.Error(err.Error())
				}
			}
			err = f.Close()
			if err != nil {
				slog.Error(err.Error())
			}
			err = pw.Close()
			if err != nil {
				slog.Error(err.Error())
			}
			err = pr.Close()
			if err != nil {
				slog.Error(err.Error())
			}

			if ff.Process != nil {
				if err := command.KillProcess(ff.Process); err != nil {
					slog.Error(err.Error())
					firstErr = err
				}
			}

			if ytStderr != nil {
				err = ytStderr.Close()
				if err != nil {
					slog.Error(err.Error())
				}
			}
			if ffStderr != nil {
				err = ffStderr.Close()
				if err != nil {
					slog.Error(err.Error())
				}
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
				Line:    afterKeyword(line, "error"),
				LogType: ERROR,
			})
		} else if strings.Contains(strings.ToLower(line), "warning") {
			scanFunc(ScanFuncArgs{
				Line:    afterKeyword(line, "warning"),
				LogType: WARNING,
			})
		} else if strings.Contains(strings.ToLower(line), "download") {
			scanFunc(ScanFuncArgs{
				Line:    afterKeyword(line, "download"),
				LogType: DOWNLOAD,
			})
		} else if strings.Contains(strings.ToLower(line), "youtube") {
			scanFunc(ScanFuncArgs{
				Line:    afterKeyword(line, "youtube"),
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

func afterKeyword(line, keyword string) string {
	lower := strings.ToLower(line)
	keyword = strings.ToLower(keyword)

	idx := strings.Index(lower, keyword)
	if idx == -1 {
		return line
	}

	return strings.TrimSpace(line[idx+len(keyword):])
}
