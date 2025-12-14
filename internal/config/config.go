package config

import (
	"encoding/json"
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
	YtDlpArgs     *YtDlpArgs `json:"yt-dlp-args"`
	HeadlessMode  bool       `json:"headless-mode"`
}

func GetConfigDir() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "clispot")
}

func GetStateDir() string {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, _ := os.UserHomeDir()
		stateDir = filepath.Join(homeDir, ".local", "state")
	}
	return filepath.Join(stateDir, "clispot")
}

func GetDefaultConfig() *Config {
	defaultDebugDir := filepath.Join(GetStateDir(), "logs")
	return &Config{
		DebugDir:      &defaultDebugDir,
		CacheDisabled: true,
		YtDlpArgs:     &YtDlpArgs{},
		HeadlessMode:  false,
	}
}

func GetUserConfig() *Config {
	configPath := filepath.Join(GetConfigDir(), "config.json")
	fileStat, err := os.Stat(configPath)
	if err != nil {
		return GetDefaultConfig()
	}
	if fileStat.IsDir() {
		return GetDefaultConfig()
	}
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return GetDefaultConfig()
	}
	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return GetDefaultConfig()
	}
	return &config
}

var AppConfig Config

func SetConfig(config *Config) {
	AppConfig = *config
}

func GetConfig() *Config {
	return &AppConfig
}
