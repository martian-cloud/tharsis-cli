package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// run represents a Tharsis run in MCP responses.
type run struct {
	ID                     string   `json:"id" jsonschema:"The unique identifier of the run"`
	Status                 string   `json:"status" jsonschema:"Overall run status (e.g. pending plan_queued planned applied errored canceled)"`
	WorkspaceID            string   `json:"workspace_id" jsonschema:"The unique identifier of the workspace"`
	CreatedBy              string   `json:"created_by" jsonschema:"Username or service account that created this run"`
	TerraformVersion       string   `json:"terraform_version" jsonschema:"Version of Terraform used to execute this run"`
	IsDestroy              bool     `json:"is_destroy" jsonschema:"True if this run will destroy resources instead of creating/updating them"`
	Speculative            bool     `json:"speculative" jsonschema:"True if this is a speculative plan (plan-only no apply will occur)"`
	Refresh                bool     `json:"refresh" jsonschema:"True if this run will refresh the state"`
	RefreshOnly            bool     `json:"refresh_only" jsonschema:"True if this run will only refresh the state without applying changes"`
	ConfigurationVersionID *string  `json:"configuration_version_id,omitempty" jsonschema:"ID of the configuration version used for this run"`
	ModuleSource           *string  `json:"module_source,omitempty" jsonschema:"Source location of the Terraform module used for this run"`
	ModuleVersion          *string  `json:"module_version,omitempty" jsonschema:"Version of the module used for this run"`
	ModuleDigest           *string  `json:"module_digest,omitempty" jsonschema:"Digest of the module used for this run"`
	TargetAddresses        []string `json:"target_addresses,omitempty" jsonschema:"List of resource addresses targeted by this run"`
	ForceCanceled          bool     `json:"force_canceled,omitempty" jsonschema:"True if this run was force canceled"`
	ForceCanceledBy        *string  `json:"force_canceled_by,omitempty" jsonschema:"Username or service account that force canceled this run"`
	PlanID                 string   `json:"plan_id" jsonschema:"ID of the plan associated with this run"`
	ApplyID                string   `json:"apply_id" jsonschema:"ID of the apply associated with this run"`
	HasChanges             bool     `json:"has_changes" jsonschema:"True if the plan has changes"`
	TRN                    string   `json:"trn" jsonschema:"Tharsis Resource Name"`
}

// toRun converts a proto run to MCP run.
func toRun(r *pb.Run) *run {
	return &run{
		ID:                     r.Metadata.Id,
		TRN:                    r.Metadata.Trn,
		Status:                 r.Status,
		WorkspaceID:            r.WorkspaceId,
		CreatedBy:              r.CreatedBy,
		TerraformVersion:       r.TerraformVersion,
		IsDestroy:              r.IsDestroy,
		Speculative:            r.Speculative,
		Refresh:                r.Refresh,
		RefreshOnly:            r.RefreshOnly,
		ConfigurationVersionID: r.ConfigurationVersionId,
		ModuleSource:           r.ModuleSource,
		ModuleVersion:          r.ModuleVersion,
		ModuleDigest:           r.ModuleDigest,
		TargetAddresses:        r.TargetAddresses,
		ForceCanceled:          r.ForceCanceled,
		ForceCanceledBy:        r.ForceCanceledBy,
		PlanID:                 r.PlanId,
		ApplyID:                r.ApplyId,
		HasChanges:             r.HasChanges,
	}
}

// listRunsInput is the input for listing runs.
type listRunsInput struct {
	WorkspaceID *string `json:"workspace_id,omitempty" jsonschema:"Filter runs to this workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace-name)"`
	Sort        *string `json:"sort,omitempty" jsonschema:"Sort order: CREATED_AT_ASC, CREATED_AT_DESC, UPDATED_AT_ASC, or UPDATED_AT_DESC"`
	Limit       *int32  `json:"limit,omitempty" jsonschema:"Maximum number of runs to return (default: 10, max: 50)"`
	Cursor      *string `json:"cursor,omitempty" jsonschema:"Pagination cursor from previous response"`
}

