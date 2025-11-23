package config

type Config struct {
	DebugDir      string
	CacheDisabled bool
}

var AppConfig Config

func SetConfig(config *Config) {
	AppConfig = *config
}

func GetConfig() *Config {
	return &AppConfig
}
