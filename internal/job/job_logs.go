// Package job contains the logic to display
// job logs from the log channel.
package job

import (
	"strings"

	"github.com/mitchellh/cli"
)

// DisplayLogs displays logs from a channel.
func DisplayLogs(logChannel chan string, ui cli.Ui) error {
	// Fetch and display logs until the job finishes.
	for {

		// Block on channel read until it has a clump of logs.
		newLogs, ok := <-logChannel
		if !ok {
			// Channel has been closed, no new logs.
			return nil
		}

		// Print something _only_ if there is something to print.
		// Don't output spurious newlines.
		if len(newLogs) > 0 {
			// The UI's Output function appends a newline, so we must remove the
			// trailing newline if one is present.
			newLogs = strings.TrimSuffix(newLogs, "\n")
			ui.Output(newLogs)
		}
	}
}

// The End.