// listRunsOutput is the output for listing runs.
type listRunsOutput struct {
	Runs     []*run   `json:"runs" jsonschema:"List of runs"`
	PageInfo pageInfo `json:"page_info" jsonschema:"Pagination information"`
}

// ListRuns returns an MCP tool for listing runs.
func listRuns(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*listRunsInput, *listRunsOutput]) {
	tool := mcp.Tool{
		Name:        "list_runs",
		Description: "List Tharsis runs with optional filtering by workspace. Supports pagination for large result sets.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Runs",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *listRunsInput) (*mcp.CallToolResult, *listRunsOutput, error) {
		req := &pb.GetRunsRequest{
			PaginationOptions: buildPaginationOptions(input.Limit, input.Cursor),
			Sort:              toSortEnum[pb.RunSortableField](input.Sort, pb.RunSortableField_value),
			WorkspaceId:       input.WorkspaceID,
		}

		resp, err := tc.grpcClient.RunsClient.GetRuns(ctx, req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list runs: %w", err)
		}

		runs := make([]*run, len(resp.Runs))
		for i, r := range resp.Runs {
			runs[i] = toRun(r)
		}

		return nil, &listRunsOutput{
			Runs:     runs,
			PageInfo: buildPageInfo(resp.PageInfo),
		}, nil
	}

	return tool, handler
}

// getRunInput is the input for the get_run tool.
type getRunInput struct {
	ID string `json:"id" jsonschema:"required,Run ID or TRN (e.g. Ul8yZ... or trn:run:my-group/my-workspace/run-id)"`
}

// getRunOutput is the output for the get_run tool.
type getRunOutput struct {
	Run *run `json:"run,omitempty" jsonschema:"The run details"`
}

// GetRun returns an MCP tool for retrieving Tharsis run information.
func getRun(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getRunInput, *getRunOutput]) {
	tool := mcp.Tool{
		Name:        "get_run",
		Description: "Retrieve a Tharsis run by ID. Returns run status, workspace ID, stage information, and error messages for troubleshooting failed runs.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Run",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getRunInput) (*mcp.CallToolResult, *getRunOutput, error) {
		resp, err := tc.grpcClient.RunsClient.GetRunByID(ctx, &pb.GetRunByIDRequest{Id: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get run %q: %w", input.ID, err)
		}

		return nil, &getRunOutput{
			Run: toRun(resp),
		}, nil
	}

	return tool, handler
}

// createRunInput is the input for the create_run tool.
type createRunInput struct {
	WorkspaceID            string   `json:"workspace_id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace-name)"`
	ConfigurationVersionID *string  `json:"configuration_version_id,omitempty" jsonschema:"ID of an existing configuration version to use for this run"`
	ModuleSource           *string  `json:"module_source,omitempty" jsonschema:"Source location of the Terraform module (e.g. registry.terraform.io/namespace/module)"`
	ModuleVersion          *string  `json:"module_version,omitempty" jsonschema:"Version of the module to use (e.g. 1.0.0)"`
	TerraformVersion       *string  `json:"terraform_version,omitempty" jsonschema:"Version of Terraform to use (e.g. 1.5.0)"`
	IsDestroy              bool     `json:"is_destroy,omitempty" jsonschema:"True to destroy resources instead of creating/updating them"`
	Speculative            *bool    `json:"speculative,omitempty" jsonschema:"True for a speculative plan (plan-only no apply will occur)"`
	Refresh                bool     `json:"refresh,omitempty" jsonschema:"True to refresh the state"`
	RefreshOnly            bool     `json:"refresh_only,omitempty" jsonschema:"True to only refresh the state without applying changes"`
	TargetAddresses        []string `json:"target_addresses,omitempty" jsonschema:"List of resource addresses to target (e.g. aws_instance.example)"`
}

// createRunOutput is the output for the create_run tool.
type createRunOutput struct {
	Run *run `json:"run,omitempty" jsonschema:"The created run"`
}

// CreateRun returns an MCP tool for creating a Tharsis run.
func createRun(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*createRunInput, *createRunOutput]) {
	tool := mcp.Tool{
		Name:        "create_run",
		Description: "Create a new Tharsis run in a workspace. Runs execute Terraform plans and applies.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Run",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *createRunInput) (*mcp.CallToolResult, *createRunOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.WorkspaceID, trn.ResourceTypeWorkspace); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.RunsClient.CreateRun(ctx, &pb.CreateRunRequest{
			WorkspaceId:            input.WorkspaceID,
			ConfigurationVersionId: input.ConfigurationVersionID,
			ModuleSource:           input.ModuleSource,
			ModuleVersion:          input.ModuleVersion,
			TerraformVersion:       input.TerraformVersion,
			IsDestroy:              input.IsDestroy,
			Speculative:            input.Speculative,
			Refresh:                input.Refresh,
			RefreshOnly:            input.RefreshOnly,
			TargetAddresses:        input.TargetAddresses,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create run in workspace %q: %w", input.WorkspaceID, err)
		}

		return nil, &createRunOutput{
			Run: toRun(resp),
		}, nil
	}

	return tool, handler
}

