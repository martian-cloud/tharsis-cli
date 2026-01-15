package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// runnerAgentUnassignServiceAccountCommand is the top-level structure for the runner-agent unassign-service-account command.
type runnerAgentUnassignServiceAccountCommand struct {
	meta *Metadata
}

// NewRunnerAgentUnassignServiceAccountCommandFactory returns a runnerAgentUnassignServiceAccountCommand struct.
func NewRunnerAgentUnassignServiceAccountCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runnerAgentUnassignServiceAccountCommand{
			meta: meta,
		}, nil
	}
}

func (rac runnerAgentUnassignServiceAccountCommand) Run(args []string) int {
	rac.meta.Logger.Debugf("Starting the 'runner-agent unassign-service-account' command with %d arguments:", len(args))
	for ix, arg := range args {
		rac.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := rac.meta.GetSDKClient()
	if err != nil {
		rac.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return rac.doRunnerAgentUnassignServiceAccount(ctx, client, args)
}

func (rac runnerAgentUnassignServiceAccountCommand) doRunnerAgentUnassignServiceAccount(ctx context.Context, client *tharsis.Client, opts []string) int {
	rac.meta.Logger.Debugf("will do runner-agent unassign-service-account, %d opts", len(opts))

	_, cmdArgs, err := optparser.ParseCommandOptions(rac.meta.BinaryName+" runner-agent unassign-service-account", optparser.OptionDefinitions{}, opts)
	if err != nil {
		rac.meta.Logger.Error(output.FormatError("failed to parse runner-agent unassign-service-account options", err))
		return 1
	}
	if len(cmdArgs) < 2 {
		rac.meta.Logger.Error(output.FormatError("missing runner-agent unassign-service-account resource paths", nil), rac.HelpRunnerAgentUnassignServiceAccount())
		return 1
	}
	if len(cmdArgs) > 2 {
		msg := fmt.Sprintf("excessive runner-agent unassign-service-account arguments: %s", cmdArgs)
		rac.meta.Logger.Error(output.FormatError(msg, nil), rac.HelpRunnerAgentUnassignServiceAccount())
		return 1
	}

	// Validate both resource paths.
	for _, path := range cmdArgs {
		actualPath := trn.ToPath(path)
		if !isResourcePathValid(rac.meta, actualPath) {
			return 1
		}
	}

	// Prepare the inputs.
	input := &sdktypes.AssignServiceAccountToRunnerInput{
		ServiceAccountPath: cmdArgs[0],
		RunnerPath:         cmdArgs[1],
	}
	rac.meta.Logger.Debugf("runner-agent unassign-service-account input: %#v", input)

	// Unassign the service account from runner agent.
	if err = client.RunnerAgent.UnassignServiceAccountFromRunnerAgent(ctx, input); err != nil {
		rac.meta.Logger.Error(output.FormatError("failed to unassign service account from runner agent", err))
		return 1
	}

	// Cannot show the unassigned service account, but say something.
	rac.meta.UI.Output("service account unassigned from runner agent successfully.")

	return 0
}

func (rac runnerAgentUnassignServiceAccountCommand) Synopsis() string {
	return "Unassign a service account to a runner agent."
}

func (rac runnerAgentUnassignServiceAccountCommand) Help() string {
	return rac.HelpRunnerAgentUnassignServiceAccount()
}

// HelpRunnerAgentUnassignServiceAccount prints the help string for the 'runner-agent unassign-service-account' command.
func (rac runnerAgentUnassignServiceAccountCommand) HelpRunnerAgentUnassignServiceAccount() string {
	return fmt.Sprintf(`
Usage: %s [global options] runner-agent unassign-service-account <service_account_path> <runner_path>

   The runner-agent unassign-service-account command removes
   a service account from a runner agent. Service accounts
   allow a runner to interact with the Tharsis API.

`, rac.meta.BinaryName)
}
