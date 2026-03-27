package command

//go:generate go run ../../cmd/tharsis/... documentation generate -output ../../docs/commands.md

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	md "github.com/nao1215/markdown"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

var markdownSanitizer = strings.NewReplacer(
	"[", "\\[",
	"]", "\\]",
	"<", "\\<",
	">", "\\>",
	"{", "\\{",
	"}", "\\}",
)

var markerHTMLColors = map[flag.Marker]string{
	flag.MarkerRequired:   "red",
	flag.MarkerDeprecated: "orange",
	flag.MarkerRepeatable: "green",
}

// documentationGenerateCommand is the structure for documentation generate command.
type documentationGenerateCommand struct {
	*BaseCommand

	outputFilename *string
	allCommands    map[string]Factory
	globalFlags    *flag.Set
}

var _ Command = (*documentationGenerateCommand)(nil)

func (c *documentationGenerateCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
}

// NewDocumentationGenerateCommandFactory returns a documentationGenerateCommand struct.
func NewDocumentationGenerateCommandFactory(
	baseCommand *BaseCommand,
	allCommands map[string]Factory,
	globalFlags *flag.Set,
) func() (Command, error) {
	return func() (Command, error) {
		return &documentationGenerateCommand{
			BaseCommand: baseCommand,
			allCommands: allCommands,
			globalFlags: globalFlags,
		}, nil
	}
}

func (c *documentationGenerateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("documentation generate"),
		WithFlags(c.Flags()),
		WithInputValidator(c.validate),
	); code != 0 {
		return code
	}

	writer := os.Stdout
	if c.outputFilename != nil {
		file, err := os.Create(*c.outputFilename)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to create output file: %s", err))
			return 1
		}
		defer file.Close()
		writer = file
	}

	if err := generateDocumentation(writer, c.allCommands, c.globalFlags); err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to generate documentation: %s", err))
		return 1
	}

	return 0
}

func (c *documentationGenerateCommand) Synopsis() string {
	return "Generate documentation of commands."
}

func (c *documentationGenerateCommand) Usage() string {
	return "tharsis [global options] documentation generate"
}

func (c *documentationGenerateCommand) Description() string {
	return `
  The documentation generate command generates markdown documentation
  for the entire CLI.
`
}

func (c *documentationGenerateCommand) Example() string {
	return `
tharsis documentation generate
`
}

func (c *documentationGenerateCommand) Flags() *flag.Set {
	f := flag.NewSet("command options")
	f.StringVar(
		&c.outputFilename,
		"output",
		"The output filename.",
	)

	return f
}

func generateDocumentation(writer io.Writer, allCommands map[string]Factory, globalFlags *flag.Set) error {
	commands, err := instantiateCommands(allCommands)
	if err != nil {
		return err
	}

	names := sortedCommandNames(commands)
	isSubcommand, subCommands := analyzeSubcommands(commands, names)

	m := md.NewMarkdown(writer)

	// Frontmatter + command list.
	writeFrontmatter(m)
	writeCommandList(m, commands, names)

	// Global options.
	m.HorizontalRule().H2("Global Options").LF()
	writeFlags(m, globalFlags)

	// Command details.
	for _, name := range names {
		// Skip the documentation commands from public docs.
		if strings.HasPrefix(name, "documentation") {
			continue
		}

		cmd := commands[name]
		if isSubcommand[name] {
			m.HorizontalRule().H3f("%s subcommand", name)
		} else {
			m.HorizontalRule().H2f("%s command", name)
		}

		m.PlainTextf("**%s**", ensurePeriod(cmd.Synopsis())).LF()

		if subs, ok := subCommands[name]; ok {
			m.PlainText("**Subcommands:**").LF()
			items := make([]string, 0, len(subs))
			for _, sub := range subs {
				items = append(items, fmt.Sprintf("[`%s`](#%s) - %s", sub.name, sub.anchor, sub.synopsis))
			}
			m.BulletList(items...)
			m.LF()
		}

		if desc := sanitizeForMarkdown(cmd.Description()); desc != "" {
			m.PlainText(desc).LF()
		}

		if example := strings.TrimSpace(cmd.Example()); example != "" {
			writeExample(m, example)
		}

		if cmd.Flags() != nil {
			m.H4("Options").LF()
			writeFlags(m, cmd.Flags())
		}
	}

	// FAQ.
	m.HorizontalRule()
	writeFAQ(m)

	return m.Build()
}

func writeFrontmatter(m *md.Markdown) {
	m.PlainText("---").
		PlainText("title: Commands").
		PlainText(`description: "An introduction to the CLI commands"`).
		PlainText("---").LF()
}

func writeCommandList(m *md.Markdown, commands map[string]Command, names []string) {
	m.H2("Available Commands").
		PlainText("Currently, the CLI supports the following commands:").LF()

	items := make([]string, 0, len(names))
	for _, name := range names {
		if !strings.Contains(name, " ") && name != "documentation" {
			anchor := name + "-command"
			items = append(items, fmt.Sprintf("[%s](#%s) — %s", name, anchor, ensurePeriod(commands[name].Synopsis())))
		}
	}

	m.BulletList(items...).LF()

	m.PlainText(":::tip\n`tharsis [command]` or `tharsis [command] -h` will output the help menu for that specific command.\n:::")
	m.PlainText(":::info\nCommands and options may evolve between major versions. Options **must** come before any arguments.\n:::")
	m.PlainText(":::tip Have a question?\nCheck the [FAQ](#frequently-asked-questions-faq) to see if there's already an answer.\n:::")
	m.PlainText(":::info Legend\n" +
		"- <span style={{color:'red'}}>\\*&nbsp;&nbsp;</span> required\n" +
		"- <span style={{color:'orange'}}>!&nbsp;&nbsp;</span> deprecated\n" +
		"- <span style={{color:'green'}}>...</span> repeatable\n" +
		":::")
	m.LF()
}

