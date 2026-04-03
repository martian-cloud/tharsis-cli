package terminal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/fatih/color"
	"github.com/kr/text"
	"golang.org/x/term"
	"google.golang.org/grpc/status"
)

const (
	errorBar             = "│ "
	errorBarWidth        = 2 // visual width of errorBar (│ is 1 char + 1 space)
	defaultTerminalWidth = 80
)

// GetTerminalWidth returns the current terminal width, falling back to a
// default when stdout is not a TTY (e.g. piped output, CI environments).
func GetTerminalWidth() int {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if w > 0 {
		return w
	}

	return defaultTerminalWidth
}

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
		body := prettifyJSON(buildErrorMessage(err))

		var se interface{ GRPCStatus() *status.Status }
		if errors.As(err, &se) {
			writeBar("Code: " + se.GRPCStatus().Code().String())
			writeBar("")
		}

		for _, line := range wrapPreservingNewlines(body, GetTerminalWidth()-errorBarWidth) {
			writeBar(line)
		}

		writeBar("")
	}

	out.WriteString("\n")
	return out.String()
}

// buildErrorMessage walks the error chain from innermost to outermost,
// extracting clean messages (using gRPC status message when available)
// and joining them in outer: inner order.
func buildErrorMessage(err error) string {
	var messages []string
	for err != nil {
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

// prettifyJSON finds the last embedded JSON object or array in a string
// and replaces it with indented, highlighted JSON for readability.
// Any text before and after the JSON is preserved.
func prettifyJSON(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != '{' && s[i] != '[' {
			continue
		}

		var raw json.RawMessage
		if err := json.Unmarshal([]byte(s[i:]), &raw); err != nil {
			continue
		}

		// Determine where the JSON ends in the original string.
		jsonLen := len(json.RawMessage(raw))
		after := strings.TrimSpace(s[i+jsonLen:])

		pretty, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			return s
		}

		formatted := string(pretty)
		if !color.NoColor {
			var buf bytes.Buffer
			if err := quick.Highlight(&buf, formatted, "json", "terminal256", "monokai"); err == nil {
				formatted = strings.TrimRight(buf.String(), "\n")
			}
		}

		result := s[:i] + "\n" + formatted
		if after != "" {
			result += "\n" + after
		}

		return result
	}

	return s
}

// wrapPreservingNewlines splits on newlines first to preserve intentional
// breaks, then wraps each paragraph to fit within the given width.
// text.Wrap alone collapses embedded newlines into spaces.
// Indented JSON lines are passed through unwrapped.
func wrapPreservingNewlines(s string, width int) []string {
	var lines []string
	for paragraph := range strings.SplitSeq(s, "\n") {
		// Don't wrap lines with ANSI codes (e.g. highlighted JSON).
		if strings.Contains(paragraph, "\x1b[") {
			lines = append(lines, paragraph)
			continue
		}

		for line := range strings.SplitSeq(text.Wrap(paragraph, width), "\n") {
			lines = append(lines, line)
		}
	}

	return lines
}
