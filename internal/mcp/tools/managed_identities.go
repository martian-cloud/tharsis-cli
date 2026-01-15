package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// assignManagedIdentityInput is the input for assigning a managed identity.
type assignManagedIdentityInput struct {
	WorkspacePath     string `json:"workspace_path" jsonschema:"required,Full path to the workspace (e.g. group/subgroup/workspace-name)"`
	ManagedIdentityID string `json:"managed_identity_id" jsonschema:"required,Managed identity ID or TRN (e.g. Ul8yZ... or trn:managed_identity:group/identity-name)"`
}

// assignManagedIdentityOutput is the output for assigning a managed identity.
type assignManagedIdentityOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
}

// AssignManagedIdentity returns an MCP tool for assigning a managed identity to a workspace.
func assignManagedIdentity(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[assignManagedIdentityInput, assignManagedIdentityOutput]) {
	tool := mcp.Tool{
		Name:        "assign_managed_identity",
		Description: "Assign a managed identity to a workspace to grant it access to cloud resources.",
		Annotations: &mcp.ToolAnnotations{
			Title: "Assign Managed Identity",
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input assignManagedIdentityInput) (*mcp.CallToolResult, assignManagedIdentityOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, assignManagedIdentityOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, trn.ToTRN(input.WorkspacePath, trn.ResourceTypeWorkspace), trn.ResourceTypeWorkspace); err != nil {
			return nil, assignManagedIdentityOutput{}, err
		}

		// Assign the managed identity
		assignInput := &sdktypes.AssignManagedIdentityInput{
			ManagedIdentityID: &input.ManagedIdentityID,
			WorkspacePath:     input.WorkspacePath,
		}

		if _, err := client.ManagedIdentities().AssignManagedIdentityToWorkspace(ctx, assignInput); err != nil {
			return nil, assignManagedIdentityOutput{}, fmt.Errorf("failed to assign managed identity: %w", err)
		}

		return nil, assignManagedIdentityOutput{
			Message: fmt.Sprintf("Managed identity %s assigned to workspace %s", input.ManagedIdentityID, input.WorkspacePath),
			Success: true,
		}, nil
	}

	return tool, handler
}

// unassignManagedIdentityInput is the input for unassigning a managed identity.
type unassignManagedIdentityInput struct {
	WorkspacePath     string `json:"workspace_path" jsonschema:"required,Full path to the workspace (e.g. group/subgroup/workspace-name)"`
	ManagedIdentityID string `json:"managed_identity_id" jsonschema:"required,Managed identity ID or TRN (e.g. Ul8yZ... or trn:managed_identity:group/identity-name)"`
}

// unassignManagedIdentityOutput is the output for unassigning a managed identity.
type unassignManagedIdentityOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
}

// UnassignManagedIdentity returns an MCP tool for unassigning a managed identity from a workspace.
func unassignManagedIdentity(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[unassignManagedIdentityInput, unassignManagedIdentityOutput]) {
	tool := mcp.Tool{
		Name:        "unassign_managed_identity",
		Description: "Remove a managed identity assignment from a workspace.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Unassign Managed Identity",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input unassignManagedIdentityInput) (*mcp.CallToolResult, unassignManagedIdentityOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, unassignManagedIdentityOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, trn.ToTRN(input.WorkspacePath, trn.ResourceTypeWorkspace), trn.ResourceTypeWorkspace); err != nil {
			return nil, unassignManagedIdentityOutput{}, err
		}

		// Unassign the managed identity
		unassignInput := &sdktypes.AssignManagedIdentityInput{
			ManagedIdentityID: &input.ManagedIdentityID,
			WorkspacePath:     input.WorkspacePath,
		}

		if _, err := client.ManagedIdentities().UnassignManagedIdentityFromWorkspace(ctx, unassignInput); err != nil {
			return nil, unassignManagedIdentityOutput{}, fmt.Errorf("failed to unassign managed identity: %w", err)
		}

		return nil, unassignManagedIdentityOutput{
			Message: fmt.Sprintf("Managed identity %s unassigned from workspace %s", input.ManagedIdentityID, input.WorkspacePath),
			Success: true,
		}, nil
	}

	return tool, handler
}