func writeFlags(m *md.Markdown, flagSet *flag.Set) {
	var buf strings.Builder
	flagSet.VisitAll(func(f *flag.Flag) {
		parts := []string{strings.Join(f.Names(), ", ")}
		for _, mk := range f.Markers() {
			parts = append(parts, coloredMarker(mk))
		}

		buf.WriteString("#### " + strings.Join(parts, " ") + "\n\n")

		var meta []string
		if vals := f.ValidValues(); len(vals) > 0 {
			meta = append(meta, "**Values:** `"+strings.Join(vals, "`, `")+"`")
		}

		if dv := f.DefaultValue(); dv != "" {
			meta = append(meta, "**Default:** `"+dv+"`")
		}

		if dm := f.DeprecationMessage(); dm != "" {
			meta = append(meta, "**Deprecated**: "+dm)
		}

		if ev := f.EnvVar(); ev != "" {
			meta = append(meta, "**Env:** `"+ev+"`")
		}

		if len(meta) > 0 {
			buf.WriteString(f.Usage + "\\\n")
			buf.WriteString(strings.Join(meta, "\\\n"))
		} else {
			buf.WriteString(f.Usage)
		}

		buf.WriteString("\n\n")
	})

	if buf.Len() > 0 {
		m.PlainText(buf.String())
	}
}

// writeExample writes an example to the markdown output. Examples containing
// code block markers are written as-is; others are wrapped in a bash block.
func writeExample(m *md.Markdown, example string) {
	if strings.Contains(example, "```") {
		m.PlainText(example).LF()
	} else {
		m.CodeBlocks("bash", example).LF()
	}
}

func writeFAQ(m *md.Markdown) {
	m.H2("Frequently asked questions (FAQ)")

	m.H3("Is configuring a profile necessary?")
	m.PlainText("By default, the CLI will use the default Tharsis endpoint passed in at build-time. Unless a different endpoint is needed, no profile configuration is necessary. Simply run `tharsis sso login` and the `default` profile will be created and stored in the settings file.")

	m.H3("How do I use profiles?")
	m.PlainText("The profile can be specified using the `-p` global flag or the `THARSIS_PROFILE` environment variable. The flag **must** come before a command name. For example, `tharsis -p local group list` will list all the groups using the Tharsis endpoint in the `local` profile. Service accounts can use profiles in the same manner as human users.")

	m.H3("Where are the settings and credentials files located?")
	m.PlainText("The settings file is located at `~/.tharsis/settings.json` and contains profile configuration (endpoints, options). Credentials are stored separately in `~/.tharsis/credentials.json` so they can have stricter permissions.")
	m.PlainText(":::caution\n**Never** share the credentials file as it contains sensitive data like the authentication token from SSO!\n:::")

	m.H3("How do I disable colored output?")
	m.PlainText("Set the `NO_COLOR` environment variable to any value to disable colored output. For example, `NO_COLOR=1 tharsis group list`.")

	m.H3("Can I use Terraform variables from the CLI's environment inside a run?")
	m.PlainText("Yes, environment variables with the `TF_VAR_` prefix are passed as Terraform variables with the prefix stripped. For example, `TF_VAR_region=us-east-1` sets a Terraform variable named `region` to `us-east-1`.")
}

type subcommandInfo struct {
	name     string
	anchor   string
	synopsis string
}

func instantiateCommands(allCommands map[string]Factory) (map[string]Command, error) {
	commands := make(map[string]Command, len(allCommands))
	for name, factory := range allCommands {
		cmd, err := factory()
		if err != nil {
			return nil, fmt.Errorf("failed to create command %s: %w", name, err)
		}
		commands[name] = cmd
	}

	return commands, nil
}

func sortedCommandNames(commands map[string]Command) []string {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func analyzeSubcommands(commands map[string]Command, names []string) (map[string]bool, map[string][]subcommandInfo) {
	isSubcommand := make(map[string]bool)
	subCommands := make(map[string][]subcommandInfo)

	for _, name := range names {
		idx := strings.LastIndex(name, " ")
		if idx == -1 {
			continue
		}

		isSubcommand[name] = true
		parentName := name[:idx]

		if _, exists := commands[parentName]; exists {
			subCommands[parentName] = append(subCommands[parentName], subcommandInfo{
				name:     name[idx+1:],
				anchor:   strings.ReplaceAll(name, " ", "-") + "-subcommand",
				synopsis: commands[name].Synopsis(),
			})
		}
	}

	return isSubcommand, subCommands
}

func sanitizeForMarkdown(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	return markdownSanitizer.Replace(strings.Join(lines, "\n"))
}

func coloredMarker(m flag.Marker) string {
	if c, ok := markerHTMLColors[m]; ok {
		return fmt.Sprintf("<span style={{color:'%s'}}>%s</span>", c, m)
	}

	return m.String()
}

func ensurePeriod(s string) string {
	s = strings.TrimSpace(s)
	if s != "" && !strings.HasSuffix(s, ".") {
		s += "."
	}
	return s
}
