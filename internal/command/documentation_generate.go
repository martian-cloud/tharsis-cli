package command

//go:generate go run ../../cmd/tharsis/tharsis.go documentation generate -output ../../docs/commands.md

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"
)

// Go templates for markdown generation.
const markdownTemplateForCommandList = `
# Tharsis CLI Commands

## Available Commands

Currently, the following commands are available:

` + "```bash" + `
{{block "list" .}}{{range .}}{{printf "%s   %s\n" .Name .Synopsis}}{{end}}{{end}}` + "```" + `{{"\n"}}{{"\n"}}
`

// For readability, newlines in this string are removed before it is parsed.
const markdownTemplateForCommandDetails = `
{{define "DetailElement"}}
{{if .IsSubcommand}}---{{"\n"}}{{"\n"}}#### {{.Name}}
{{else}}---{{"\n"}}{{"\n"}}## {{.Name}}{{end}}{{"\n"}}{{"\n"}}

{{.Synopsis}}{{"\n"}}{{"\n"}}

{{if .HasSubcommands}}
:::info Subcommands{{"\n"}}
{{range .Subcommands}}` + "- `" + `{{.Name}}` + "`" + ` - {{.Synopsis}}{{"\n"}}{{end}}
:::{{"\n"}}{{"\n"}}
{{end}}

{{if .UsageLine}}
` + "```bash" + `{{"\n"}}
{{.UsageLine}}{{"\n"}}
` + "```" + `{{"\n"}}
{{"\n"}}{{end}}

{{with .Description}}{{.}}{{"\n"}}{{"\n"}}{{end}}

{{if .Flags}}
{{with .Flags}}
<details>{{"\n"}}
<summary>Options</summary>{{"\n"}}{{"\n"}}
{{range .}}` + "- `" + `--{{.Name}}` + "`" + ` - {{.Description}}{{"\n\n"}}{{end}}{{end}}
</details>
{{"\n"}}{{"\n"}}
{{end}}

{{if .Example}}
:::note Example{{"\n"}}
` + "```bash" + `{{"\n"}}
{{.Example}}{{"\n"}}
` + "```" + `{{"\n"}}
:::{{"\n"}}
{{"\n"}}{{end}}

{{end}}

{{block "details" .}}{{range .}}{{template "DetailElement" .}}{{end}}{{end}}
`

type markdownCommandListElem struct {
	Name     string
	Synopsis string
}

type markdownCommandDetail struct {
	Name           string
	Synopsis       string
	UsageLine      string
	Description    string
	Example        string
	Flags          []markdownFlag
	IsSubcommand   bool
	HasSubcommands bool
	Subcommands    []markdownCommandListElem
}

type markdownFlag struct {
	Name        string
	Description string
}

var _ Command = (*documentationGenerateCommand)(nil)

// documentationGenerateCommand is the structure for documentation generate command.
type documentationGenerateCommand struct {
	*BaseCommand

	outputFilename *string
	allCommands    map[string]Factory
}

func (c *documentationGenerateCommand) validate() error {
	return nil
}

