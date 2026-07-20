package main

import (
	"fmt"
	"log/slog"

	"github.com/kumneger0/clispot/cmd"
)

var (
	version = ""
	Debug   = "false"
)

func main() {
	fmt.Println("debug", Debug)
	err := cmd.Execute(version, Debug == "true")
	if err != nil {
		slog.Error(err.Error())
	}
}
