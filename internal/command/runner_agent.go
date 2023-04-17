package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// runnerAgentCommand is the top-level structure for the runner-agent command.
type runnerAgentCommand struct {
	meta *Metadata
}

// NewRunnerAgentCommandFactory returns a runnerAgentCommand struct.
func NewRunnerAgentCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runnerAgentCommand{
			meta: meta,
		}, nil
	}
}

func (ra runnerAgentCommand) Run(args []string) int {
	ra.meta.Logger.Debugf("Starting the 'runner-agent' command with %d arguments:", len(args))
	for ix, arg := range args {
		ra.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	ra.meta.UI.Output(ra.HelpRunnerAgent(true))
	return 1
}

func (runnerAgentCommand) Synopsis() string {
	return "Do operations on runner agents."
}

func (ra runnerAgentCommand) Help() string {
	return ra.HelpRunnerAgent(false)
}

// HelpRunnerAgent produces the help string for the 'runner-agent' command.
func (ra runnerAgentCommand) HelpRunnerAgent(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] runner-agent ...

   The runner-agent commands do operations on runner agents.
   Runner agents are responsible for launching Terraform jobs
   that deploy your infrastructure to the cloud.
`, ra.meta.BinaryName)
	sc := `

Subcommands:
    assign-service-account      Assign a service account to a runner agent.
    create                      Create a new runner agent.
    delete                      Delete a runner agent.
    get                         Get a single runner agent.
    unassign-service-account    Unassign a service account to a runner agent.
    update                      Update a runner agent.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

// The End.
