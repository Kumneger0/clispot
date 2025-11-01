package main

import (
	"log/slog"

	"github.com/kumneger0/clispot/cmd"
	logSetup "github.com/kumneger0/clispot/internal/logger"
)

var (
	version             = ""
	spotifyClientID     = ""
	spotifyClientSecret = ""
)

func main() {
	logger := logSetup.Init()
	defer logger.Close()

	slog.Info("starting the application")
	err := cmd.Execute(version, spotifyClientID, spotifyClientSecret)
	if err != nil {
		slog.Error(err.Error())
	}
}
