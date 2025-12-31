package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gofrs/flock"
	"github.com/kumneger0/clispot/internal/config"
	"github.com/kumneger0/clispot/internal/headless"
	logSetup "github.com/kumneger0/clispot/internal/logger"
	"github.com/kumneger0/clispot/internal/youtube"
	"go.dalton.dog/bubbleup"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/kumneger0/clispot/internal/mpris"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/ui"
)

var (
	Program *tea.Program
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clispot",
		Short: "spotify music player",
		RunE: func(cmd *cobra.Command, args []string) error {
			lockFilePath := filepath.Join(os.TempDir(), "clispot.lock")

			fileLock := flock.New(lockFilePath)
			locked, err := fileLock.TryLock()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error trying to acquire lock: %v\n", err)
				os.Exit(1)
			}

			if !locked {
				showAnotherProcessIsRunning(lockFilePath)
				os.Exit(1)
			}
			defer func() {
				if err := fileLock.Unlock(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not unlock file: %v\n", err)
				}
				_ = os.Remove(lockFilePath)
			}()

			if runtime.GOOS != "windows" {
				pid := os.Getpid()
				if err := os.WriteFile(lockFilePath, []byte(strconv.Itoa(pid)), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not write PID to lock file: %v\n", err)
				}
			}

			return runRoot(cmd)
		},
	}

	cmd.AddCommand(newVersionCmd(version))
	cmd.AddCommand(clispotLog())
	cmd.AddCommand(ManCmd(cmd))
	return cmd
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func showAnotherProcessIsRunning(lockFilePath string) {
	if runtime.GOOS == "windows" {
		// Windows doesn't allow us to read the content of the file if the file is acquired by another process
		fmt.Fprintf(os.Stderr, "Another instance of clispot is already running.\n")
		return
	}
	pidBytes, readErr := os.ReadFile(lockFilePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading lock file: %v\n", readErr)
		os.Exit(1)
	}
	pid, parseErr := strconv.Atoi(string(pidBytes))

	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error parsing PID from lock file: %v\n", parseErr)
		os.Exit(1)
	}

	if !isProcessRunning(pid) {
		fmt.Fprintf(os.Stderr, "Another instance of clispot is not running (stale lock file for PID %d).\n", pid)
		fmt.Fprintf(os.Stderr, "Please try removing %s and running again if this persists.\n", lockFilePath)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Another instance of clispot is already running (PID: %d).\n", pid)
}

