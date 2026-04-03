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
	reHelpHeader = regexp.MustCompile(`^[a-zA-Z0-9_-].*:$`)                                // matches "Section:" style headers
	reFlagMeta   = regexp.MustCompile(`^      (Values|Default|Deprecated|Env|Conflicts):`) // matches metadata labels
	reCodeBlock  = regexp.MustCompile("(?s)([ ]*)```(\\w+)\\n(.*?)```")                    // matches ```lang\ncode``` blocks
	reFlagToken  = regexp.MustCompile(`(\s)-(\w[\w-]*)`)                                   // matches -flag tokens in command lines
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
	// Process code blocks: strip ``` markers and apply syntax highlighting.
	// Runs before NoColor check because marker removal is structural.
	raw = processCodeBlocks(raw)

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

	productPrefix := c.productName + " "

	// Lines are processed individually because colorization rules are
	// context-dependent (e.g. section headers affect subsequent lines).
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
			buf.WriteString(c.colorizeFlags(line[len(c.productName):]))

		case reHelpHeader.MatchString(line):
			buf.WriteString(c.bold.Sprint(line))

		case reFlagMeta.MatchString(line):
			buf.WriteString(c.colorizeFlagMeta(line))

		default:
			if c.productName != "" {
				if s, ok := c.colorizeCmdRefs(line); ok {
					buf.WriteString(s)
					break
				}
			}

			buf.WriteString(c.colorizeFlags(line))
		}
	}

	return buf.String()
}

// colorizeFlags highlights -flag tokens in green and markers in their colors.
func (c *colorizer) colorizeFlags(s string) string {
	markers := map[string]*color.Color{
		flag.MarkerRequired.String():   color.New(color.FgRed),
		flag.MarkerDeprecated.String(): color.New(color.FgYellow),
		flag.MarkerRepeatable.String(): c.highlight,
	}

	// Colorize markers on flag definition lines before flag token
	// highlighting injects ANSI codes that break the prefix check.
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "  -") {
			for marker, mc := range markers {
				line = strings.ReplaceAll(line, " "+marker, " "+mc.Sprint(marker))
			}

			lines[i] = line
		}
	}

	// Highlight -flag tokens.
	return reFlagToken.ReplaceAllStringFunc(strings.Join(lines, "\n"), func(match string) string {
		i := strings.Index(match, "-")
		return match[:i] + c.highlight.Sprint(match[i:])
	})
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

	return prefix + c.bold.Sprint(label) + c.colorizeFlags(rest)
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

// processCodeBlocks strips ```lang markers and applies syntax highlighting.
func processCodeBlocks(s string) string {
	if !reCodeBlock.MatchString(s) {
		return s
	}

	return reCodeBlock.ReplaceAllStringFunc(s, func(match string) string {
		parts := reCodeBlock.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		return chromaHighlight(parts[3], parts[2])
	})
}

// chromaHighlight returns plain code when NoColor is set or
// highlights the code otherwise.
func chromaHighlight(code, lang string) string {
	code = strings.TrimSpace(code)

	if color.NoColor {
		return code
	}

	var buf bytes.Buffer
	if err := quick.Highlight(&buf, code, lang, "terminal16m", "monokai"); err != nil {
		return code
	}

	return strings.TrimSpace(buf.String())
}
