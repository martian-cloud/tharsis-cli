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

// runnerAgentAssignServiceAccountCommand is the top-level structure for the runner-agent assign-service-account command.
type runnerAgentAssignServiceAccountCommand struct {
	meta *Metadata
}

// NewRunnerAgentAssignServiceAccountCommandFactory returns a runnerAgentAssignServiceAccountCommand struct.
func NewRunnerAgentAssignServiceAccountCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runnerAgentAssignServiceAccountCommand{
			meta: meta,
		}, nil
	}
}

func (rac runnerAgentAssignServiceAccountCommand) Run(args []string) int {
	rac.meta.Logger.Debugf("Starting the 'runner-agent assign-service-account' command with %d arguments:", len(args))
	for ix, arg := range args {
		rac.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := rac.meta.GetSDKClient()
	if err != nil {
		rac.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return rac.doRunnerAgentAssignServiceAccount(ctx, client, args)
}

func (rac runnerAgentAssignServiceAccountCommand) doRunnerAgentAssignServiceAccount(ctx context.Context, client *tharsis.Client, opts []string) int {
	rac.meta.Logger.Debugf("will do runner-agent assign-service-account, %d opts", len(opts))

	_, cmdArgs, err := optparser.ParseCommandOptions(rac.meta.BinaryName+" runner-agent assign-service-account", optparser.OptionDefinitions{}, opts)
	if err != nil {
		rac.meta.Logger.Error(output.FormatError("failed to parse runner-agent assign-service-account options", err))
		return 1
	}
	if len(cmdArgs) < 2 {
		rac.meta.Logger.Error(output.FormatError("missing runner-agent assign-service-account resource paths", nil), rac.HelpRunnerAgentAssignServiceAccount())
		return 1
	}
	if len(cmdArgs) > 2 {
		msg := fmt.Sprintf("excessive runner-agent assign-service-account arguments: %s", cmdArgs)
		rac.meta.Logger.Error(output.FormatError(msg, nil), rac.HelpRunnerAgentAssignServiceAccount())
		return 1
	}

	// Validate both resource paths.
	for _, path := range cmdArgs {
		if !isResourcePathValid(rac.meta, path) {
			return 1
		}
	}

	// Prepare the inputs.
	input := &sdktypes.AssignServiceAccountToRunnerInput{
		ServiceAccountPath: cmdArgs[0],
		RunnerPath:         cmdArgs[1],
	}
	rac.meta.Logger.Debugf("runner-agent assign-service-account input: %#v", input)

	// Assign the service account to runner agent.
	if err = client.RunnerAgent.AssignServiceAccountToRunnerAgent(ctx, input); err != nil {
		rac.meta.Logger.Error(output.FormatError("failed to assign service account to runner agent", err))
		return 1
	}

	// Cannot show the assigned service account, but say something.
	rac.meta.UI.Output("service account assigned to runner agent successfully.")

	return 0
}

func (rac runnerAgentAssignServiceAccountCommand) Synopsis() string {
	return "Assign a service account to a runner agent."
}

func (rac runnerAgentAssignServiceAccountCommand) Help() string {
	return rac.HelpRunnerAgentAssignServiceAccount()
}

// HelpRunnerAgentAssignServiceAccount prints the help string for the 'runner-agent assign-service-account' command.
func (rac runnerAgentAssignServiceAccountCommand) HelpRunnerAgentAssignServiceAccount() string {
	return fmt.Sprintf(`
Usage: %s [global options] runner-agent assign-service-account <service_account_path> <runner_path>

   The runner-agent assign-service-account command assigns
   a service account to a runner agent. Service accounts
   allow a runner to interact with the Tharsis API.

`, rac.meta.BinaryName)
}
