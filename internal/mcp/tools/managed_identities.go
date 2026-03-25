package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// assignManagedIdentityInput is the input for assigning a managed identity.
type assignManagedIdentityInput struct {
	WorkspaceID       string `json:"workspace_id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace-name)"`
	ManagedIdentityID string `json:"managed_identity_id" jsonschema:"required,Managed identity ID or TRN (e.g. Ul8yZ... or trn:managed_identity:group/identity-name)"`
}

// assignManagedIdentityOutput is the output for assigning a managed identity.
type assignManagedIdentityOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
}

// AssignManagedIdentity returns an MCP tool for assigning a managed identity to a workspace.
func assignManagedIdentity(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*assignManagedIdentityInput, *assignManagedIdentityOutput]) {
	tool := mcp.Tool{
		Name:        "assign_managed_identity",
		Description: "Assign a managed identity to a workspace to grant it access to cloud resources.",
		Annotations: &mcp.ToolAnnotations{
			Title: "Assign Managed Identity",
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *assignManagedIdentityInput) (*mcp.CallToolResult, *assignManagedIdentityOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.WorkspaceID, trn.ResourceTypeWorkspace); err != nil {
			return nil, nil, err
		}

		if _, err := tc.grpcClient.ManagedIdentitiesClient.AssignManagedIdentityToWorkspace(ctx, &pb.AssignManagedIdentityToWorkspaceRequest{
			ManagedIdentityId: input.ManagedIdentityID,
			WorkspaceId:       input.WorkspaceID,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to assign managed identity: %w", err)
		}

		return nil, &assignManagedIdentityOutput{
			Message: fmt.Sprintf("Managed identity %s assigned to workspace %s", input.ManagedIdentityID, input.WorkspaceID),
			Success: true,
		}, nil
	}

	return tool, handler
}

// unassignManagedIdentityInput is the input for unassigning a managed identity.
type unassignManagedIdentityInput struct {
	WorkspaceID       string `json:"workspace_id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace-name)"`
	ManagedIdentityID string `json:"managed_identity_id" jsonschema:"required,Managed identity ID or TRN (e.g. Ul8yZ... or trn:managed_identity:group/identity-name)"`
}

// unassignManagedIdentityOutput is the output for unassigning a managed identity.
type unassignManagedIdentityOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
}

// UnassignManagedIdentity returns an MCP tool for unassigning a managed identity from a workspace.
func unassignManagedIdentity(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*unassignManagedIdentityInput, *unassignManagedIdentityOutput]) {
	tool := mcp.Tool{
		Name:        "unassign_managed_identity",
		Description: "Remove a managed identity assignment from a workspace.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Unassign Managed Identity",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *unassignManagedIdentityInput) (*mcp.CallToolResult, *unassignManagedIdentityOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.WorkspaceID, trn.ResourceTypeWorkspace); err != nil {
			return nil, nil, err
		}

		if _, err := tc.grpcClient.ManagedIdentitiesClient.RemoveManagedIdentityFromWorkspace(ctx, &pb.RemoveManagedIdentityFromWorkspaceRequest{
			ManagedIdentityId: input.ManagedIdentityID,
			WorkspaceId:       input.WorkspaceID,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to unassign managed identity: %w", err)
		}

		return nil, &unassignManagedIdentityOutput{
			Message: fmt.Sprintf("Managed identity %s unassigned from workspace %s", input.ManagedIdentityID, input.WorkspaceID),
			Success: true,
		}, nil
	}

	return tool, handler
}
