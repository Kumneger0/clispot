//go:build windows

package mpris

import (
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/ui"
)

func GetDbusInstance() (*ui.Instance, *chan types.DBusMessage, error) {
	// TODO: implement mpris for windows
	return nil, nil, nil
}
