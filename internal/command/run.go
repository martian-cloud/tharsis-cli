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

func (rc runCommand) Run(args []string) int {
	rc.meta.Logger.Debugf("Starting the 'run' command with %d arguments:", len(args))
	for ix, arg := range args {
		rc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	rc.meta.UI.Output(rc.HelpRun(true))
	return 1
}

func (rc runCommand) Synopsis() string {
	return "Do operations on runs."
}

func (rc runCommand) Help() string {
	return rc.HelpRun(false)
}

// HelpRun produces the help string for the 'run' command.
func (rc runCommand) HelpRun(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] run ...

   The run commands do operations on runs. Currently, allows
   cancelling runs.
`, rc.meta.BinaryName)
	sc := `

Subcommands:
    cancel    Cancel a run.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}
