package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
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
}

func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		if runtime.GOOS == "windows" {
			return filepath.Join(os.Getenv("APPDATA"), "clispot")
		}
		if runtime.GOOS == "darwin" {
			return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "clispot")
		}
		return filepath.Join(os.Getenv("HOME"), ".config", "clispot")
	}
	return filepath.Join(configDir, "clispot")
}

func GetStateDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "clispot")
	}
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, _ := os.UserHomeDir()
		stateDir = filepath.Join(homeDir, ".local", "state")
	}
	return filepath.Join(stateDir, "clispot")
}

func GetCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		if runtime.GOOS == "windows" {
			return filepath.Join(os.Getenv("LOCALAPPDATA"), "clispot")
		}
		if runtime.GOOS == "darwin" {
			return filepath.Join(os.Getenv("HOME"), "Library", "Caches", "clispot")
		}
		return filepath.Join(homeDir, ".cache", "clispot")
	}
	return filepath.Join(cacheDir, "clispot")
}

func GetDefaultConfig() *Config {
	defaultDebugDir := filepath.Join(GetStateDir(), "logs")
	defaultCacheDir := GetCacheDir()
	return &Config{
		DebugDir:      &defaultDebugDir,
		CacheDisabled: true,
		CacheDir:      &defaultCacheDir,
		YtDlpArgs:     &YtDlpArgs{},
		HeadlessMode:  false,
	}
}

func GetUserConfig() *Config {
	configPath := filepath.Join(GetConfigDir(), "config.json")
	fileStat, err := os.Stat(configPath)
	if err != nil {
		slog.Error("Failed to get user config", "err", err)
		return GetDefaultConfig()
	}
	if fileStat.IsDir() {
		slog.Error("User config is a directory", "path", configPath)
		return GetDefaultConfig()
	}
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Failed to read user config", "err", err)
		return GetDefaultConfig()
	}

	config := GetDefaultConfig()
	err = json.Unmarshal(configFile, config)
	if err != nil {
		slog.Error("Failed to unmarshal user config", "err", err)
		return GetDefaultConfig()
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
