package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// runCommand is the top-level structure for the run command.
type runCommand struct {
	meta *Metadata
}

// NewRunCommandFactory returns a runCommand struct.
func NewRunCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runCommand{
			meta: meta,
		}, nil
	}
}

func (wc runCommand) Run(args []string) int {
	wc.meta.Logger.Debugf("Starting the 'run' command with %d arguments:", len(args))
	for ix, arg := range args {
		wc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	wc.meta.UI.Output(wc.HelpRun(true))
	return 1
}

func (wc runCommand) Synopsis() string {
	return "Do operations on runs."
}

func (wc runCommand) Help() string {
	return wc.HelpRun(false)
}

// HelpRun produces the help string for the 'run' command.
func (wc runCommand) HelpRun(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] run ...

   The run commands do operations on runs. Currently, allows
   cancelling runs.
`, wc.meta.BinaryName)
	sc := `

Subcommands:
    cancel    Cancel a run.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

// The End.
