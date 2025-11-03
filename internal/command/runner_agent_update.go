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

// runnerAgentUpdateCommand is the top-level structure for the runner-agent update command.
type runnerAgentUpdateCommand struct {
	meta *Metadata
}

// NewRunnerAgentUpdateCommandFactory returns a runnerAgentUpdateCommand struct.
func NewRunnerAgentUpdateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runnerAgentUpdateCommand{
			meta: meta,
		}, nil
	}
}

func (ruc runnerAgentUpdateCommand) Run(args []string) int {
	ruc.meta.Logger.Debugf("Starting the 'runner-agent update' command with %d arguments:", len(args))
	for ix, arg := range args {
		ruc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := ruc.meta.GetSDKClient()
	if err != nil {
		ruc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return ruc.doRunnerAgentUpdate(ctx, client, args)
}

func (ruc runnerAgentUpdateCommand) doRunnerAgentUpdate(ctx context.Context, client *tharsis.Client, opts []string) int {
	ruc.meta.Logger.Debugf("will do runner-agent update, %d opts", len(opts))

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ruc.meta.BinaryName+" runner-agent update", ruc.buildRunnerAgentUpdateDefs(), opts)
	if err != nil {
		ruc.meta.Logger.Error(output.FormatError("failed to parse runner-agent update options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ruc.meta.Logger.Error(output.FormatError("missing runner-agent update id", nil), ruc.HelpRunnerAgentUpdate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive runner-agent update arguments: %s", cmdArgs)
		ruc.meta.Logger.Error(output.FormatError(msg, nil), ruc.HelpRunnerAgentUpdate())
		return 1
	}

	description := getOption("description", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		ruc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	runUntaggedJobs, err := getBoolOptionValue("run-untagged-jobs", "false", cmdOpts)
	if err != nil {
		ruc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	tags := getOptionSlice("tag", cmdOpts)

	// Prepare the inputs.
	input := &sdktypes.UpdateRunnerInput{
		ID:              cmdArgs[0],
		Description:     description,
		RunUntaggedJobs: &runUntaggedJobs,
		Tags:            &tags,
	}

	ruc.meta.Logger.Debugf("runner-agent update input: %#v", input)

	updatedRunnerAgent, err := client.RunnerAgent.UpdateRunnerAgent(ctx, input)
	if err != nil {
		ruc.meta.Logger.Error(output.FormatError("failed to update runner-agent", err))
		return 1
	}

	return outputRunnerAgent(ruc.meta, toJSON, updatedRunnerAgent)
}

func (ruc runnerAgentUpdateCommand) buildRunnerAgentUpdateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  "New description for the runner agent.",
		},
		"run-untagged-jobs": {
			Arguments: []string{},
			Synopsis:  "Run untagged jobs.",
		},
		"tag": {
			Arguments: []string{"Runner_Tag"},
			Synopsis:  "Runner tag to assign to the runner agent. (This flag may be repeated).",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (ruc runnerAgentUpdateCommand) Synopsis() string {
	return "Update a runner agent."
}

func (ruc runnerAgentUpdateCommand) Help() string {
	return ruc.HelpRunnerAgentUpdate()
}

// HelpRunnerAgentUpdate prints the help string for the 'runner-agent update' command.
func (ruc runnerAgentUpdateCommand) HelpRunnerAgentUpdate() string {
	return fmt.Sprintf(`
Usage: %s [global options] runner-agent update [options] <id>

   The runner-agent update command updates a runner agent.
   Shows final output as JSON, if specified.

%s

`, ruc.meta.BinaryName, buildHelpText(ruc.buildRunnerAgentUpdateDefs()))
}
