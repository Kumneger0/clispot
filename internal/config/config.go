package config

type Browser string

const (
	CHROME  Browser = "chrome"
	FIREFOX Browser = "firefox"
	BRAVE   Browser = "brave"
)

type YtDlpArgs struct {
	CookiesFromBrowser Browser
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
