package main

import (
	"log"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/kumneger0/clispot/cmd"
	logSetup "github.com/kumneger0/clispot/internal/logger"
)

var version = ""

func main() {
	logger := logSetup.Init()
	defer logger.Close()

	slog.Info("starting the application")

	err := godotenv.Load()
	if err != nil {
		slog.Info(err.Error())
		log.Fatal("Error loading .env file")
	}

	err = cmd.Execute(version)
	if err != nil {
		slog.Error(err.Error())
	}
}
