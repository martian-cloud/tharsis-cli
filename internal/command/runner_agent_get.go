package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// runnerAgentGetCommand is the top-level structure for the runner-agent get command.
type runnerAgentGetCommand struct {
	meta *Metadata
}

// NewRunnerAgentGetCommandFactory returns a runnerAgentGetCommand struct.
func NewRunnerAgentGetCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runnerAgentGetCommand{
			meta: meta,
		}, nil
	}
}

func (rgc runnerAgentGetCommand) Run(args []string) int {
	rgc.meta.Logger.Debugf("Starting the 'runner-agent get' command with %d arguments:", len(args))
	for ix, arg := range args {
		rgc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := rgc.meta.GetSDKClient()
	if err != nil {
		rgc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return rgc.doRunnerAgentGet(ctx, client, args)
}

func (rgc runnerAgentGetCommand) doRunnerAgentGet(ctx context.Context, client *tharsis.Client, opts []string) int {
	rgc.meta.Logger.Debugf("will do runner-agent get, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(rgc.meta.BinaryName+" runner-agent get", defs, opts)
	if err != nil {
		rgc.meta.Logger.Error(output.FormatError("failed to parse runner-agent get argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		rgc.meta.Logger.Error(output.FormatError("missing runner-agent get id", nil), rgc.HelpRunnerAgentGet())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive runner-agent get arguments: %s", cmdArgs)
		rgc.meta.Logger.Error(output.FormatError(msg, nil), rgc.HelpRunnerAgentGet())
		return 1
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		rgc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.GetRunnerInput{ID: cmdArgs[0]}
	rgc.meta.Logger.Debugf("runner-agent get input: %#v", input)

	// Get the runner agent.
	foundRunnerAgent, err := client.RunnerAgent.GetRunnerAgent(ctx, input)
	if err != nil {
		rgc.meta.Logger.Error(output.FormatError("failed to get runner agent", err))
		return 1
	}

	return outputRunnerAgent(rgc.meta, toJSON, foundRunnerAgent)
}

// outputRunnerAgent is the final output for most runner-agent operations.
func outputRunnerAgent(meta *Metadata, toJSON bool, runnerAgent *sdktypes.RunnerAgent) int {
	if toJSON {
		buf, err := objectToJSON(runnerAgent)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{
				"id",
				"name",
				"description",
				"resource path",
				"type",
				"run untagged jobs",
				"tags",
			},
			{
				runnerAgent.Metadata.ID,
				runnerAgent.Name,
				runnerAgent.Description,
				runnerAgent.ResourcePath,
				string(runnerAgent.Type),
				strconv.FormatBool(runnerAgent.RunUntaggedJobs),
				strings.Join(runnerAgent.Tags, ", "),
			},
		}
		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

func (rgc runnerAgentGetCommand) Synopsis() string {
	return "Get a single runner agent."
}

func (rgc runnerAgentGetCommand) Help() string {
	return rgc.HelpRunnerAgentGet()
}

// HelpRunnerAgentGet prints the help string for the 'runner-agent get' command.
func (rgc runnerAgentGetCommand) HelpRunnerAgentGet() string {
	return fmt.Sprintf(`
Usage: %s [global options] runner-agent get [options] <id>

   The runner-agent get command prints information about
   one runner agent.

%s

`, rgc.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