// NewDocumentationGenerateCommandFactory returns a documentationGenerateCommand struct.
func NewDocumentationGenerateCommandFactory(
	baseCommand *BaseCommand,
	allCommands map[string]Factory,
) func() (Command, error) {
	return func() (Command, error) {
		return &documentationGenerateCommand{
			BaseCommand: baseCommand,
			allCommands: allCommands,
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

	generator, err := newMarkdownGenerator()
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to create generator: %s", err))
		return 1
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

	if err := generator.Generate(writer, c.allCommands); err != nil {
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

func (c *documentationGenerateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("command options", flag.ContinueOnError)
	f.Func(
		"output",
		"The output filename.",
		func(s string) error {
			c.outputFilename = &s
			return nil
		},
	)

	return f
}

// markdownGenerator handles markdown documentation generation
type markdownGenerator struct {
	listTemplate   *template.Template
	detailTemplate *template.Template
}

func newMarkdownGenerator() (*markdownGenerator, error) {
	listTmpl, err := template.New("command-list").Parse(strings.TrimSpace(markdownTemplateForCommandList))
	if err != nil {
		return nil, fmt.Errorf("failed to parse command list template: %w", err)
	}

	detailTmpl, err := template.New("command-detail").Parse(strings.ReplaceAll(markdownTemplateForCommandDetails, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("failed to parse command details template: %w", err)
	}

	return &markdownGenerator{
		listTemplate:   listTmpl,
		detailTemplate: detailTmpl,
	}, nil
}

func (g *markdownGenerator) Generate(writer io.Writer, allCommands map[string]Factory) error {
	commands, err := g.instantiateCommands(allCommands)
	if err != nil {
		return err
	}

	list := g.buildCommandList(commands)
	if err := g.listTemplate.Execute(writer, list); err != nil {
		return fmt.Errorf("failed to execute command list template: %w", err)
	}

	details := g.buildCommandDetails(commands)
	if err := g.detailTemplate.Execute(writer, details); err != nil {
		return fmt.Errorf("failed to execute command details template: %w", err)
	}

	return nil
}

func (g *markdownGenerator) instantiateCommands(allCommands map[string]Factory) (map[string]Command, error) {
	commandNames := make([]string, 0, len(allCommands))
	for name := range allCommands {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	commands := make(map[string]Command, len(allCommands))
	for _, name := range commandNames {
		command, err := allCommands[name]()
		if err != nil {
			return nil, fmt.Errorf("failed to create command %s: %w", name, err)
		}
		commands[name] = command
	}

	return commands, nil
}

func (g *markdownGenerator) buildCommandList(commands map[string]Command) []markdownCommandListElem {
	commandNames := g.getSortedCommandNames(commands)
	longestName := g.getLongestNameLength(commandNames)
	formatString := fmt.Sprintf("%%-%ds", longestName)

	list := make([]markdownCommandListElem, 0)
	for _, name := range commandNames {
		// Don't add subcommands to the summary list
		if !strings.Contains(name, " ") {
			list = append(list, markdownCommandListElem{
				Name:     fmt.Sprintf(formatString, name),
				Synopsis: commands[name].Synopsis(),
			})
		}
	}

	return list
}

func (g *markdownGenerator) buildCommandDetails(commands map[string]Command) []markdownCommandDetail {
	commandNames := g.getSortedCommandNames(commands)
	longestName := g.getLongestNameLength(commandNames)
	formatString := fmt.Sprintf("%%-%ds", longestName)

	isSubcommand, hasSubcommands, subCommands := g.analyzeSubcommands(commands, commandNames, formatString)

	details := make([]markdownCommandDetail, 0, len(commandNames))
	for _, name := range commandNames {
		command := commands[name]
		detail := markdownCommandDetail{
			Name:           name,
			Synopsis:       command.Synopsis(),
			UsageLine:      strings.TrimSpace(command.Usage()),
			Description:    sanitizeForMarkdown(strings.TrimSuffix(strings.TrimPrefix(command.Description(), "\n"), "\n")),
			Example:        strings.TrimSpace(command.Example()),
			Flags:          extractFlags(command),
			IsSubcommand:   isSubcommand[name],
			HasSubcommands: hasSubcommands[name],
			Subcommands:    subCommands[name],
		}
		details = append(details, detail)
	}

	return details
}

func (g *markdownGenerator) analyzeSubcommands(commands map[string]Command, commandNames []string, formatString string) (
	map[string]bool,
	map[string]bool,
	map[string][]markdownCommandListElem,
) {
	isSubcommand := make(map[string]bool)
	hasSubcommands := make(map[string]bool)
	subCommands := make(map[string][]markdownCommandListElem)

	for _, name := range commandNames {
		idx := strings.LastIndex(name, " ")
		if idx == -1 {
			continue
		}

		isSubcommand[name] = true
		parentName := name[:idx]
		subName := name[idx+1:]

		if _, exists := commands[parentName]; exists {
			hasSubcommands[parentName] = true
		}

		if _, ok := subCommands[parentName]; !ok {
			subCommands[parentName] = make([]markdownCommandListElem, 0)
		}
		subCommands[parentName] = append(subCommands[parentName], markdownCommandListElem{
			Name:     fmt.Sprintf(formatString, subName),
			Synopsis: commands[name].Synopsis(),
		})
	}

	return isSubcommand, hasSubcommands, subCommands
}

func (g *markdownGenerator) getSortedCommandNames(commands map[string]Command) []string {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (g *markdownGenerator) getLongestNameLength(names []string) int {
	longest := 0
	for _, name := range names {
		if len(name) > longest {
			longest = len(name)
		}
	}
	return longest
}

func extractFlags(command Command) []markdownFlag {
	flags := make([]markdownFlag, 0)
	flagSet := command.Flags()
	if flagSet != nil {
		flagSet.VisitAll(func(f *flag.Flag) {
			flags = append(flags, markdownFlag{
				Name:        f.Name,
				Description: f.Usage,
			})
		})
	}
	return flags
}

func sanitizeForMarkdown(s string) string {
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "<", "\\<")
	s = strings.ReplaceAll(s, ">", "\\>")
	return s
}
