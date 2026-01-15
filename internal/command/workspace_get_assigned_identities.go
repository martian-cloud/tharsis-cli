package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspaceGetAssignedManagedIdentitiesCommand is the top-level structure for
// retrieving assigned managed identities.
type workspaceGetAssignedManagedIdentitiesCommand struct {
	meta *Metadata
}

// NewWorkspaceGetAssignedManagedIdentitiesCommandFactory returns a new workspaceGetAssignedManagedIdentitiesCommand.
func NewWorkspaceGetAssignedManagedIdentitiesCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceGetAssignedManagedIdentitiesCommand{
			meta: meta,
		}, nil
	}
}

func (wam workspaceGetAssignedManagedIdentitiesCommand) Run(args []string) int {
	wam.meta.Logger.Debugf("Starting the 'workspace get-assigned-managed-identities' command with %d arguments:", len(args))
	for ix, arg := range args {
		wam.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wam.meta.GetSDKClient()
	if err != nil {
		wam.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wam.doWorkspaceGetAssignedManagedIdentities(ctx, client, args)
}

func (wam workspaceGetAssignedManagedIdentitiesCommand) doWorkspaceGetAssignedManagedIdentities(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	wam.meta.Logger.Debugf("will do workspace get-assigned-managed-identities, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wam.meta.BinaryName+" workspace get-assigned-managed-identities", defs, opts)
	if err != nil {
		wam.meta.Logger.Error(output.FormatError("failed to parse workspace get-assigned-managed-identities argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wam.meta.Logger.Error(output.FormatError("missing workspace get-assigned-managed-identities full path", nil), wam.HelpWorkspaceGetAssignedManagedIdentities())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace get-assigned-managed-identities arguments: %s", cmdArgs)
		wam.meta.Logger.Error(output.FormatError(msg, nil), wam.HelpWorkspaceGetAssignedManagedIdentities())
		return 1
	}

	workspacePath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wam.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(workspacePath)
	if !isNamespacePathValid(wam.meta, actualPath) {
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.GetAssignedManagedIdentitiesInput{Path: &actualPath} // Use extracted path
	wam.meta.Logger.Debugf("workspace get-assigned-managed-identities input: %#v", input)

	// Get the managed identities.
	identities, err := client.Workspaces.GetAssignedManagedIdentities(ctx, input)
	if err != nil {
		wam.meta.Logger.Error(output.FormatError("failed to get assigned managed identities", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(identities)
		if err != nil {
			wam.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		wam.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(identities)+1)
		tableInput[0] = []string{"name", "resourcePath", "description", "id"}
		for ix, identity := range identities {
			tableInput[ix+1] = []string{
				identity.Name, identity.ResourcePath,
				identity.Description, identity.Metadata.ID,
			}
		}
		wam.meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

func (wam workspaceGetAssignedManagedIdentitiesCommand) Synopsis() string {
	return "Get assigned managed identities for a workspace."
}

func (wam workspaceGetAssignedManagedIdentitiesCommand) Help() string {
	return wam.HelpWorkspaceGetAssignedManagedIdentities()
}

// HelpWorkspaceGetAssignedManagedIdentities prints the help string for the
// 'workspace get-assigned-managed-identities' command.
func (wam workspaceGetAssignedManagedIdentitiesCommand) HelpWorkspaceGetAssignedManagedIdentities() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace get-assigned-managed-identities [options] <full_path>

   The workspace get-assigned-managed-identities command
   prints information about managed identities assigned
   to a workspace. Shows final output as JSON, if
   specified.

%s

`, wam.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
