// Package output provides formatting utilities for CLI output.
package output

import (
	"bytes"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/fatih/color"
	"github.com/kr/text"
)

var (
	reHelpHeader = regexp.MustCompile(`^[a-zA-Z0-9_-].*:$`)             // matches "Section:" style headers
	reHelpFlag   = regexp.MustCompile(`(\s|^|")(-[\w-]+)(\s|$|")`)      // matches -flag style options
	reCodeBlock  = regexp.MustCompile("(?s)([ ]*)```(\\w*)\\n(.*?)```") // matches ```lang\ncode``` with optional indent
)

// CommandHelpInfo contains the components of command help text.
type CommandHelpInfo struct {
	ProductName string
	Usage       string
	Description string
	Flags       *flag.FlagSet
	Example     string
}

// CommandHelp builds and formats help text for a command with syntax highlighting.
// It assembles usage, description, flags, and example sections, then applies color
// highlighting to headers, flags, and command references.
func CommandHelp(info CommandHelpInfo) string {
	var buf bytes.Buffer
	bold := color.New(color.Bold)
	highlightColor := color.New(color.FgMagenta)

	// Build help sections
	buf.WriteString("\n" + strings.TrimSpace(info.Usage) + "\n")

	if desc := strings.TrimSpace(info.Description); desc != "" {
		buf.WriteString("\n" + highlightCode(desc, info.ProductName) + "\n")
	}

	if info.Flags != nil {
		buf.WriteString("\n" + bold.Sprint("Command options:") + "\n")
		var optionsBuf bytes.Buffer
		writer := tabwriter.NewWriter(&optionsBuf, 0, 80, 0, ' ', 0)
		info.Flags.VisitAll(func(f *flag.Flag) {
			var defValue string
			if f.DefValue != "" {
				defValue = fmt.Sprintf("(default %s)", f.DefValue)
			}
			wrapped := text.Wrap(f.Usage, 70)
			lines := strings.Split(wrapped, "\n")
			for i, line := range lines {
				lines[i] = strings.Repeat(" ", 6) + line
			}
			fmt.Fprintf(writer, "  %s %s\n\t%s\n", highlightColor.Sprintf("-%s", f.Name), defValue, strings.Join(lines, "\n"))
		})
		writer.Flush()
		buf.WriteString(optionsBuf.String())
	}

	if example := strings.TrimSuffix(strings.TrimPrefix(info.Example, "\n"), "\n"); example != "" {
		buf.WriteString("\n" + bold.Sprint("Example:") + "\n\n" + highlightCode(example, info.ProductName) + "\n\n")
	}

	// Apply syntax highlighting line by line
	v := strings.TrimSpace(buf.String())
	buf.Reset()

	seenHeader := false
	productPrefix := info.ProductName + " "
	for _, line := range strings.Split(v, "\n") {
		// Highlight product name at start of usage line
		if info.ProductName != "" && strings.HasPrefix(line, productPrefix) {
			buf.WriteString(highlightColor.Sprint(info.ProductName))
			buf.WriteString(line[len(info.ProductName):])
			buf.WriteString("\n")
			continue
		}

		// Highlight "Usage:" prefix
		if strings.HasPrefix(line, "Usage: ") {
			buf.WriteString(highlightColor.Sprint("Usage: "))
			buf.WriteString(line[7:])
			buf.WriteString("\n")
			continue
		}

		// Bold section headers like "Commands:" or "Options:"
		if reHelpHeader.MatchString(line) {
			seenHeader = true
			buf.WriteString(bold.Sprint(line))
			buf.WriteString("\n")
			continue
		}

		// Highlight quoted command references like "product run"
		if info.ProductName != "" {
			reHelpCmd := regexp.MustCompile(`"` + regexp.QuoteMeta(info.ProductName) + ` (\w\s?)+"`)
			if matches := reHelpCmd.FindAllStringIndex(line, -1); len(matches) > 0 {
				idx := 0
				for _, match := range matches {
					buf.WriteString(line[idx : match[0]+1])
					buf.WriteString(highlightColor.Sprint(line[match[0]+1 : match[1]-1]))
					idx = match[1] - 1
				}
				buf.WriteString(line[idx:])
				buf.WriteString("\n")
				continue
			}
		}

		// Highlight flags (only before first header to avoid coloring subcommand descriptions)
		if !seenHeader {
			if matches := reHelpFlag.FindAllStringSubmatchIndex(line, -1); len(matches) > 0 {
				idx := 0
				for _, match := range matches {
					start, end := match[4], match[5]
					buf.WriteString(line[idx:start])
					buf.WriteString(highlightColor.Sprint(line[start:end]))
					idx = end
				}
				buf.WriteString(line[idx:])
				buf.WriteString("\n")
				continue
			}
		}

		buf.WriteString(line)
		buf.WriteString("\n")
	}

	return strings.TrimSuffix(buf.String(), "\n")
}

// highlightCode applies syntax highlighting to code blocks and shell commands.
func highlightCode(s, productName string) string {
	// Check for explicit code blocks
	if reCodeBlock.MatchString(s) {
		return reCodeBlock.ReplaceAllStringFunc(s, func(match string) string {
			parts := reCodeBlock.FindStringSubmatch(match)
			if len(parts) < 4 || parts[2] == "" {
				return parts[3] // return code without markers if no language
			}
			return highlight(parts[3], parts[2])
		})
	}

	// Auto-detect shell commands in examples
	s = strings.TrimSpace(s)
	if productName != "" {
		if strings.HasPrefix(s, productName+" ") || strings.HasPrefix(s, "./"+productName+" ") {
			return highlight(s, "bash")
		}
	}

	return s
}

func highlight(code, lang string) string {
	if color.NoColor {
		return strings.TrimSpace(code)
	}

	var buf bytes.Buffer
	if err := quick.Highlight(&buf, strings.TrimSpace(code), lang, "terminal16m", "monokai"); err != nil {
		return code
	}
	return strings.TrimSpace(buf.String())
}
