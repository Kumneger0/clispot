package main

import (
	"log/slog"

	"github.com/kumneger0/clispot/cmd"
)

var (
	version             = ""
	spotifyClientID     = ""
	spotifyClientSecret = ""
)

func main() {
	err := cmd.Execute(version, spotifyClientID, spotifyClientSecret)
	if err != nil {
		slog.Error(err.Error())
	}
}
