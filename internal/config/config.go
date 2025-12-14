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

func GetDefaultConfig() *Config {
	userHomeDir, _ := os.UserHomeDir()
	defaultDebugDir := filepath.Join(userHomeDir, ".clispot", "logs")
	return &Config{
		DebugDir:      &defaultDebugDir,
		CacheDisabled: true,
		YtDlpArgs:     &YtDlpArgs{},
		HeadlessMode:  false,
	}
}

func GetUserConfig() *Config {
	userHomeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(userHomeDir, ".config", "clispot", "config.json")
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
