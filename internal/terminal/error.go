package terminal

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kr/text"
	"golang.org/x/term"
	"google.golang.org/grpc/status"
)

const (
	errorBar             = "│ "
	defaultTerminalWidth = 80
)

// formatError builds a user-facing error display with a colored bar prefix on
// every line. The error body may contain newlines (e.g. multi-field validation
// errors from gRPC), so it is split before rendering to ensure the bar is
// never broken.
func formatError(err error, format string, a ...any) string {
	var out strings.Builder
	bar := colorError.Sprint(errorBar)

	writeBar := func(content string) { out.WriteString(bar + content + "\n") }

	out.WriteString("\n")
	writeBar("")

	summary := strings.ReplaceAll(fmt.Sprintf(format, a...), "\n", " ")
	writeBar(colorError.Sprint("Error:") + " " + summary)

	writeBar("")

	if err != nil {
		body := buildErrorMessage(err)

		var se interface{ GRPCStatus() *status.Status }
		if errors.As(err, &se) {
			writeBar("Code: " + se.GRPCStatus().Code().String())
			writeBar("")
		}

		for _, line := range wrapPreservingNewlines(body, getTerminalWidth()-len(errorBar)) {
			writeBar(line)
		}

		writeBar("")
	}

	out.WriteString("\n")
	return out.String()
}

// wrapPreservingNewlines splits on newlines first to preserve intentional
// breaks, then wraps each paragraph to fit within the given width.
// text.Wrap alone collapses embedded newlines into spaces.
func wrapPreservingNewlines(s string, width int) []string {
	var lines []string
	for paragraph := range strings.SplitSeq(s, "\n") {
		for line := range strings.SplitSeq(text.Wrap(paragraph, width), "\n") {
			lines = append(lines, line)
		}
	}

	return lines
}

// buildErrorMessage walks the error chain from innermost to outermost,
// extracting clean messages (using gRPC status message when available)
// and joining them in outer: inner order.
func buildErrorMessage(err error) string {
	var messages []string
	for err != nil {
		// Use a direct type assertion instead of errors.As so we only match
		// the current error in the chain, not a wrapped inner gRPC error.
		// This lets us extract the clean message at each level separately.
		if se, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
			messages = append(messages, se.GRPCStatus().Message())
			break
		}

		inner := errors.Unwrap(err)
		if inner == nil {
			messages = append(messages, err.Error())
		} else {
			messages = append(messages, outerError(err))
		}

		err = inner
	}

	return strings.Join(messages, ": ")
}

// outerError returns the error's own message without the wrapped inner text.
func outerError(err error) string {
	if err == nil {
		return ""
	}

	if prefix, _, ok := strings.Cut(err.Error(), ": "); ok {
		return prefix
	}

	return err.Error()
}

// getTerminalWidth returns the current terminal width, falling back to a
// default when stdout is not a TTY (e.g. piped output, CI environments).
func getTerminalWidth() int {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if w > 0 {
		return w
	}

	return defaultTerminalWidth
}
