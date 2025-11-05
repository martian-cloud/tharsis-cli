package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspaceUpdateCommand is the top-level structure for the workspace update command.
type workspaceUpdateCommand struct {
	meta *Metadata
}

// NewWorkspaceUpdateCommandFactory returns a workspaceUpdateCommand struct.
func NewWorkspaceUpdateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceUpdateCommand{
			meta: meta,
		}, nil
	}
}

func (wuc workspaceUpdateCommand) Run(args []string) int {
	wuc.meta.Logger.Debugf("Starting the 'workspace update' command with %d arguments:", len(args))
	for ix, arg := range args {
		wuc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wuc.meta.GetSDKClient()
	if err != nil {
		wuc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wuc.doWorkspaceUpdate(ctx, client, args)
}

func (wuc workspaceUpdateCommand) doWorkspaceUpdate(ctx context.Context, client *tharsis.Client, opts []string) int {
	wuc.meta.Logger.Debugf("will do workspace update, %d opts", len(opts))

	defs := buildCommonUpdateOptionDefs("workspace")
	buildCommonWorkspaceDefs(defs)
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wuc.meta.BinaryName+" workspace update", defs, opts)
	if err != nil {
		wuc.meta.Logger.Error(output.FormatError("failed to parse workspace update options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wuc.meta.Logger.Error(output.FormatError("missing workspace update full path", nil), wuc.HelpWorkspaceUpdate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace update arguments: %s", cmdArgs)
		wuc.meta.Logger.Error(output.FormatError(msg, nil), wuc.HelpWorkspaceUpdate())
		return 1
	}

	path := cmdArgs[0]
	description := getOption("description", "", cmdOpts)[0]
	maxJobDuration := getOption("max-job-duration", "720", cmdOpts)[0]
	terraformVersion := getOption("terraform-version", "", cmdOpts)[0]
	preventDestroyPlan, err := getBoolOptionValue("prevent-destroy-plan", "false", cmdOpts)
	if err != nil {
		wuc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wuc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(path)
	if !isNamespacePathValid(wuc.meta, actualPath) {
		return 1
	}

	// Convert maxJobDuration to an int.
	var jobDuration *int32
	duration, pErr := parseMaximumJobDuration(maxJobDuration)
	if pErr != nil {
		wuc.meta.Logger.Error(output.FormatError("failed to parse max job duration", pErr))
		return 1
	}
	jobDuration = &duration

	input := &sdktypes.UpdateWorkspaceInput{
		Description:        description,
		MaxJobDuration:     jobDuration,
		PreventDestroyPlan: &preventDestroyPlan,
	}

	// Convert path to TRN and use ID field
	trnID := trn.ToTRN(path, trn.ResourceTypeWorkspace)
	input.ID = &trnID

	if terraformVersion != "" {
		input.TerraformVersion = &terraformVersion
	}

	wuc.meta.Logger.Debugf("workspace update input: %#v", input)

	updatedWorkspace, err := client.Workspaces.UpdateWorkspace(ctx, input)
	if err != nil {
		wuc.meta.Logger.Error(output.FormatError("failed to update a workspace", err))
		return 1
	}

	return outputWorkspace(wuc.meta, toJSON, updatedWorkspace)
}

// buildCommonUpdateOptionDefs returns the common defs used by
// workspace and group update commands.
func buildCommonUpdateOptionDefs(synopsis string) optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  fmt.Sprintf("New description for the %s.", synopsis),
		},
		"terraform-version": {
			Arguments: []string{"Terraform_Version"},
			Synopsis:  fmt.Sprintf("The default Terraform CLI version for the new %s.", synopsis),
		},
	}

	return buildJSONOptionDefs(defs)
}

// buildCommonWorkspaceDefs contains common defs used by multiple workspace commands.
func buildCommonWorkspaceDefs(defs optparser.OptionDefinitions) {
	maxJobDef := optparser.OptionDefinition{
		Arguments: []string{"Max_Job_Duration"},
		Synopsis:  "The amount of minutes before a job is gracefully canceled (Default 720).",
	}

	defs["max-job-duration"] = &maxJobDef

	// The --prevent-destroy-plan option should be available only for workspace create and update.
	preventDestroyPlanDef := optparser.OptionDefinition{
		Arguments: []string{},
		Synopsis:  "boolean value--whether a run/plan will be prevented from destroying deployed resources.",
	}
	defs["prevent-destroy-plan"] = &preventDestroyPlanDef
}

func (wuc workspaceUpdateCommand) Synopsis() string {
	return "Update a workspace."
}

func (wuc workspaceUpdateCommand) Help() string {
	return wuc.HelpWorkspaceUpdate()
}

// HelpWorkspaceUpdate produces the help string for the 'workspace update' command.
func (wuc workspaceUpdateCommand) HelpWorkspaceUpdate() string {
	defs := buildCommonUpdateOptionDefs("workspace")
	buildCommonWorkspaceDefs(defs)

	return fmt.Sprintf(`
Usage: %s [global options] workspace update [options] <full_path>

   The workspace update command updates a workspace.
   Currently, it supports updating the description and the
   maximum job duration. Shows final output as JSON, if
   specified.

%s

`, wuc.meta.BinaryName, buildHelpText(defs))
}
