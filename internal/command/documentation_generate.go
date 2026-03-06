package command

//go:generate go run ../../cmd/tharsis/tharsis.go documentation generate -output ../../docs/commands.md

import (
	"flag"
	"fmt"
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

` + "```" + `
{{block "list" .}}{{range .}}{{printf "%s   %s\n" .Name .Synopsis}}{{end}}{{end}}` + "```" + `{{"\n"}}{{"\n"}}
`

// For readability, newlines in this string are removed before it is parsed.
const markdownTemplateForCommandDetails = `
{{define "DetailElement"}}
{{if .IsSubcommand}}---{{"\n"}}{{"\n"}}#### {{.Name}}
{{else}}***{{"\n"}}{{"\n"}}### Command: {{.Name}}{{end}}{{"\n"}}{{"\n"}}

##### {{.Synopsis}}{{"\n"}}{{"\n"}}

{{if .HasSubcommands}}
**Subcommands:**{{"\n"}}{{"\n"}}` + "```" + `{{"\n"}}
{{range .Subcommands}}{{printf "%s   %s\n" .Name .Synopsis}}{{end}}` + "```" + `{{"\n"}}{{"\n"}}
{{end}}

{{if .UsageLine}}
` + "```" + `{{"\n"}}
Usage: {{.UsageLine}}{{"\n"}}
` + "```" + `{{"\n"}}
{{"\n"}}{{end}}

{{with .Description}}{{.}}{{"\n"}}{{"\n"}}{{end}}

{{if .Flags}}
{{with .Flags}}
<details>{{"\n"}}
<summary>Expand options</summary>{{"\n"}}{{"\n"}}
{{range .}}{{printf "- ` + "`" + `--%s` + "`" + `: %s\n\n" .Name .Description}}{{end}}{{end}}
</details>
{{"\n"}}{{"\n"}}
{{end}}

{{if .Example}}##### Example:{{"\n"}}{{"\n"}}
` + "```" + `{{"\n"}}
{{.Example}}{{"\n"}}
` + "```" + `{{"\n"}}
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

	outputFilename string
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

	oFile := os.Stdout
	if c.outputFilename != "" {
		var err error
		oFile, err = os.Create(c.outputFilename)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to create output file: %s", err))
			return 1
		}
	}

	// If a preamble is desired, it could be added or copied in at this point.

	commandNames := make([]string, 0, len(c.allCommands))
	for name := range c.allCommands {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	// Run the factories to cache the commands.
	commands := make(map[string]Command, len(c.allCommands))
	for _, name := range commandNames {
		command, err := c.allCommands[name]()
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to create command: %s", name))
			return 1
		}

		commands[name] = command
	}

	// One line per command.
	commandList := make([]markdownCommandListElem, 0, len(commandNames))
	longestName := 0
	for _, name := range commandNames {
		if len(name) > longestName {
			longestName = len(name)
		}
	}
	formatString := fmt.Sprintf("%%-%ds", longestName) // This is used a few paragraphs below, as well.
	for _, name := range commandNames {
		// Don't add subcommands to the summary list.
		if !strings.Contains(name, " ") {
			commandList = append(commandList, markdownCommandListElem{
				Name:     fmt.Sprintf(formatString, name),
				Synopsis: commands[name].Synopsis(),
			})
		}
	}

	listTemplate, err := template.New("command-list").Parse(strings.TrimSpace(markdownTemplateForCommandList))
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to parse command list template: %s", err))
		return 1
	}

	if eErr := listTemplate.Execute(oFile, commandList); eErr != nil {
		c.Logger.Error(fmt.Sprintf("Failed to execute command list template: %s", eErr))
		return 1
	}

	// If a middle section is desired (general comments about command syntax, etc.), it could be added or copied in at this point.

	// Scan to determine which commands have subcommands.
	isSubcommand := make(map[string]bool, len(c.allCommands))
	hasSubcommands := make(map[string]bool, len(c.allCommands))
	subCommands := make(map[string][]markdownCommandListElem, len(c.allCommands))
	for _, name := range commandNames {
		if strings.Contains(name, " ") {
			isSubcommand[name] = true
			nameParts := strings.Split(name, " ")

			// Determine the parent name by removing the last part
			// For "project variable-set create", parent is "project variable-set"
			// For "project create", parent is "project"
			parentName := strings.Join(nameParts[:len(nameParts)-1], " ")
			subName := nameParts[len(nameParts)-1]

			if _, ok := c.allCommands[parentName]; ok {
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
	}

	// One section per command.
	commandDetailList := make([]markdownCommandDetail, 0, len(commandNames))
	for _, name := range commandNames {
		command := commands[name]
		listElem := markdownCommandDetail{
			Name:           name,
			Synopsis:       command.Synopsis(),
			UsageLine:      strings.TrimSpace(command.Usage()),
			Description:    c.sanitizeForMarkdown(strings.TrimSuffix(strings.TrimPrefix(command.Description(), "\n"), "\n")),
			Example:        strings.TrimSpace(command.Example()),
			Flags:          make([]markdownFlag, 0),
			IsSubcommand:   isSubcommand[name],
			HasSubcommands: hasSubcommands[name],
			Subcommands:    subCommands[name],
		}

		// Get all the flags in proper order.
		flagSet := command.Flags()
		if flagSet != nil {
			flagSet.VisitAll(func(f *flag.Flag) {
				listElem.Flags = append(listElem.Flags, markdownFlag{
					Name:        f.Name,
					Description: f.Usage,
				})
			})
		}

		commandDetailList = append(commandDetailList, listElem)
	}

	detailTemplate, err := template.New("command-detail").Parse(strings.Replace(markdownTemplateForCommandDetails, "\n", "", -1))
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to parse command details template: %s", err))
		return 1
	}

	if eErr := detailTemplate.Execute(oFile, commandDetailList); eErr != nil {
		c.Logger.Error(fmt.Sprintf("Failed to execute command details template: %s", eErr))
		return 1
	}

	// If a FAQ or other material should follow the command descriptions, it could be added or copied in at this point.

	return 0
}

// sanitizeForMarkdown escapes brackets to hide them from the Markdown interpreter.
func (c *documentationGenerateCommand) sanitizeForMarkdown(s string) string {
	s = strings.Replace(s, "[", "\\[", -1)
	s = strings.Replace(s, "]", "\\]", -1)
	s = strings.Replace(s, "<", "\\<", -1)
	s = strings.Replace(s, ">", "\\>", -1)

	return s
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
	f.StringVar(
		&c.outputFilename,
		"output",
		"",
		"The output filename.",
	)

	return f
}
