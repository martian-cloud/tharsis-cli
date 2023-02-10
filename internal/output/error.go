// Package output is responsible for formatting
// the final output / errors the user sees in the
// CLI. It automatically parses any GraphQL errors
// from a given error and displays the 'detail'
// field if present.
package output

import (
	"strings"
)

const (
	bar           = "| "
	reset         = "\033[0m"
	redBar        = "\033[31m" + bar
	newline       = "\n"
	redBarNewline = redBar + newline
)

// FormatError outputs the error with proper formatting for CLI.
func FormatError(summary string, err error) string {
	msg := formatError(summary, "")

	if err != nil {
		return msg + formatError("", strings.TrimSpace(err.Error()))
	}

	return msg + newline
}

func formatError(summary, err string) string {
	if summary != "" {
		return newline +
			redBarNewline +
			redBar +
			"Error: " +
			reset +
			summary +
			newline +
			redBarNewline +
			reset
	}

	return redBar +
		reset +
		strings.ReplaceAll(err, newline, newline+redBar+reset) +
		newline +
		redBarNewline +
		reset
}
