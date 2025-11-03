package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// runnerAgentDeleteCommand is the top-level structure for the runner-agent delete command.
type runnerAgentDeleteCommand struct {
	meta *Metadata
}

// NewRunnerAgentDeleteCommandFactory returns a runnerAgentDeleteCommand struct.
func NewRunnerAgentDeleteCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runnerAgentDeleteCommand{
			meta: meta,
		}, nil
	}
}

func (rdc runnerAgentDeleteCommand) Run(args []string) int {
	rdc.meta.Logger.Debugf("Starting the 'runner-agent delete' command with %d arguments:", len(args))
	for ix, arg := range args {
		rdc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := rdc.meta.GetSDKClient()
	if err != nil {
		rdc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return rdc.doRunnerAgentDelete(ctx, client, args)
}

func (rdc runnerAgentDeleteCommand) doRunnerAgentDelete(ctx context.Context, client *tharsis.Client, opts []string) int {
	rdc.meta.Logger.Debugf("will do runner-agent delete, %d opts", len(opts))

	_, cmdArgs, err := optparser.ParseCommandOptions(rdc.meta.BinaryName+" runner-agent delete", optparser.OptionDefinitions{}, opts)
	if err != nil {
		rdc.meta.Logger.Error(output.FormatError("failed to parse runner-agent delete options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		rdc.meta.Logger.Error(output.FormatError("missing runner-agent delete id", nil), rdc.HelpRunnerAgentDelete())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive runner-agent delete arguments: %s", cmdArgs)
		rdc.meta.Logger.Error(output.FormatError(msg, nil), rdc.HelpRunnerAgentDelete())
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.DeleteRunnerInput{ID: cmdArgs[0]}
	rdc.meta.Logger.Debugf("runner-agent delete input: %#v", input)

	// Delete the runner agent.
	err = client.RunnerAgent.DeleteRunnerAgent(ctx, input)
	if err != nil {
		rdc.meta.Logger.Error(output.FormatError("failed to delete runner agent", err))
		return 1
	}

	// Cannot show the deleted runner agent, but say something.
	rdc.meta.UI.Output("runner agent delete succeeded.")

	return 0
}

func (rdc runnerAgentDeleteCommand) Synopsis() string {
	return "Delete a runner agent."
}

func (rdc runnerAgentDeleteCommand) Help() string {
	return rdc.HelpRunnerAgentDelete()
}

// HelpRunnerAgentDelete prints the help string for the 'runner-agent delete' command.
func (rdc runnerAgentDeleteCommand) HelpRunnerAgentDelete() string {
	return fmt.Sprintf(`
Usage: %s [global options] runner-agent delete <id>

   The runner-agent delete command deletes a runner agent.

`, rdc.meta.BinaryName)
}
