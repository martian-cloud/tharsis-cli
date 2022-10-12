package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspaceCreateCommand is the top-level structure for the workspace create command.
type workspaceCreateCommand struct {
	meta *Metadata
}

// NewWorkspaceCreateCommandFactory returns a workspaceCreateCommand struct.
func NewWorkspaceCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceCreateCommand{
			meta: meta,
		}, nil
	}
}

func (wcc workspaceCreateCommand) Run(args []string) int {
	wcc.meta.Logger.Debugf("Starting the 'workspace create' command with %d arguments:", len(args))
	for ix, arg := range args {
		wcc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := wcc.meta.ReadSettings()
	if err != nil {
		wcc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		wcc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wcc.doWorkspaceCreate(ctx, client, args)
}

func (wcc workspaceCreateCommand) doWorkspaceCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	wcc.meta.Logger.Debugf("will do workspace create, %d opts", len(opts))

	defs := buildWorkspaceCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wcc.meta.BinaryName+" workspace create", defs, opts)
	if err != nil {
		wcc.meta.Logger.Error(output.FormatError("failed to parse workspace create options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wcc.meta.Logger.Error(output.FormatError("missing workspace create full path", nil), wcc.HelpWorkspaceCreate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace create arguments: %s", cmdArgs)
		wcc.meta.Logger.Error(output.FormatError(msg, nil), wcc.HelpWorkspaceCreate())
		return 1
	}

	workspacePath := cmdArgs[0]
	description := getOption("description", "", cmdOpts)[0]
	ifNotExists := getOption("if-not-exists", "", cmdOpts)[0] == "1"
	terraformVersion := getOption("terraform-version", "", cmdOpts)[0]
	identityPaths := getOption("managed-identity", "", cmdOpts)
	maxJobDuration := getOption("max-job-duration", "", cmdOpts)[0]
	toJSON := getOption("json", "", cmdOpts)[0] == "1"

	// Error is already logged.
	if !isNamespacePathValid(wcc.meta, workspacePath) {
		return 1
	}

	// Validate managed identity paths.
	for _, identity := range identityPaths {
		if !isResourcePathValid(wcc.meta, identity) {
			return 1
		}
	}

	// Convert maxJobDuration to an int.
	var jobDuration *int32
	if maxJobDuration != "" {
		duration, pErr := parseMaximumJobDuration(maxJobDuration)
		if pErr != nil {
			wcc.meta.Logger.Error(output.FormatError("failed to parse max job duration", pErr))
			return 1
		}
		jobDuration = &duration
	}

	// Check if workspace already exists.
	if ifNotExists {
		ws, gErr := client.Workspaces.GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{Path: workspacePath})
		if gErr != nil {
			wcc.meta.Logger.Error(output.FormatError("failed to get workspace", gErr))
			return 1
		}

		if ws != nil {
			return outputWorkspace(wcc.meta, toJSON, ws, "create")
		}
	}

	// Prepare the inputs. Output an error or slice out of bounds in input preparation.
	index := strings.LastIndex(workspacePath, sep)
	if index == -1 {
		wcc.meta.Logger.Error(output.FormatError("workspace path is invalid", nil))
		return 1
	}

	input := &sdktypes.CreateWorkspaceInput{
		Name:           workspacePath[index+1:],
		GroupPath:      workspacePath[:index],
		Description:    description,
		MaxJobDuration: jobDuration,
	}

	if terraformVersion != "" {
		input.TerraformVersion = &terraformVersion
	}

	wcc.meta.Logger.Debugf("workspace create input: %#v", input)

	// Create the workspace.
	createdWorkspace, err := client.Workspaces.CreateWorkspace(ctx, input)
	if err != nil {
		wcc.meta.Logger.Error(output.FormatError("failed to create a workspace", err))
		return 1
	}
	if createdWorkspace == nil {
		wcc.meta.Logger.Error(output.FormatError("failed to create a workspace", nil))
		return 1
	}

	if len(identityPaths) > 0 {
		createdWorkspace, err = assignManagedIdentities(ctx, workspacePath, identityPaths, client)
		if err != nil {
			wcc.meta.Logger.Error(output.FormatError("failed to assign managed identity to workspace", err))
			return 1
		}

		if createdWorkspace == nil {
			wcc.meta.Logger.Error(output.FormatError("failed to assign managed identity to workspace", nil))
			return 1
		}
	}

	return outputWorkspace(wcc.meta, toJSON, createdWorkspace, "create")
}

// outputWorkspace is the final output for most workspace operations.
func outputWorkspace(meta *Metadata, toJSON bool, workspace *sdktypes.Workspace, action string) int {
	if toJSON {
		buf, err := objectToJSON(workspace)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		// Format the output.
		meta.UI.Output(fmt.Sprintf("workspace %s output:", action))
		meta.UI.Output(fmt.Sprintf("\n             name: %s", workspace.Name))
		meta.UI.Output(fmt.Sprintf("         fullpath: %s", workspace.FullPath))
		meta.UI.Output(fmt.Sprintf("      description: %s", workspace.Description))
		meta.UI.Output(fmt.Sprintf(" max job duration: %dm", workspace.MaxJobDuration))
		meta.UI.Output(fmt.Sprintf("terraform version: %s", workspace.TerraformVersion))
		meta.UI.Output(fmt.Sprintf("               ID: %s", workspace.Metadata.ID))
	}

	return 0
}

// parseMaximumJobDuration parses the maxJobDuration and returns an int32.
func parseMaximumJobDuration(maxJobDuration string) (int32, error) {
	value, err := strconv.ParseInt(maxJobDuration, 10, 64)
	if err != nil {
		return 0, err
	}

	return int32(value), nil
}

// buildCommonCreateOptionDefs returns the common defs used by
// workspace and group create commands.
func buildCommonCreateOptionDefs(synopsis string) optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  fmt.Sprintf("Description for the new %s.", synopsis),
		},
		"if-not-exists": {
			Arguments: []string{},
			Synopsis:  fmt.Sprintf("Create a %s if it does not already exist.", synopsis),
		},
		"terraform-version": {
			Arguments: []string{"Terraform_Version"},
			Synopsis:  fmt.Sprintf("The default Terraform CLI version for the new %s.", synopsis),
		},
	}

	return buildJSONOptionDefs(defs)
}

// buildWorkspaceCreateDefs returns defs used by workspace create command.
func buildWorkspaceCreateDefs() optparser.OptionDefinitions {
	defs := buildCommonCreateOptionDefs("workspace")
	identityDef := optparser.OptionDefinition{
		Arguments: []string{"Managed_Identity"},
		Synopsis:  "The full resource path to a managed identity.",
	}
	defs["managed-identity"] = &identityDef

	// Get common defs used by multiple workspace commands.
	buildCommonWorkspaceDefs(defs)

	return defs
}

func (wcc workspaceCreateCommand) Synopsis() string {
	return "Create a new workspace."
}

func (wcc workspaceCreateCommand) Help() string {
	return wcc.HelpWorkspaceCreate()
}

// HelpWorkspaceCreate produces the help string for the 'workspace create' command.
func (wcc workspaceCreateCommand) HelpWorkspaceCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace create [options] <full_path>

   The workspace create command creates a new workspace. It
   allows setting a workspace's description (optional),
   maximum job duration and managed identity. Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.

%s

`, wcc.meta.BinaryName, buildHelpText(buildWorkspaceCreateDefs()))
}

// The End.
