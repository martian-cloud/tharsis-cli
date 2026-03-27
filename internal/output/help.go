// Package output provides formatting utilities for CLI output.
package output

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/fatih/color"
	"github.com/kr/text"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

var (
	reHelpHeader = regexp.MustCompile(`^[a-zA-Z0-9_-].*:$`)             // matches "Section:" style headers
	reHelpFlag   = regexp.MustCompile(`(\s|^|")(-[\w-]+)(\W|$)`)        // matches -flag style options
	reCodeBlock  = regexp.MustCompile("(?s)([ ]*)```(\\w*)\\n(.*?)```") // matches ```lang\ncode``` with optional indent
)

// PrimaryColor returns a new Color instance with the primary brand color.
func PrimaryColor() *color.Color {
	return color.New(color.FgHiGreen)
}

// CommandHelpInfo contains the components of command help text.
type CommandHelpInfo struct {
	ProductName string
	Usage       string
	Description string
	Flags       *flag.Set
	Example     string
}

// helpBuilder assembles and colorizes command help text.
type helpBuilder struct {
	info      CommandHelpInfo
	bold      *color.Color
	highlight *color.Color
	warn      *color.Color
	danger    *color.Color
	reCmdRef  *regexp.Regexp
}

// CommandHelp builds and formats help text for a command with syntax highlighting.
func CommandHelp(info CommandHelpInfo) string {
	h := &helpBuilder{
		info:      info,
		bold:      color.New(color.Bold),
		highlight: PrimaryColor(),
		warn:      color.New(color.FgYellow),
		danger:    color.New(color.FgRed),
	}

	if info.ProductName != "" {
		h.reCmdRef = regexp.MustCompile(`"` + regexp.QuoteMeta(info.ProductName) + ` (\w\s?)+"`)
	}

	var buf bytes.Buffer
	h.writeUsage(&buf)
	h.writeDescription(&buf)
	h.writeFlags(&buf)
	h.writeExample(&buf)

	return h.colorize(buf.String())
}

func (h *helpBuilder) writeUsage(buf *bytes.Buffer) {
	buf.WriteString(strings.TrimSpace(h.info.Usage) + "\n")
}

func (h *helpBuilder) writeDescription(buf *bytes.Buffer) {
	if desc := strings.TrimSpace(h.info.Description); desc != "" {
		buf.WriteString("\n" + highlightCode(desc, h.info.ProductName) + "\n")
	}
}

func (h *helpBuilder) writeFlags(buf *bytes.Buffer) {
	if h.info.Flags == nil {
		return
	}

	title := h.info.Flags.Name() + ":"

	buf.WriteString("\n" + h.bold.Sprint(title) + "\n")

	var optBuf bytes.Buffer
	w := tabwriter.NewWriter(&optBuf, 0, 80, 0, ' ', 0)

	h.info.Flags.VisitAll(func(f *flag.Flag) {
		names := h.flagNames(f)
		usage := h.indentUsage(f.Usage)
		meta := h.flagMeta(f)

		fmt.Fprintf(w, "  %s\n%s\n", names, usage)
		for _, line := range meta {
			fmt.Fprintf(w, "      %s\n", line)
		}
		fmt.Fprintln(w)
	})

	w.Flush()
	buf.WriteString(optBuf.String())
}

// flagNames formats the primary name and aliases: -name*, -n
// Required flags get a red * suffix via Flag.Marker().
func (h *helpBuilder) flagNames(f *flag.Flag) string {
	var b strings.Builder

	for i, name := range f.Names() {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(h.highlight.Sprintf("-%s", name))
	}

	termColors := map[string]*color.Color{
		"red":    h.danger,
		"orange": h.warn,
		"green":  h.highlight,
	}

	for _, m := range f.Markers() {
		b.WriteByte(' ')
		if c, ok := termColors[m.Color()]; ok {
			b.WriteString(c.Sprint(m))
		} else {
			b.WriteString(m.String())
		}
	}

	return b.String()
}

