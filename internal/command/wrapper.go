package command

import (
	"github.com/fatih/color"
	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
)

// Command wraps a CLI command.
type Command interface {
	Usage() string
	Description() string
	Example() string
	Flags() *flag.Set
	PredictArgs() complete.Predictor
	Run(args []string) int
	Synopsis() string
}

// Wrapper implements cli.Command and wraps a Command.
type Wrapper struct {
	command     Command
	productName string
}

var (
	_ cli.Command             = (*Wrapper)(nil)
	_ cli.CommandAutocomplete = (*Wrapper)(nil)
)

// Factory is a function that returns a Command.
type Factory func() (Command, error)

// helpInterceptor is implemented by commands that want to customize help output.
// InterceptHelp returns (helpText, true) to replace the default Tharsis-formatted
// help, or ("", false) to fall back to standard help rendering. This allows
// commands to forward help to an external binary (e.g. terraform) when appropriate.
type helpInterceptor interface {
	InterceptHelp() (string, bool)
}

// NewWrapper creates a new Wrapper for a Command.
func NewWrapper(command Command) Wrapper {
	return Wrapper{command: command, productName: "tharsis"}
}

// Help formats and returns the help text. If the wrapped command implements
// helpInterceptor and returns handled=true, that text is used directly.
func (c Wrapper) Help() string {
	if hi, ok := c.command.(helpInterceptor); ok {
		if text, handled := hi.InterceptHelp(); handled {
			return text
		}
	}
	return output.CommandHelp(output.CommandHelpInfo{
		ProductName: c.productName,
		Usage:       c.command.Usage(),
		Description: c.command.Description(),
		Flags:       c.command.Flags(),
		Example:     c.command.Example(),
	})
}

// HelpTemplate returns a custom help template if the wrapped command provides one.
// Otherwise it returns the default template with bold subcommand headers.
func (c Wrapper) HelpTemplate() string {
	if t, ok := c.command.(interface{ HelpTemplate() string }); ok {
		return t.HelpTemplate()
	}

	return `{{.Help}}
{{- if gt (len .Subcommands) 0}}

` + color.New(color.Bold).Sprint("Subcommands:") + `
{{- range $value := .Subcommands }}
    {{ $value.NameAligned }}    {{ $value.Synopsis }}{{ end }}
{{- end }}
`
}

// Run executes the command.
func (c Wrapper) Run(args []string) int {
	return c.command.Run(args)
}

// Synopsis returns a short description of the command.
func (c Wrapper) Synopsis() string {
	return c.command.Synopsis()
}

// AutocompleteArgs returns argument completions (none by default).
func (c Wrapper) AutocompleteArgs() complete.Predictor {
	return c.command.PredictArgs()
}

// AutocompleteFlags returns flag completions derived from the command's Flags().
func (c Wrapper) AutocompleteFlags() complete.Flags {
	fs := c.command.Flags()
	if fs == nil {
		return nil
	}

	result := make(complete.Flags)
	fs.VisitAll(func(f *flag.Flag) {
		var predictor complete.Predictor
		if f.IsBool() {
			predictor = complete.PredictNothing
		} else if preds := f.Predictors(); len(preds) > 0 {
			predictor = complete.PredictSet(preds...)
		} else {
			predictor = complete.PredictAnything
		}

		for _, name := range f.Names() {
			result["-"+name] = predictor
		}
	})

	return result
}
