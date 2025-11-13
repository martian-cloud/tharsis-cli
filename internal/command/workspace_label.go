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

// workspaceLabelCommand is the top-level structure for the workspace label command.
type workspaceLabelCommand struct {
	meta *Metadata
}

// labelOperation represents a single label operation (add/update or remove).
type labelOperation struct {
	key       string
	value     string
	isRemoval bool
}

// NewWorkspaceLabelCommandFactory returns a workspaceLabelCommand struct.
func NewWorkspaceLabelCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceLabelCommand{
			meta: meta,
		}, nil
	}
}

// Run executes the workspace label command.
func (wlc workspaceLabelCommand) Run(args []string) int {
	wlc.meta.Logger.Debugf("Starting the 'workspace label' command with %d arguments:", len(args))
	for ix, arg := range args {
		wlc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wlc.meta.GetSDKClient()
	if err != nil {
		wlc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wlc.doWorkspaceLabel(ctx, client, args)
}

// Synopsis returns a brief description of the workspace label command.
func (wlc workspaceLabelCommand) Synopsis() string {
	return "Manage labels on a workspace."
}

// Help returns detailed help text for the workspace label command.
func (wlc workspaceLabelCommand) Help() string {
	return wlc.HelpWorkspaceLabel()
}

// HelpWorkspaceLabel prints the help string for the 'workspace label' command.
func (wlc workspaceLabelCommand) HelpWorkspaceLabel() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace label [options] <workspace_path> [label_operations...]

   The workspace label command manages labels on a workspace.
   It supports adding, updating, removing, and overwriting labels.

   Label operations:
     key=value  Add or update a label
     key-       Remove a label (not allowed with --overwrite)

   Examples:
     # Add or update labels
     %s workspace label group/workspace env=prod tier=frontend

     # Remove a label
     %s workspace label group/workspace tier-

     # Mixed operations
     %s workspace label group/workspace env=dev tier- region=us-east-1

     # Replace all labels
     %s workspace label --overwrite group/workspace env=prod

     # Remove all labels
     %s workspace label --overwrite group/workspace

%s

`, wlc.meta.BinaryName, wlc.meta.BinaryName, wlc.meta.BinaryName, wlc.meta.BinaryName, wlc.meta.BinaryName, wlc.meta.BinaryName, buildHelpText(wlc.buildWorkspaceLabelDefs()))
}

// buildWorkspaceLabelDefs returns the option definitions for the workspace label command.
func (wlc workspaceLabelCommand) buildWorkspaceLabelDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"overwrite": {
			Arguments: []string{},
			Synopsis:  "Replace all existing labels with the specified labels.",
		},
	}
	return buildJSONOptionDefs(defs)
}

// parseLabelOperations parses label operation arguments into structured operations.
func parseLabelOperations(args []string, overwrite bool) ([]labelOperation, error) {
	operations := []labelOperation{}

	for _, arg := range args {
		// Check if this is a removal operation (ends with -)
		if len(arg) > 0 && arg[len(arg)-1] == '-' {
			// Extract the key (everything except the trailing -)
			key := arg[:len(arg)-1]
			if key == "" {
				return nil, fmt.Errorf("invalid label removal format: key cannot be empty")
			}

			// Validate that removal operations are not used with --overwrite
			if overwrite {
				return nil, fmt.Errorf("label removal syntax (key-) cannot be used with --overwrite flag")
			}

			operations = append(operations, labelOperation{
				key:       key,
				isRemoval: true,
			})
		} else if len(arg) > 0 && contains(arg, '=') {
			// This is an add/update operation (contains =)
			parts := splitOnFirst(arg, '=')
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid label format: %s (expected key=value)", arg)
			}

			key := parts[0]
			value := parts[1]

			if key == "" {
				return nil, fmt.Errorf("invalid label format: key cannot be empty in %s", arg)
			}
			if value == "" {
				return nil, fmt.Errorf("invalid label format: value cannot be empty in %s", arg)
			}

			operations = append(operations, labelOperation{
				key:       key,
				value:     value,
				isRemoval: false,
			})
		} else {
			// Invalid format - doesn't contain = or end with -
			return nil, fmt.Errorf("invalid label format: %s (expected key=value or key-)", arg)
		}
	}

	return operations, nil
}

// contains checks if a string contains a specific character.
func contains(s string, c rune) bool {
	for _, ch := range s {
		if ch == c {
			return true
		}
	}
	return false
}

// splitOnFirst splits a string on the first occurrence of a character.
func splitOnFirst(s string, c rune) []string {
	for i, ch := range s {
		if ch == c {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// applyLabelOperations applies label operations to existing labels and returns the final label set.
func applyLabelOperations(existingLabels map[string]string, operations []labelOperation, overwrite bool) ([]sdktypes.WorkspaceLabelInput, error) {
	// Initialize working map
	var workingLabels map[string]string
	if overwrite {
		// Start with empty map when overwriting
		workingLabels = make(map[string]string)
	} else {
		// Copy existing labels when not overwriting
		workingLabels = make(map[string]string, len(existingLabels))
		for k, v := range existingLabels {
			workingLabels[k] = v
		}
	}

	// Apply each operation
	for _, op := range operations {
		if op.isRemoval {
			// Remove the label key
			delete(workingLabels, op.key)
		} else {
			// Add or update the label
			workingLabels[op.key] = op.value
		}
	}

	// Convert map to slice of WorkspaceLabelInput
	result := make([]sdktypes.WorkspaceLabelInput, 0, len(workingLabels))
	for key, value := range workingLabels {
		result = append(result, sdktypes.WorkspaceLabelInput{
			Key:   key,
			Value: value,
		})
	}

	return result, nil
}

// doWorkspaceLabel contains the main logic for the workspace label command.
func (wlc workspaceLabelCommand) doWorkspaceLabel(ctx context.Context, client *tharsis.Client, opts []string) int {
	wlc.meta.Logger.Debugf("will do workspace label, %d opts", len(opts))

	// Parse command options and arguments
	defs := wlc.buildWorkspaceLabelDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wlc.meta.BinaryName+" workspace label", defs, opts)
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to parse workspace label options", err))
		return 1
	}

	// Validate exactly one workspace path argument provided
	if len(cmdArgs) < 1 {
		wlc.meta.Logger.Error(output.FormatError("missing workspace path", nil), wlc.HelpWorkspaceLabel())
		return 1
	}

	// Extract workspace path from first argument
	workspacePath := cmdArgs[0]

	// Extract remaining arguments as label operations
	labelOperationArgs := cmdArgs[1:]

	// Extract --overwrite flag
	overwrite, err := getBoolOptionValue("overwrite", "false", cmdOpts)
	if err != nil {
		wlc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract --json flag
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wlc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Validate workspace path
	// Extract path from TRN if needed
	actualPath := trn.ToPath(workspacePath)
	if !isNamespacePathValid(wlc.meta, actualPath) {
		return 1
	}

	// Parse label operations
	operations, err := parseLabelOperations(labelOperationArgs, overwrite)
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to parse label operations", err))
		return 1
	}

	// If no operations and not overwrite, display error and help text
	if len(operations) == 0 && !overwrite {
		wlc.meta.Logger.Error(output.FormatError("no label operations specified", nil), wlc.HelpWorkspaceLabel())
		return 1
	}

	// Fetch current workspace
	// Convert path to TRN
	trnID := trn.ToTRN(workspacePath, trn.ResourceTypeWorkspace)

	// Create GetWorkspaceInput with TRN ID
	getInput := &sdktypes.GetWorkspaceInput{ID: &trnID}
	wlc.meta.Logger.Debugf("workspace label get input: %#v", getInput)

	// Call client.Workspaces.GetWorkspace() to fetch workspace
	workspace, err := client.Workspaces.GetWorkspace(ctx, getInput)
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to get workspace", err))
		return 1
	}

	// Apply label operations and update workspace
	// Call applyLabelOperations() with current labels, operations, and overwrite flag
	newLabels, err := applyLabelOperations(workspace.Labels, operations, overwrite)
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to apply label operations", err))
		return 1
	}

	// Create UpdateWorkspaceInput with workspace ID, description, and new labels
	// Description must be preserved to avoid wiping it during update
	updateInput := &sdktypes.UpdateWorkspaceInput{
		ID:          &workspace.Metadata.ID,
		Description: workspace.Description,
		Labels:      newLabels,
	}
	wlc.meta.Logger.Debugf("workspace label update input: %#v", updateInput)

	// Call client.Workspaces.UpdateWorkspace() to update workspace
	updatedWorkspace, err := client.Workspaces.UpdateWorkspace(ctx, updateInput)
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to update workspace", err))
		return 1
	}

	// Output results
	return outputWorkspace(wlc.meta, toJSON, updatedWorkspace)
}