// flagMeta builds metadata lines shown below the usage text.
func (h *helpBuilder) flagMeta(f *flag.Flag) []string {
	var lines []string

	if vals := f.ValidValues(); len(vals) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", h.bold.Sprint("Values:"), strings.Join(vals, ", ")))
	}

	if dv := f.DefaultValue(); dv != "" {
		lines = append(lines, fmt.Sprintf("%s %s", h.bold.Sprint("Default:"), dv))
	}

	if dm := f.DeprecationMessage(); dm != "" {
		lines = append(lines, h.bold.Sprint("Deprecated")+": "+dm)
	}

	if env := f.EnvVar(); env != "" {
		lines = append(lines, fmt.Sprintf("%s %s", h.bold.Sprint("Env:"), env))
	}

	return lines
}

func (h *helpBuilder) indentUsage(usage string) string {
	var b strings.Builder
	first := true
	for line := range strings.SplitSeq(text.Wrap(usage, 70), "\n") {
		if !first {
			b.WriteByte('\n')
		}
		b.WriteString(strings.Repeat(" ", 6))
		b.WriteString(line)
		first = false
	}

	return b.String()
}

func (h *helpBuilder) writeExample(buf *bytes.Buffer) {
	example := strings.TrimSuffix(strings.TrimPrefix(h.info.Example, "\n"), "\n")
	if example == "" {
		return
	}

	buf.WriteString("\n" + h.bold.Sprint("Example:") + "\n" + highlightCode(example, h.info.ProductName) + "\n\n")
}

// colorize applies line-level syntax highlighting to the assembled help text.
func (h *helpBuilder) colorize(raw string) string {
	v := strings.TrimSpace(raw)
	var buf bytes.Buffer

	seenHeader := false
	productPrefix := h.info.ProductName + " "

	for line := range strings.SplitSeq(v, "\n") {
		switch {
		case h.info.ProductName != "" && strings.HasPrefix(line, productPrefix):
			buf.WriteString(h.highlight.Sprint(h.info.ProductName))
			buf.WriteString(line[len(h.info.ProductName):])

		case strings.HasPrefix(line, "Usage: "):
			buf.WriteString(h.highlight.Sprint("Usage: "))
			buf.WriteString(line[7:])

		case reHelpHeader.MatchString(line):
			seenHeader = true
			buf.WriteString(h.bold.Sprint(line))

		default:
			if h.info.ProductName != "" {
				if s, ok := h.colorizeCmdRefs(line); ok {
					buf.WriteString(s)
					break
				}
			}

			if !seenHeader {
				if s, ok := h.colorizeFlags(line); ok {
					buf.WriteString(s)
					break
				}
			}

			buf.WriteString(line)
		}

		buf.WriteByte('\n')
	}

	return strings.TrimSuffix(buf.String(), "\n")
}

// colorizeCmdRefs highlights quoted command references like "product run".
func (h *helpBuilder) colorizeCmdRefs(line string) (string, bool) {
	matches := h.reCmdRef.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		return "", false
	}

	var buf strings.Builder
	idx := 0
	for _, m := range matches {
		buf.WriteString(line[idx : m[0]+1])
		buf.WriteString(h.highlight.Sprint(line[m[0]+1 : m[1]-1]))
		idx = m[1] - 1
	}

	buf.WriteString(line[idx:])

	return buf.String(), true
}

// colorizeFlags highlights -flag tokens in lines before the first header.
func (h *helpBuilder) colorizeFlags(line string) (string, bool) {
	matches := reHelpFlag.FindAllStringSubmatchIndex(line, -1)
	if len(matches) == 0 {
		return "", false
	}

	var buf strings.Builder
	idx := 0
	for _, m := range matches {
		start, end := m[4], m[5]
		buf.WriteString(line[idx:start])
		buf.WriteString(h.highlight.Sprint(line[start:end]))
		idx = end
	}

	buf.WriteString(line[idx:])

	return buf.String(), true
}

// highlightCode applies syntax highlighting to code blocks and shell commands.
func highlightCode(s, productName string) string {
	if reCodeBlock.MatchString(s) {
		return reCodeBlock.ReplaceAllStringFunc(s, func(match string) string {
			parts := reCodeBlock.FindStringSubmatch(match)
			if len(parts) < 4 || parts[2] == "" {
				return parts[3]
			}

			return highlight(parts[3], parts[2])
		})
	}

	s = strings.TrimSpace(s)
	if productName != "" && (strings.HasPrefix(s, productName+" ") || strings.HasPrefix(s, "./"+productName+" ")) {
		return highlight(s, "bash")
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
