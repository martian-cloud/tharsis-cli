package output

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/fatih/color"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

var (
	reHelpHeader = regexp.MustCompile(`^[a-zA-Z0-9_-].*:$`)                      // matches "Section:" style headers
	reHelpFlag   = regexp.MustCompile(`(\s|^|")(-[\w-]+)(\W|$)`)                 // matches -flag style options
	reCodeBlock  = regexp.MustCompile("(?s)([ ]*)```(\\w*)\\n(.*?)```")          // matches ```lang\ncode``` with optional indent
	reFlagName   = regexp.MustCompile(`^  -[\w-]+`)                              // matches flag definition lines
	reFlagMeta   = regexp.MustCompile(`^      (Values|Default|Deprecated|Env):`) // matches metadata labels
)

// PrimaryColor returns a new Color instance with the primary brand color.
func PrimaryColor() *color.Color {
	return color.New(color.FgHiGreen)
}

// colorizer applies syntax highlighting to assembled help text.
type colorizer struct {
	productName string
	bold        *color.Color
	highlight   *color.Color
	reCmdRef    *regexp.Regexp
}

// Colorize applies all syntax highlighting to help text: product name,
// flags, section headers, code blocks, and quoted command references.
// Returns the input unchanged when color is disabled.
func Colorize(raw, productName string) string {
	if color.NoColor {
		return raw
	}

	c := &colorizer{
		productName: productName,
		bold:        color.New(color.Bold),
		highlight:   PrimaryColor(),
	}

	if productName != "" {
		c.reCmdRef = regexp.MustCompile(`"` + regexp.QuoteMeta(productName) + ` (\w\s?)+"`)
	}

	return c.colorize(raw)
}

func (c *colorizer) colorize(raw string) string {
	var buf bytes.Buffer

	seenHeader := false
	productPrefix := c.productName + " "

	// Lines are processed individually because colorization rules are
	// context-dependent (e.g. flag highlighting stops after the first header).
	// Newlines are written as separators to preserve the input's structure exactly.
	first := true
	for line := range strings.SplitSeq(raw, "\n") {
		if !first {
			buf.WriteByte('\n')
		}

		first = false

		switch {
		case c.productName != "" && strings.HasPrefix(line, productPrefix):
			buf.WriteString(c.highlight.Sprint(c.productName))
			buf.WriteString(line[len(c.productName):])

		case reHelpHeader.MatchString(line):
			seenHeader = true
			buf.WriteString(c.bold.Sprint(line))

		case reFlagName.MatchString(line):
			buf.WriteString(c.colorizeFlagDef(line))

		case reFlagMeta.MatchString(line):
			buf.WriteString(c.colorizeFlagMeta(line))

		default:
			if c.productName != "" {
				if s, ok := c.colorizeCmdRefs(line); ok {
					buf.WriteString(s)
					break
				}
			}

			if !seenHeader {
				if s, ok := c.colorizeFlags(line); ok {
					buf.WriteString(s)
					break
				}
			}

			buf.WriteString(c.highlightCode(line))
		}
	}

	return buf.String()
}

// colorizeFlagDef highlights flag names in green and markers in their colors.
// Input: "  -name, -alias * ..."
func (c *colorizer) colorizeFlagDef(line string) string {
	warn := color.New(color.FgYellow)
	danger := color.New(color.FgRed)

	markerColors := map[flag.Marker]*color.Color{
		flag.MarkerRequired:   danger,
		flag.MarkerDeprecated: warn,
		flag.MarkerRepeatable: c.highlight,
	}

	// Split into the flag names part and any trailing markers.
	trimmed := strings.TrimSpace(line)
	parts := strings.Fields(trimmed)

	var result strings.Builder
	result.WriteString("  ")

	for i, part := range parts {
		if i > 0 {
			result.WriteByte(' ')
		}

		if strings.HasPrefix(part, "-") {
			// Flag name or alias — highlight in green.
			// Preserve trailing comma if present.
			name := strings.TrimSuffix(part, ",")
			result.WriteString(c.highlight.Sprint(name))
			if strings.HasSuffix(part, ",") {
				result.WriteByte(',')
			}
		} else if mc, ok := markerColors[flag.Marker(part)]; ok {
			result.WriteString(mc.Sprint(part))
		} else {
			result.WriteString(part)
		}
	}

	return result.String()
}

// colorizeFlagMeta bolds the label portion of metadata lines.
// Input: "      Values: a, b, c"
func (c *colorizer) colorizeFlagMeta(line string) string {
	loc := reFlagMeta.FindStringIndex(line)
	if loc == nil || len(line) < loc[1] {
		return line
	}

	// The regex guarantees at least 6 leading spaces + label.
	prefix := line[:6]
	label := line[6:loc[1]]
	rest := line[loc[1]:]

	return prefix + c.bold.Sprint(label) + rest
}

// colorizeCmdRefs highlights quoted command references like "product run".
func (c *colorizer) colorizeCmdRefs(line string) (string, bool) {
	if c.reCmdRef == nil {
		return "", false
	}

	matches := c.reCmdRef.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		return "", false
	}

	var buf strings.Builder
	idx := 0
	for _, m := range matches {
		if m[1]-m[0] < 3 {
			buf.WriteString(line[idx:m[1]])
			idx = m[1]
			continue
		}

		buf.WriteString(line[idx : m[0]+1])
		buf.WriteString(c.highlight.Sprint(line[m[0]+1 : m[1]-1]))
		idx = m[1] - 1
	}

	buf.WriteString(line[idx:])

	return buf.String(), true
}

// colorizeFlags highlights -flag tokens in lines before the first header.
func (c *colorizer) colorizeFlags(line string) (string, bool) {
	matches := reHelpFlag.FindAllStringSubmatchIndex(line, -1)
	if len(matches) == 0 {
		return "", false
	}

	var buf strings.Builder
	idx := 0
	for _, m := range matches {
		start, end := m[4], m[5]
		buf.WriteString(line[idx:start])
		buf.WriteString(c.highlight.Sprint(line[start:end]))
		idx = end
	}

	buf.WriteString(line[idx:])

	return buf.String(), true
}

// highlightCode applies syntax highlighting to code blocks and shell commands.
func (c *colorizer) highlightCode(s string) string {
	if reCodeBlock.MatchString(s) {
		return reCodeBlock.ReplaceAllStringFunc(s, func(match string) string {
			parts := reCodeBlock.FindStringSubmatch(match)
			if len(parts) < 4 || parts[2] == "" {
				return parts[3]
			}

			return chromaHighlight(parts[3], parts[2])
		})
	}

	trimmed := strings.TrimSpace(s)
	if c.productName != "" && (strings.HasPrefix(trimmed, c.productName+" ") || strings.HasPrefix(trimmed, "./"+c.productName+" ")) {
		return chromaHighlight(trimmed, "bash")
	}

	return s
}

func chromaHighlight(code, lang string) string {
	var buf bytes.Buffer
	if err := quick.Highlight(&buf, strings.TrimSpace(code), lang, "terminal16m", "monokai"); err != nil {
		return code
	}

	return strings.TrimSpace(buf.String())
}
