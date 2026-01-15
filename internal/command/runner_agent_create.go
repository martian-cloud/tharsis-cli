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

// runnerAgentCreateCommand is the top-level structure for the runner-agent create command.
type runnerAgentCreateCommand struct {
	meta *Metadata
}

// NewRunnerAgentCreateCommandFactory returns a runnerAgentCreateCommand struct.
func NewRunnerAgentCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runnerAgentCreateCommand{
			meta: meta,
		}, nil
	}
}

func (rcc runnerAgentCreateCommand) Run(args []string) int {
	rcc.meta.Logger.Debugf("Starting the 'runner-agent create' command with %d arguments:", len(args))
	for ix, arg := range args {
		rcc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := rcc.meta.GetSDKClient()
	if err != nil {
		rcc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return rcc.doRunnerAgentCreate(ctx, client, args)
}

func (rcc runnerAgentCreateCommand) doRunnerAgentCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	rcc.meta.Logger.Debugf("will do runner-agent create, %d opts", len(opts))

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(rcc.meta.BinaryName+" runner-agent create", rcc.buildRunnerAgentCreateDefs(), opts)
	if err != nil {
		rcc.meta.Logger.Error(output.FormatError("failed to parse runner-agent create options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive runner-agent create arguments: %s", cmdArgs)
		rcc.meta.Logger.Error(output.FormatError(msg, nil), rcc.HelpRunnerAgentCreate())
		return 1
	}

	runnerAgentName := getOption("runner-name", "", cmdOpts)[0]
	groupPath := getOption("group-path", "", cmdOpts)[0]
	description := getOption("description", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		rcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	runUntaggedJobs, err := getBoolOptionValue("run-untagged-jobs", "false", cmdOpts)
	if err != nil {
		rcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	tags := getOptionSlice("tag", cmdOpts)

	actualPath := trn.ToPath(groupPath)
	if !isNamespacePathValid(rcc.meta, actualPath) {
		return 1
	}

	input := &sdktypes.CreateRunnerInput{
		Name:            runnerAgentName,
		GroupPath:       groupPath,
		Description:     description,
		RunUntaggedJobs: runUntaggedJobs,
		Tags:            tags,
	}

	rcc.meta.Logger.Debugf("runner-agent create input: %#v", input)

	// Create the runner agent.
	createdRunnerAgent, err := client.RunnerAgent.CreateRunnerAgent(ctx, input)
	if err != nil {
		rcc.meta.Logger.Error(output.FormatError("failed to create runner agent", err))
		return 1
	}

	return outputRunnerAgent(rcc.meta, toJSON, createdRunnerAgent)
}

func (rcc runnerAgentCreateCommand) buildRunnerAgentCreateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  "Description for the runner agent.",
		},
		"group-path": {
			Arguments: []string{"Group_Path"},
			Synopsis:  "Full path of group where runner will be created.",
			Required:  true,
		},
		"runner-name": {
			Arguments: []string{"Runner_Name"},
			Synopsis:  "Name of the new runner agent.",
			Required:  true,
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

func (rcc runnerAgentCreateCommand) Synopsis() string {
	return "Create a new runner agent."
}

func (rcc runnerAgentCreateCommand) Help() string {
	return rcc.HelpRunnerAgentCreate()
}

// HelpRunnerAgentCreate prints the help string for the 'runner-agent create' command.
func (rcc runnerAgentCreateCommand) HelpRunnerAgentCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] runner-agent create [options]

   The runner-agent create command creates a new runner agent.
   Shows final output as JSON, if specified.

%s

`, rcc.meta.BinaryName, buildHelpText(rcc.buildRunnerAgentCreateDefs()))
}