func runRoot(cmd *cobra.Command) error {
	debugDir, err := cmd.Flags().GetString("debug-dir")
	configFromFile := config.GetUserConfig(runtime.GOOS)

	if !cmd.Flags().Changed("debug-dir") && configFromFile.DebugDir != nil {
		debugDir = *configFromFile.DebugDir
	}

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cacheDir, err := cmd.Flags().GetString("cache-dir")
	if !cmd.Flags().Changed("cache-dir") && configFromFile.CacheDir != nil {
		cacheDir = *configFromFile.CacheDir
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	isCacheDisabled, err := cmd.Flags().GetBool("disable-cache")
	if !cmd.Flags().Changed("disable-cache") {
		isCacheDisabled = configFromFile.CacheDisabled
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(debugDir, 0755); err != nil {
		fmt.Printf("failed to create debug directory '%s': %v\n", debugDir, err)
		os.Exit(1)
	}

	fileInfo, err := os.Stat(debugDir)
	if err != nil {
		fmt.Printf("failed to stat debug directory '%s': %v\n", debugDir, err)
		os.Exit(1)
	}

	if !fileInfo.IsDir() {
		fmt.Printf("the debug path '%v' is not a directory\n", debugDir)
		os.Exit(1)
	}

	ytDlpArgs := config.YtDlpArgs{
		CookiesFromBrowser: nil,
		Cookies:            nil,
	}

	cookiesFromBrowser, err := cmd.Flags().GetString("cookies-from-browser")

	if !cmd.Flags().Changed("cookies-from-browser") {
		if configFromFile.YtDlpArgs != nil && configFromFile.YtDlpArgs.CookiesFromBrowser != nil {
			cookiesFromBrowser = *configFromFile.YtDlpArgs.CookiesFromBrowser
		}
	}

	if err != nil {
		slog.Error(err.Error())
	}

	if cookiesFromBrowser != "" {
		ytDlpArgs.CookiesFromBrowser = (&cookiesFromBrowser)
	}

	cookiesFile, err := cmd.Flags().GetString("cookies")

	if !cmd.Flags().Changed("cookies") {
		if configFromFile.YtDlpArgs != nil && configFromFile.YtDlpArgs.Cookies != nil {
			cookiesFile = *configFromFile.YtDlpArgs.Cookies
		}
	}

	if err != nil {
		slog.Error(err.Error())
	}

	if cookiesFile != "" {
		ytDlpArgs.Cookies = &cookiesFile
	}

	config.SetConfig(&config.Config{
		DebugDir:      &debugDir,
		CacheDisabled: isCacheDisabled,
		CacheDir:      &cacheDir,
		YtDlpArgs:     &ytDlpArgs,
	})

	logger := logSetup.Init(debugDir)
	defer logger.Close()

	slog.Info("starting the application")

	err = doAllDepsInstalled()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return nil
	}
	token, err := spotify.ReadUserCredentials()

	if err != nil {
		slog.Error(err.Error())
		token, _ = spotify.Authenticate()
	}

	if token == nil {
		slog.Error("failed to get user token")
		fmt.Println("we have failed to get your access token please open up an issue on our github page")
		os.Exit(1)
	}
	if token.ExpiresAt < time.Now().Unix() && token.RefreshToken != "" {
		token, err = spotify.RefreshToken(token.RefreshToken)
		if err != nil {
			slog.Error(err.Error())
			clispotConfigDir := config.GetConfigDir(runtime.GOOS)
			fmt.Printf("we have failed to refresh ur token could you try deleting clispot dir by using rm -rf %v  ", clispotConfigDir)
			os.Exit(1)
		}
	}

	isHeadlessMode, err := cmd.Flags().GetBool("headless")

	if err != nil {
		slog.Error(err.Error())
	}

	ins, messageChan, err := mpris.GetDbusInstance()

	if err != nil {
		slog.Error(err.Error())
	}

	model := ui.Model{
		GetUserToken: func() *types.UserTokenInfo {
			token, err := validateToken(token)
			if err != nil {
				slog.Error(err.Error())
				return nil
			}
			return token
		},
		FocusedOn:     ui.SideView,
		DBusConn:      ins,
		MainViewMode:  ui.NormalMode,
		SpotifyClient: spotify.NewAPIClient(spotify.NewAPIURL(), nil),
	}

	reader, writer := io.Pipe()
	model.YtDlpErrWriter = writer
	model.YtDlpErrReader = reader

	if isHeadlessMode {
		safeModel := ui.SafeModel{
			Model: &model,
		}
		headless.StartServer(&safeModel, messageChan)
		return nil
	}

	userSavedTracksListItem := spotify.UserSavedTracksListItem{
		Name: "Liked songs",
	}

	userPlayList, err := model.SpotifyClient.GetUserPlaylists(token.AccessToken)
	if err != nil {
		slog.Error(err.Error())
		fmt.Fprintln(os.Stdout, err)
	}

	if userPlayList == nil {
		slog.Error("GetUserPlaylists returned nil")
	}

	var items []list.Item
	if userPlayList != nil {
		for _, item := range userPlayList.Items {
			items = append(items, item)
		}
	}

	playlists := list.New(append([]list.Item{userSavedTracksListItem}, items...), ui.CustomDelegate{Model: &model}, 10, 20)
	playlistItems := list.New([]list.Item{}, ui.CustomDelegate{Model: &model}, 10, 20)

	input := textinput.New()
	input.Placeholder = "Search tracks, artists, albums..."
	input.Prompt = "> "
	input.CharLimit = 256

	model.Alert = *bubbleup.NewAlertModel(80, true, 10)

	model.Search = input
	musicQueueList := list.New([]list.Item{}, ui.CustomDelegate{Model: &model}, 10, 20)

	model.Playlist = playlists
	model.SelectedPlayListItems = playlistItems
	model.MusicQueueList = musicQueueList
	_, err = exec.LookPath("clispot-lyrics")
	if err != nil {
		model.IsLyricsServerInstalled = false
	} else {
		model.IsLyricsServerInstalled = true
	}
	Program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	go func() {
		youtube.ReadYtDlpErrReader(model.YtDlpErrReader, func(args youtube.ScanFuncArgs) {
			Program.Send(args)
		})
	}()

	go func() {
		if messageChan == nil {
			return
		}
		for v := range *messageChan {
			Program.Send(v)
		}
	}()

	_, err = Program.Run()
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}

	if ins != nil {
		ins.Conn.Close()
	}
	return nil
}

func validateToken(token *types.UserTokenInfo) (*types.UserTokenInfo, error) {
	if token.ExpiresAt > time.Now().Unix() {
		return token, nil
	}
	if token.RefreshToken != "" {
		token, err := spotify.RefreshToken(token.RefreshToken)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		return token, nil
	}
	//this means something went wrong re-authenticate
	token, err := spotify.Authenticate()
	if err != nil {
		return nil, err
	}
	return token, nil
}

func doAllDepsInstalled() error {
	toolNames := []string{"yt-dlp", "ffmpeg"}
	var error error
	for _, toolName := range toolNames {
		_, err := exec.LookPath(toolName)
		if err != nil {
			error = fmt.Errorf("failed to find %v in the path have u installed it", toolName)
			break
		}
	}
	return error
}

func Execute(version string) error {
	cmd := newRootCmd(version)
	defaultDebugDir := filepath.Join(config.GetStateDir(runtime.GOOS), "logs")
	cmd.Flags().StringP("debug-dir", "d", defaultDebugDir, "a path to store app logs")
	cmd.Flags().StringP("cache-dir", "c", config.GetCacheDir(runtime.GOOS), "a path to store app cache")
	cmd.Flags().Bool("disable-cache", false, "disable cache")
	cmd.Flags().Bool("headless", false, "Headless mode which provides api endpoint to build custom ui")
	cmd.Flags().String("cookies-from-browser", "", "The name of the browser to load cookies from this option is used by yt-dlp see yt-dlp docs to see supported browsers")
	cmd.Flags().String("cookies", "", "cookies file the option you pass for this flag will be passed to yt-dlp checkout yt-dlp docs to learn more about this flag")

	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}
	return nil
}
