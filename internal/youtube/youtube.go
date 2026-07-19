package youtube

import (
	"bufio"
	"errors"
	"net/http"
	"time"

	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ebitengine/oto/v3"
	"github.com/kumneger0/clispot/internal/command"
	"github.com/kumneger0/clispot/internal/config"
	"github.com/kumneger0/clispot/internal/types"
)

var otoContext *oto.Context
var once sync.Once

const SearchResultCount = 5

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
	FFmpeg string
}

func SearchAndDownloadMusic(
	videoID string,
	shouldWait bool,
	coreDepsPath *CoreDepsPath,
	getStreamURL func() (string, error),
) tea.Cmd {
	return func() tea.Msg {
		if coreDepsPath == nil {
			return types.SearchAndDownloadMusicMsg{
				Player:  nil,
				VideoID: videoID,
				Err:     errors.New("failed to find necessary dependencies"),
			}
		}
		streamURL, err := getStreamURL()
		if err != nil {
			slog.Error(err.Error())
			return types.SearchAndDownloadMusicMsg{Player: nil, VideoID: videoID, Err: err}
		}
		appConfig := config.GetConfig()
		logPathName := appConfig.DebugDir
		ffStderr, _ := os.Create(filepath.Join(*logPathName, "ffstderr.log"))

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(streamURL)
		if err != nil {
			slog.Error(err.Error())
			return types.SearchAndDownloadMusicMsg{Player: nil, VideoID: videoID, Err: err}
		}

		ff, _ := command.ExecCommand(coreDepsPath.FFmpeg,
			"-i", "pipe:0",
			"-f", "s16le",
			"-acodec", "pcm_s16le",
			"-ac", "2",
			"-ar", "44100",
			"pipe:1",
		)

		ff.Stdin = resp.Body
		ff.Stderr = ffStderr

		pr, pw := io.Pipe()
		ff.Stdout = pw

		if err := ff.Start(); err != nil {
			return types.SearchAndDownloadMusicMsg{
				Player:  nil,
				VideoID: videoID,
				Err:     err,
			}
		}

		otoCtx, ready, err := getOtoContext()
		if err != nil {
			return types.SearchAndDownloadMusicMsg{
				Player:  nil,
				VideoID: videoID,
				Err:     err,
			}
		}
		if shouldWait {
			<-ready
		}

		counter := &types.ByteCounterReader{
			R: pr,
		}

		player := otoCtx.NewPlayer(counter)
		go func() {
			player.Play()
		}()

		return types.SearchAndDownloadMusicMsg{Player: &types.Player{
			OtoPlayer:         player,
			ByteCounterReader: counter,
			Close: func(shouldRemoveTheCacheFile bool) error {
				var firstErr error
				if player != nil {
					player.Close()
				}
				if err := resp.Body.Close(); err != nil {
					slog.Error(err.Error())
				}
				if ff.Process != nil {
					err := command.KillProcess(ff.Process)
					if err != nil {
						slog.Error(err.Error())
						firstErr = err
					}
				}

				_ = pw.Close()
				_ = pr.Close()

				if ffStderr != nil {
					_ = ffStderr.Close()
				}

				return firstErr
			},
		},
			VideoID: videoID,
			Err:     nil}
	}
}

type YtDlpLogs string

const (
	WARNING  YtDlpLogs = "warning"
	INFO     YtDlpLogs = "info"
	ERROR    YtDlpLogs = "error"
	DOWNLOAD YtDlpLogs = "download"
	YOUTUBE  YtDlpLogs = "youtube"
	SKIPPING YtDlpLogs = "skipping"
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
		if strings.Contains(strings.ToLower(line), "skipping") {
			scanFunc(ScanFuncArgs{
				Line:    "Skipping " + afterKeyword(line, "skipping"),
				LogType: SKIPPING,
			})
		}
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
	if err := scanner.Err(); err != nil {
		slog.Error(err.Error())
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
