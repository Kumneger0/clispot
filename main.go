package main

import (
	"log/slog"

	"github.com/kumneger0/clispot/cmd"
)

var (
	version = ""
)

func main() {
	err := cmd.Execute(version)
	if err != nil {
		slog.Error(err.Error())
	}
}
