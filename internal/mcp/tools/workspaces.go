package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspace represents a Tharsis workspace in MCP responses.
type workspace struct {
	ID                 string            `json:"id" jsonschema:"The unique identifier of the workspace"`
	Path               string            `json:"path" jsonschema:"Full path to the workspace"`
	Name               string            `json:"name" jsonschema:"Workspace name"`
	GroupPath          string            `json:"group_path" jsonschema:"Path to the parent group"`
	Description        string            `json:"description" jsonschema:"Workspace description"`
	TerraformVersion   string            `json:"terraform_version" jsonschema:"Terraform CLI version"`
	MaxJobDuration     int32             `json:"max_job_duration" jsonschema:"Maximum number of minutes a Terraform job is allowed to run"`
	PreventDestroyPlan bool              `json:"prevent_destroy_plan" jsonschema:"Whether destroy plans are prevented"`
	Labels             map[string]string `json:"labels,omitempty" jsonschema:"Workspace labels"`
	TRN                string            `json:"trn" jsonschema:"Tharsis Resource Name"`
}

// toWorkspace converts an SDK workspace to MCP workspace.
func toWorkspace(ws *sdktypes.Workspace) workspace {
	return workspace{
		ID:                 ws.Metadata.ID,
		TRN:                ws.Metadata.TRN,
		Path:               ws.FullPath,
		Name:               ws.Name,
		GroupPath:          ws.GroupPath,
		Description:        ws.Description,
		TerraformVersion:   ws.TerraformVersion,
		MaxJobDuration:     ws.MaxJobDuration,
		PreventDestroyPlan: ws.PreventDestroyPlan,
		Labels:             ws.Labels,
	}
}

// listWorkspacesInput is the input for listing workspaces.
type listWorkspacesInput struct {
	GroupPath *string                          `json:"group_path,omitempty" jsonschema:"Filter workspaces to this group path (e.g. group/subgroup)"`
	Labels    map[string]string                `json:"labels,omitempty" jsonschema:"Filter workspaces by labels (key-value pairs)"`
	Sort      *sdktypes.WorkspaceSortableField `json:"sort,omitempty" jsonschema:"Sort order: FULL_PATH_ASC, FULL_PATH_DESC, UPDATED_AT_ASC, or UPDATED_AT_DESC"`
	Limit     *int32                           `json:"limit,omitempty" jsonschema:"Maximum number of workspaces to return (default: 10, max: 50)"`
	Cursor    *string                          `json:"cursor,omitempty" jsonschema:"Pagination cursor from previous response"`
}

// listWorkspacesOutput is the output for listing workspaces.
type listWorkspacesOutput struct {
	Workspaces []workspace `json:"workspaces" jsonschema:"List of workspaces"`
	PageInfo   pageInfo    `json:"page_info" jsonschema:"Pagination information"`
}

// ListWorkspaces returns an MCP tool for listing workspaces.
func listWorkspaces(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[listWorkspacesInput, listWorkspacesOutput]) {
	tool := mcp.Tool{
		Name:        "list_workspaces",
		Description: "List Tharsis workspaces with optional filtering by group. Supports pagination for large result sets.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Workspaces",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input listWorkspacesInput) (*mcp.CallToolResult, listWorkspacesOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, listWorkspacesOutput{}, err
		}

		sdkInput := &sdktypes.GetWorkspacesInput{
			PaginationOptions: buildPaginationOptions(input.Limit, input.Cursor),
			Sort:              input.Sort,
		}

		if input.GroupPath != nil || len(input.Labels) > 0 {
			sdkInput.Filter = &sdktypes.WorkspaceFilter{
				GroupPath: input.GroupPath,
			}

			if len(input.Labels) > 0 {
				labelFilters := make([]sdktypes.WorkspaceLabelFilter, 0, len(input.Labels))
				for k, v := range input.Labels {
					labelFilters = append(labelFilters, sdktypes.WorkspaceLabelFilter{
						Key:   k,
						Value: v,
					})
				}
				sdkInput.Filter.Labels = labelFilters
			}
		}

		result, err := client.Workspaces().GetWorkspaces(ctx, sdkInput)
		if err != nil {
			return nil, listWorkspacesOutput{}, fmt.Errorf("failed to list workspaces: %w", err)
		}

		workspaces := make([]workspace, len(result.Workspaces))
		for i, ws := range result.Workspaces {
			workspaces[i] = toWorkspace(&ws)
		}

		return nil, listWorkspacesOutput{
			Workspaces: workspaces,
			PageInfo:   buildPageInfo(result.PageInfo),
		}, nil
	}

	return tool, handler
}

// getWorkspaceInput is the input for getting a workspace.
type getWorkspaceInput struct {
	ID string `json:"id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/subgroup/workspace-name)"`
}

