package config

type Browser string

const (
	CHROME   Browser = "chrome"
	CHROMIUM Browser = "chromium"
	FIREFOX  Browser = "firefox"
	BRAVE    Browser = "brave"
	EDGE     Browser = "edge"
	OPERA    Browser = "opera"
	SAFARI   Browser = "safari"
	VIVALDI  Browser = "vivaldi"
	WHALE    Browser = "whale"
)

type YtDlpArgs struct {
	CookiesFromBrowser *Browser
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
