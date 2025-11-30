package config

type YtDlpArgs struct {
	CookiesFromBrowser *string
	Cookies            *string
}

type Config struct {
	DebugDir      string
	CacheDisabled bool
	YtDlpArgs     *YtDlpArgs
}

var AppConfig Config

func SetConfig(config *Config) {
	AppConfig = *config
}

func GetConfig() *Config {
	return &AppConfig
}
