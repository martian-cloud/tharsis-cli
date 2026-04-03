// Package output provides formatting utilities for CLI output.
package output

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/kr/text"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// maxHelpTextWidth is the max line width for help text such as
// descriptions and flag usage, accounting for indentation.
const maxHelpTextWidth = 70

// CommandHelpInfo contains the components of command help text.
type CommandHelpInfo struct {
	ProductName string
	Usage       string
	Description string
	Flags       *flag.Set
	Example     string
}

// CommandHelp builds and formats help text for a command with syntax highlighting.
func CommandHelp(info CommandHelpInfo) string {
	var sections []string

	// Usage.
	if usage := strings.TrimSpace(info.Usage); usage != "" {
		sections = append(sections, usage)
	}

	// Description.
	if desc := normalizeDescription(info.Description); desc != "" {
		sections = append(sections, desc)
	}

	// Flags.
	if info.Flags != nil {
		sections = append(sections, info.Flags.Name()+":\n"+formatFlags(info.Flags))
	}

	// Example.
	if example := strings.TrimSpace(info.Example); example != "" {
		sections = append(sections, "Example:\n"+example)
	}

	return Colorize(strings.Join(sections, "\n\n"), info.ProductName)
}

// Wrap wraps text to fit the terminal width.
func Wrap(s string) string {
	return text.Wrap(s, maxHelpTextWidth)
}

// normalizeDescription trims and indents description text with 2-space indent.
func normalizeDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return ""
	}

	lines := strings.Split(desc, "\n")
	for i, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines[i] = "  " + trimmed
		} else {
			lines[i] = ""
		}
	}

	return strings.Join(lines, "\n")
}

func formatFlags(flagSet *flag.Set) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 80, 0, ' ', 0)

	flagSet.VisitAll(func(f *flag.Flag) {
		// Flag names.
		names := f.Names()
		for i, name := range names {
			names[i] = "-" + name
		}

		parts := []string{strings.Join(names, ", ")}
		for _, m := range f.Markers() {
			parts = append(parts, m.String())
		}

		// Usage text.
		var usage strings.Builder
		first := true
		for line := range strings.SplitSeq(text.Wrap(f.Usage, maxHelpTextWidth), "\n") {
			if !first {
				usage.WriteByte('\n')
			}
			usage.WriteString("      ")
			usage.WriteString(line)
			first = false
		}

		fmt.Fprintf(w, "  %s\n%s\n", strings.Join(parts, " "), usage.String())

		// Metadata lines.
		if vals := f.ValidValues(); len(vals) > 0 {
			fmt.Fprintf(w, "      Values: %s\n", strings.Join(vals, ", "))
		}

		if dv := f.DefaultValue(); dv != "" {
			fmt.Fprintf(w, "      Default: %s\n", dv)
		}

		if dm := f.DeprecationMessage(); dm != "" {
			fmt.Fprintf(w, "      Deprecated: %s\n", dm)
		}

		if env := f.EnvVar(); env != "" {
			fmt.Fprintf(w, "      Env: %s\n", env)
		}

		if ex := f.ExclusiveWith(); len(ex) > 0 {
			names := make([]string, len(ex))
			for i, name := range ex {
				names[i] = "-" + name
			}

			fmt.Fprintf(w, "      Conflicts: %s\n", strings.Join(names, ", "))
		}

		fmt.Fprintln(w)
	})

	w.Flush()

	return strings.TrimRight(buf.String(), "\n")
}
