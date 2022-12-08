package output

import (
	"github.com/caarlos0/log"
	"github.com/mitchellh/cli"
)

// UI is used for interfacing with the terminal
type UI struct {
	BasicUI cli.Ui
}

// Ask asks the user for input using the given query. The response is returned as the given string, or an error.
func (u *UI) Ask(query string) (string, error) {
	return u.BasicUI.Ask(query)
}

// AskSecret asks the user for input using the given query, but does not echo the keystrokes to the terminal.
func (u *UI) AskSecret(query string) (string, error) {
	return u.BasicUI.AskSecret(query)
}

// Error is used to output error messages
func (u *UI) Error(message string) {
	u.BasicUI.Error(message)
}

// Info is used to output info messages
func (u *UI) Info(message string) {
	log.Info(message)
}

// Output is used to output messages with no formatting applied
func (u *UI) Output(message string) {
	u.BasicUI.Output(message)
}

// Warn is used to output warn messages
func (u *UI) Warn(message string) {
	log.Warn(message)
}