// getWorkspaceOutput is the output for getting a workspace.
type getWorkspaceOutput struct {
	Workspace workspace `json:"workspace" jsonschema:"The workspace details"`
}

// GetWorkspace returns an MCP tool for getting a workspace.
func getWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getWorkspaceInput, getWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "get_workspace",
		Description: "Retrieve details about a Tharsis workspace including its configuration and current state.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Workspace",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getWorkspaceInput) (*mcp.CallToolResult, getWorkspaceOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getWorkspaceOutput{}, err
		}

		workspace, err := client.Workspaces().GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{ID: &input.ID})
		if err != nil {
			return nil, getWorkspaceOutput{}, fmt.Errorf("failed to get workspace: %w", err)
		}

		return nil, getWorkspaceOutput{
			Workspace: toWorkspace(workspace),
		}, nil
	}

	return tool, handler
}

// createWorkspaceInput is the input for creating a workspace.
type createWorkspaceInput struct {
	Name               string            `json:"name" jsonschema:"required,Name of the workspace"`
	GroupPath          string            `json:"group_path" jsonschema:"required,Path to the parent group (e.g. group/subgroup)"`
	Description        string            `json:"description,omitempty" jsonschema:"Description of the workspace"`
	TerraformVersion   *string           `json:"terraform_version,omitempty" jsonschema:"Terraform CLI version to use (e.g. 1.5.0). Defaults to latest"`
	MaxJobDuration     *int32            `json:"max_job_duration,omitempty" jsonschema:"Maximum number of minutes a Terraform job is allowed to run"`
	PreventDestroyPlan *bool             `json:"prevent_destroy_plan,omitempty" jsonschema:"Prevent runs from destroying resources"`
	Labels             map[string]string `json:"labels,omitempty" jsonschema:"Labels for the workspace as key-value pairs"`
}

// createWorkspaceOutput is the output for creating a workspace.
type createWorkspaceOutput struct {
	Workspace workspace `json:"workspace" jsonschema:"The created workspace"`
}

// CreateWorkspace returns an MCP tool for creating a workspace.
func createWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[createWorkspaceInput, createWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "create_workspace",
		Description: "Create a new Tharsis workspace within a group. Workspaces contain Terraform state and run configurations.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Workspace",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input createWorkspaceInput) (*mcp.CallToolResult, createWorkspaceOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, createWorkspaceOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, trn.ToTRN(input.GroupPath, trn.ResourceTypeGroup), trn.ResourceTypeGroup); err != nil {
			return nil, createWorkspaceOutput{}, err
		}

		var labels []sdktypes.WorkspaceLabelInput
		for k, v := range input.Labels {
			labels = append(labels, sdktypes.WorkspaceLabelInput{Key: k, Value: v})
		}

		createInput := &sdktypes.CreateWorkspaceInput{
			Name:               input.Name,
			GroupPath:          input.GroupPath,
			Description:        input.Description,
			TerraformVersion:   input.TerraformVersion,
			MaxJobDuration:     input.MaxJobDuration,
			PreventDestroyPlan: input.PreventDestroyPlan,
			Labels:             labels,
		}

		workspace, err := client.Workspaces().CreateWorkspace(ctx, createInput)
		if err != nil {
			return nil, createWorkspaceOutput{}, fmt.Errorf("failed to create workspace: %w", err)
		}

		return nil, createWorkspaceOutput{
			Workspace: toWorkspace(workspace),
		}, nil
	}

	return tool, handler
}

// updateWorkspaceInput is the input for updating a workspace.
type updateWorkspaceInput struct {
	ID                 string            `json:"id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/subgroup/workspace-name)"`
	Description        string            `json:"description,omitempty" jsonschema:"New description for the workspace"`
	TerraformVersion   *string           `json:"terraform_version,omitempty" jsonschema:"New Terraform CLI version"`
	MaxJobDuration     *int32            `json:"max_job_duration,omitempty" jsonschema:"Maximum number of minutes a Terraform job is allowed to run"`
	PreventDestroyPlan *bool             `json:"prevent_destroy_plan,omitempty" jsonschema:"Update prevent destroy plan setting"`
	Labels             map[string]string `json:"labels,omitempty" jsonschema:"Labels for the workspace as key-value pairs"`
}

// updateWorkspaceOutput is the output for updating a workspace.
type updateWorkspaceOutput struct {
	Workspace workspace `json:"workspace" jsonschema:"The updated workspace"`
}

