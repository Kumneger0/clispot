package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

type YtDlpArgs struct {
	CookiesFromBrowser *string `json:"cookies-from-browser"`
	Cookies            *string `json:"cookies"`
}

type Config struct {
	DebugDir      *string    `json:"debug-dir"`
	CacheDisabled bool       `json:"disable-cache"`
	CacheDir      *string    `json:"cache-dir"`
	YtDlpArgs     *YtDlpArgs `json:"yt-dlp-args"`
	HeadlessMode  bool       `json:"headless-mode"`
	SkipOnNoMatch bool       `json:"skip-on-no-match"`
}

var userConfigDir = os.UserConfigDir
var userCacheDir = os.UserCacheDir
var userHomeDir = os.UserHomeDir

func GetConfigDir(goos string) string {
	configDir, err := userConfigDir()
	if err != nil {
		if goos == "windows" {
			return filepath.Join(os.Getenv("APPDATA"), "clispot")
		}
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			return filepath.Join(xdgConfig, "clispot")
		}
		if goos == "darwin" {
			return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "clispot")
		}
		return filepath.Join(os.Getenv("HOME"), ".config", "clispot")
	}
	return filepath.Join(configDir, "clispot")
}

func GetStateDir(goos string) string {
	if goos == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "clispot")
	}
	if goos == "darwin" {
		return filepath.Join(os.Getenv("HOME"), "Library", "State", "clispot")
	}
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, _ := userHomeDir()
		stateDir = filepath.Join(homeDir, ".local", "state")
	}
	return filepath.Join(stateDir, "clispot")
}

func GetCacheDir(goos string) string {
	cacheDir, err := userCacheDir()
	if err != nil {
		homeDir, _ := userHomeDir()
		if goos == "windows" {
			return filepath.Join(os.Getenv("LOCALAPPDATA"), "clispot")
		}
		if goos == "darwin" {
			return filepath.Join(os.Getenv("HOME"), "Library", "Caches", "clispot")
		}
		return filepath.Join(homeDir, ".cache", "clispot")
	}
	return filepath.Join(cacheDir, "clispot")
}

func GetDefaultConfig(goos string) *Config {
	defaultDebugDir := filepath.Join(GetStateDir(goos), "logs")
	defaultCacheDir := GetCacheDir(goos)
	return &Config{
		DebugDir:      &defaultDebugDir,
		CacheDisabled: true,
		CacheDir:      &defaultCacheDir,
		YtDlpArgs:     &YtDlpArgs{},
		HeadlessMode:  false,
		SkipOnNoMatch: true,
	}
}

func GetUserConfig(goos string) *Config {
	configPath := filepath.Join(GetConfigDir(goos), "config.json")
	fileStat, err := os.Stat(configPath)
	if err != nil {
		return GetDefaultConfig(goos)
	}
	if fileStat.IsDir() {
		slog.Error("User config is a directory", "path", configPath)
		return GetDefaultConfig(goos)
	}
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Failed to read user config", "err", err)
		return GetDefaultConfig(goos)
	}

	config := GetDefaultConfig(goos)
	err = json.Unmarshal(configFile, config)
	if err != nil {
		slog.Error("Failed to unmarshal user config", "err", err)
		return GetDefaultConfig(goos)
	}
	return config
}

var AppConfig Config

func SetConfig(config *Config) {
	AppConfig = *config
}

func GetConfig() *Config {
	return &AppConfig
}
