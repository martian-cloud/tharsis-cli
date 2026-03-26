package command

import (
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

// NewWrapper creates a new Wrapper for a Command.
func NewWrapper(command Command) Wrapper {
	return Wrapper{command: command, productName: "tharsis"}
}

// Help formats and returns the help text.
func (c Wrapper) Help() string {
	return output.CommandHelp(output.CommandHelpInfo{
		ProductName: c.productName,
		Usage:       c.command.Usage(),
		Description: c.command.Description(),
		Flags:       c.command.Flags(),
		Example:     c.command.Example(),
	})
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
		if vals := f.ValidValues(); len(vals) > 0 {
			predictor = complete.PredictSet(vals...)
		} else {
			predictor = complete.PredictAnything
		}

		result["-"+f.Name] = predictor
		for _, alias := range f.Aliases() {
			result["-"+alias] = predictor
		}
	})

	return result
}