// UpdateWorkspace returns an MCP tool for updating a workspace.
func updateWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[updateWorkspaceInput, updateWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "update_workspace",
		Description: "Update an existing Tharsis workspace's configuration including description, Terraform version, and job settings.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Update Workspace",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input updateWorkspaceInput) (*mcp.CallToolResult, updateWorkspaceOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, updateWorkspaceOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, input.ID, trn.ResourceTypeWorkspace); err != nil {
			return nil, updateWorkspaceOutput{}, err
		}

		var labels []sdktypes.WorkspaceLabelInput
		for k, v := range input.Labels {
			labels = append(labels, sdktypes.WorkspaceLabelInput{Key: k, Value: v})
		}

		updateInput := &sdktypes.UpdateWorkspaceInput{
			ID:                 &input.ID,
			Description:        input.Description,
			TerraformVersion:   input.TerraformVersion,
			MaxJobDuration:     input.MaxJobDuration,
			PreventDestroyPlan: input.PreventDestroyPlan,
			Labels:             labels,
		}

		workspace, err := client.Workspaces().UpdateWorkspace(ctx, updateInput)
		if err != nil {
			return nil, updateWorkspaceOutput{}, fmt.Errorf("failed to update workspace: %w", err)
		}

		return nil, updateWorkspaceOutput{
			Workspace: toWorkspace(workspace),
		}, nil
	}

	return tool, handler
}

// deleteWorkspaceInput is the input for deleting a workspace.
// Note: Force deletion is intentionally not supported to prevent accidental
// deletion of workspaces with deployed resources.
type deleteWorkspaceInput struct {
	ID string `json:"id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/subgroup/workspace-name)"`
}

// deleteWorkspaceOutput is the output for deleting a workspace.
type deleteWorkspaceOutput struct {
	Message string `json:"message" jsonschema:"Deletion confirmation message"`
	Success bool   `json:"success" jsonschema:"Whether deletion was successful"`
}

// DeleteWorkspace returns an MCP tool for deleting a workspace.
func deleteWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[deleteWorkspaceInput, deleteWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "delete_workspace",
		Description: "Delete a Tharsis workspace. Use with caution as this operation is irreversible.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Workspace",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input deleteWorkspaceInput) (*mcp.CallToolResult, deleteWorkspaceOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, deleteWorkspaceOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, input.ID, trn.ResourceTypeWorkspace); err != nil {
			return nil, deleteWorkspaceOutput{}, err
		}

		deleteInput := &sdktypes.DeleteWorkspaceInput{
			ID: &input.ID,
		}

		if err := client.Workspaces().DeleteWorkspace(ctx, deleteInput); err != nil {
			return nil, deleteWorkspaceOutput{}, fmt.Errorf("failed to delete workspace: %w", err)
		}

		return nil, deleteWorkspaceOutput{
			Message: fmt.Sprintf("Workspace %s deleted successfully", input.ID),
			Success: true,
		}, nil
	}

	return tool, handler
}

// getWorkspaceOutputsInput is the input for getting workspace outputs.
type getWorkspaceOutputsInput struct {
	ID string `json:"id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/subgroup/workspace-name)"`
}

// getWorkspaceOutputsOutput is the output for getting workspace outputs.
type getWorkspaceOutputsOutput struct {
	Outputs map[string]string `json:"outputs" jsonschema:"Non-sensitive Terraform state outputs as key-value pairs"`
}

// GetWorkspaceOutputs returns an MCP tool for getting workspace outputs.
func getWorkspaceOutputs(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getWorkspaceOutputsInput, getWorkspaceOutputsOutput]) {
	tool := mcp.Tool{
		Name:        "get_workspace_outputs",
		Description: "Retrieve non-sensitive Terraform state outputs from a workspace. Sensitive outputs are filtered out for security.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Workspace Outputs",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getWorkspaceOutputsInput) (*mcp.CallToolResult, getWorkspaceOutputsOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getWorkspaceOutputsOutput{}, err
		}

		workspace, err := client.Workspaces().GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{ID: &input.ID})
		if err != nil {
			return nil, getWorkspaceOutputsOutput{}, fmt.Errorf("failed to get workspace: %w", err)
		}

		if workspace.CurrentStateVersion == nil {
			return nil, getWorkspaceOutputsOutput{Outputs: map[string]string{}}, nil
		}

		stateVersion, err := client.StateVersions().GetStateVersion(ctx, &sdktypes.GetStateVersionInput{ID: workspace.CurrentStateVersion.Metadata.ID})
		if err != nil {
			return nil, getWorkspaceOutputsOutput{}, fmt.Errorf("failed to get state version: %w", err)
		}

		outputs := make(map[string]string)
		for _, output := range stateVersion.Outputs {
			// Skip sensitive outputs
			if output.Sensitive {
				continue
			}

			valueBytes, err := ctyjson.Marshal(output.Value, output.Type)
			if err != nil {
				return nil, getWorkspaceOutputsOutput{}, fmt.Errorf("failed to marshal output %s: %w", output.Name, err)
			}

			outputs[output.Name] = string(valueBytes)
		}

		return nil, getWorkspaceOutputsOutput{Outputs: outputs}, nil
	}

	return tool, handler
}
