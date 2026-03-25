package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// workspace represents a Tharsis workspace in MCP responses.
type workspace struct {
	ID                    string            `json:"id" jsonschema:"The unique identifier of the workspace"`
	Path                  string            `json:"path" jsonschema:"Full path to the workspace"`
	Name                  string            `json:"name" jsonschema:"Workspace name"`
	GroupID               string            `json:"group_id" jsonschema:"ID of the parent group"`
	Description           string            `json:"description" jsonschema:"Workspace description"`
	Labels                map[string]string `json:"labels,omitempty" jsonschema:"Workspace labels"`
	Locked                bool              `json:"locked" jsonschema:"Whether the workspace is locked"`
	DirtyState            bool              `json:"dirty_state" jsonschema:"Whether the workspace has uncommitted changes"`
	CurrentJobID          string            `json:"current_job_id,omitempty" jsonschema:"ID of the current running job"`
	CurrentStateVersionID string            `json:"current_state_version_id,omitempty" jsonschema:"ID of the current state version"`
	CreatedBy             string            `json:"created_by" jsonschema:"Username or service account that created this workspace"`
	TRN                   string            `json:"trn" jsonschema:"Tharsis Resource Name"`
}

// toWorkspace converts a proto workspace to MCP workspace.
func toWorkspace(ws *pb.Workspace) workspace {
	return workspace{
		ID:                    ws.Metadata.Id,
		TRN:                   ws.Metadata.Trn,
		Path:                  ws.FullPath,
		Name:                  ws.Name,
		GroupID:               ws.GroupId,
		Description:           ws.Description,
		Labels:                ws.Labels,
		Locked:                ws.Locked,
		DirtyState:            ws.DirtyState,
		CurrentJobID:          ws.CurrentJobId,
		CurrentStateVersionID: ws.CurrentStateVersionId,
		CreatedBy:             ws.CreatedBy,
	}
}

// listWorkspacesInput is the input for listing workspaces.
type listWorkspacesInput struct {
	GroupID *string           `json:"group_id,omitempty" jsonschema:"Filter workspaces to this group ID or TRN (e.g. Ul8yZ... or trn:group:group/subgroup)"`
	Search  *string           `json:"search,omitempty" jsonschema:"Search term to filter by workspace path"`
	Labels  map[string]string `json:"labels,omitempty" jsonschema:"Filter workspaces by labels (key-value pairs)"`
	Sort    *string           `json:"sort,omitempty" jsonschema:"Sort order: FULL_PATH_ASC, FULL_PATH_DESC, UPDATED_AT_ASC, or UPDATED_AT_DESC"`
	Limit   *int32            `json:"limit,omitempty" jsonschema:"Maximum number of workspaces to return (default: 10, max: 50)"`
	Cursor  *string           `json:"cursor,omitempty" jsonschema:"Pagination cursor from previous response"`
}

// listWorkspacesOutput is the output for listing workspaces.
type listWorkspacesOutput struct {
	Workspaces []workspace `json:"workspaces" jsonschema:"List of workspaces"`
	PageInfo   pageInfo    `json:"page_info" jsonschema:"Pagination information"`
}