// applyRunInput is the input for the apply_run tool.
type applyRunInput struct {
	ID string `json:"id" jsonschema:"required,Run ID or TRN (e.g. Ul8yZ... or trn:run:my-group/my-workspace/run-id)"`
}

// applyRunOutput is the output for the apply_run tool.
type applyRunOutput struct {
	RunID  string `json:"run_id" jsonschema:"The unique identifier of the run"`
	Status string `json:"status" jsonschema:"Status of the run after apply"`
}

// ApplyRun returns an MCP tool for applying a Tharsis run.
func applyRun(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*applyRunInput, *applyRunOutput]) {
	tool := mcp.Tool{
		Name:        "apply_run",
		Description: "Apply a Tharsis run to execute the planned infrastructure changes.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Apply Run",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *applyRunInput) (*mcp.CallToolResult, *applyRunOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeRun); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.RunsClient.ApplyRun(ctx, &pb.ApplyRunRequest{RunId: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to apply run %q: %w", input.ID, err)
		}

		return nil, &applyRunOutput{
			RunID:  resp.Metadata.Id,
			Status: resp.Status,
		}, nil
	}

	return tool, handler
}

// cancelRunInput is the input for canceling a run.
type cancelRunInput struct {
	ID    string `json:"id" jsonschema:"required,Run ID or TRN (e.g. Ul8yZ... or trn:run:my-group/my-workspace/run-id)"`
	Force *bool  `json:"force,omitempty" jsonschema:"Force cancel the run (use when graceful cancel is not enough)"`
}

// cancelRunOutput is the output for canceling a run.
type cancelRunOutput struct {
	Message string `json:"message" jsonschema:"Cancellation confirmation message"`
	Success bool   `json:"success" jsonschema:"Whether cancellation was successful"`
}

// CancelRun returns an MCP tool for canceling a run.
func cancelRun(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*cancelRunInput, *cancelRunOutput]) {
	tool := mcp.Tool{
		Name:        "cancel_run",
		Description: "Cancel a Terraform run. Use force option when graceful cancellation is not sufficient.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Cancel Run",
			DestructiveHint: ptr.Bool(false),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *cancelRunInput) (*mcp.CallToolResult, *cancelRunOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeRun); err != nil {
			return nil, nil, err
		}

		if _, err := tc.grpcClient.RunsClient.CancelRun(ctx, &pb.CancelRunRequest{
			Id:    input.ID,
			Force: input.Force,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to cancel run: %w", err)
		}

		return nil, &cancelRunOutput{
			Message: fmt.Sprintf("Run %s cancellation initiated", input.ID),
			Success: true,
		}, nil
	}

	return tool, handler
}