// ListWorkspaces returns an MCP tool for listing workspaces.
func listWorkspaces(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*listWorkspacesInput, *listWorkspacesOutput]) {
	tool := mcp.Tool{
		Name:        "list_workspaces",
		Description: "List Tharsis workspaces with optional filtering by group. Supports pagination for large result sets.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Workspaces",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *listWorkspacesInput) (*mcp.CallToolResult, *listWorkspacesOutput, error) {
		resp, err := tc.grpcClient.WorkspacesClient.GetWorkspaces(ctx, &pb.GetWorkspacesRequest{
			GroupId:           input.GroupID,
			Search:            input.Search,
			LabelFilters:      input.Labels,
			Sort:              toSortEnum[pb.WorkspaceSortableField](input.Sort, pb.WorkspaceSortableField_value),
			PaginationOptions: buildPaginationOptions(input.Limit, input.Cursor),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list workspaces: %w", err)
		}

		workspaces := make([]workspace, len(resp.Workspaces))
		for i, ws := range resp.Workspaces {
			workspaces[i] = toWorkspace(ws)
		}

		return nil, &listWorkspacesOutput{
			Workspaces: workspaces,
			PageInfo:   buildPageInfo(resp.PageInfo),
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
func getWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getWorkspaceInput, *getWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "get_workspace",
		Description: "Retrieve details about a Tharsis workspace including its configuration and current state.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Workspace",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getWorkspaceInput) (*mcp.CallToolResult, *getWorkspaceOutput, error) {
		resp, err := tc.grpcClient.WorkspacesClient.GetWorkspaceByID(ctx, &pb.GetWorkspaceByIDRequest{Id: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get workspace: %w", err)
		}

		return nil, &getWorkspaceOutput{
			Workspace: toWorkspace(resp),
		}, nil
	}

	return tool, handler
}

// createWorkspaceInput is the input for creating a workspace.
type createWorkspaceInput struct {
	Name               string            `json:"name" jsonschema:"required,Name of the workspace"`
	GroupID            string            `json:"group_id" jsonschema:"required,Group ID or TRN (e.g. Ul8yZ... or trn:group:group/subgroup)"`
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
func createWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*createWorkspaceInput, *createWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "create_workspace",
		Description: "Create a new Tharsis workspace within a group. Workspaces contain Terraform state and run configurations.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Workspace",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *createWorkspaceInput) (*mcp.CallToolResult, *createWorkspaceOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.GroupID, trn.ResourceTypeGroup); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.WorkspacesClient.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{
			Name:               input.Name,
			GroupId:            input.GroupID,
			Description:        input.Description,
			TerraformVersion:   ptr.ToString(input.TerraformVersion),
			MaxJobDuration:     input.MaxJobDuration,
			PreventDestroyPlan: ptr.ToBool(input.PreventDestroyPlan),
			Labels:             input.Labels,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create workspace: %w", err)
		}

		return nil, &createWorkspaceOutput{
			Workspace: toWorkspace(resp),
		}, nil
	}

	return tool, handler
}

// updateWorkspaceInput is the input for updating a workspace.
type updateWorkspaceInput struct {
	ID                 string            `json:"id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/subgroup/workspace-name)"`
	Description        *string           `json:"description,omitempty" jsonschema:"New description for the workspace"`
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
func updateWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*updateWorkspaceInput, *updateWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "update_workspace",
		Description: "Update an existing Tharsis workspace's configuration including description, Terraform version, and job settings.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Update Workspace",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *updateWorkspaceInput) (*mcp.CallToolResult, *updateWorkspaceOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeWorkspace); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.WorkspacesClient.UpdateWorkspace(ctx, &pb.UpdateWorkspaceRequest{
			Id:                 input.ID,
			Description:        input.Description,
			TerraformVersion:   input.TerraformVersion,
			MaxJobDuration:     input.MaxJobDuration,
			PreventDestroyPlan: input.PreventDestroyPlan,
			Labels:             input.Labels,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update workspace: %w", err)
		}

		return nil, &updateWorkspaceOutput{
			Workspace: toWorkspace(resp),
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
func deleteWorkspace(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*deleteWorkspaceInput, *deleteWorkspaceOutput]) {
	tool := mcp.Tool{
		Name:        "delete_workspace",
		Description: "Delete a Tharsis workspace. Use with caution as this operation is irreversible.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Workspace",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *deleteWorkspaceInput) (*mcp.CallToolResult, *deleteWorkspaceOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeWorkspace); err != nil {
			return nil, nil, err
		}

		if _, err := tc.grpcClient.WorkspacesClient.DeleteWorkspace(ctx, &pb.DeleteWorkspaceRequest{
			Id: input.ID,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to delete workspace: %w", err)
		}

		return nil, &deleteWorkspaceOutput{
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
func getWorkspaceOutputs(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getWorkspaceOutputsInput, *getWorkspaceOutputsOutput]) {
	tool := mcp.Tool{
		Name:        "get_workspace_outputs",
		Description: "Retrieve non-sensitive Terraform state outputs from a workspace. Sensitive outputs are filtered out for security.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Workspace Outputs",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getWorkspaceOutputsInput) (*mcp.CallToolResult, *getWorkspaceOutputsOutput, error) {
		workspace, err := tc.grpcClient.WorkspacesClient.GetWorkspaceByID(ctx, &pb.GetWorkspaceByIDRequest{Id: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get workspace: %w", err)
		}

		if workspace.CurrentStateVersionId == "" {
			return nil, &getWorkspaceOutputsOutput{Outputs: map[string]string{}}, nil
		}

		resp, err := tc.grpcClient.StateVersionsClient.GetStateVersionOutputs(ctx, &pb.GetStateVersionOutputsRequest{
			StateVersionId: workspace.CurrentStateVersionId,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get state version outputs: %w", err)
		}

		outputs := make(map[string]string)
		for _, output := range resp.StateVersionOutputs {
			// Skip sensitive outputs
			if output.Sensitive {
				continue
			}

			outputType, err := ctyjson.UnmarshalType(output.Type)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to unmarshal type for output %s: %w", output.Name, err)
			}

			outputValue, err := ctyjson.Unmarshal(output.Value, outputType)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to unmarshal value for output %s: %w", output.Name, err)
			}

			valueBytes, err := ctyjson.Marshal(outputValue, outputType)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal output %s: %w", output.Name, err)
			}

			outputs[output.Name] = string(valueBytes)
		}

		return nil, &getWorkspaceOutputsOutput{Outputs: outputs}, nil
	}

	return tool, handler
}
