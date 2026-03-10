package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// terraformModule is the output type for Terraform modules.
type terraformModule struct {
	ID            string `json:"id" jsonschema:"Unique identifier for the module"`
	TRN           string `json:"trn" jsonschema:"Tharsis Resource Name (TRN) for the module"`
	Name          string `json:"name" jsonschema:"Module name"`
	System        string `json:"system" jsonschema:"Module system (e.g., aws, azure)"`
	GroupID       string `json:"group_id" jsonschema:"ID of the group containing this module"`
	RepositoryURL string `json:"repository_url" jsonschema:"URL to the module's source repository"`
	Private       bool   `json:"private" jsonschema:"Whether the module is private"`
	CreatedBy     string `json:"created_by" jsonschema:"Username or service account that created this module"`
}

// toTerraformModule converts a proto TerraformModule to the MCP output type.
func toTerraformModule(m *pb.TerraformModule) terraformModule {
	return terraformModule{
		ID:            m.Metadata.Id,
		TRN:           m.Metadata.Trn,
		Name:          m.Name,
		System:        m.System,
		GroupID:       m.GroupId,
		RepositoryURL: m.RepositoryUrl,
		Private:       m.Private,
		CreatedBy:     m.CreatedBy,
	}
}

// listTerraformModulesInput is the input for listing Terraform modules.
type listTerraformModulesInput struct {
	GroupID          *string `json:"group_id,omitempty" jsonschema:"Filter modules to this group ID or TRN (e.g. Ul8yZ... or trn:group:group-path)"`
	Search           *string `json:"search,omitempty" jsonschema:"Search term to filter by module path"`
	IncludeInherited bool    `json:"include_inherited,omitempty" jsonschema:"Include modules inherited from parent groups"`
	Sort             *string `json:"sort,omitempty" jsonschema:"Sort field (NAME_ASC, NAME_DESC, UPDATED_AT_ASC, UPDATED_AT_DESC, GROUP_LEVEL_ASC, GROUP_LEVEL_DESC)"`
	Limit            *int32  `json:"limit,omitempty" jsonschema:"Maximum number of modules to return"`
	Cursor           *string `json:"cursor,omitempty" jsonschema:"Pagination cursor for next page"`
}

// listTerraformModulesOutput is the output for listing Terraform modules.
type listTerraformModulesOutput struct {
	Modules  []terraformModule `json:"modules" jsonschema:"List of Terraform modules"`
	PageInfo pageInfo          `json:"page_info" jsonschema:"Pagination information"`
}

// ListTerraformModules returns an MCP tool for listing Terraform modules.
func listTerraformModules(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*listTerraformModulesInput, *listTerraformModulesOutput]) {
	tool := mcp.Tool{
		Name:        "list_terraform_modules",
		Description: "List Terraform modules in the Tharsis registry. Supports search, sorting, and pagination.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Terraform Modules",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *listTerraformModulesInput) (*mcp.CallToolResult, *listTerraformModulesOutput, error) {
		resp, err := tc.grpcClient.TerraformModulesClient.GetTerraformModules(ctx, &pb.GetTerraformModulesRequest{
			GroupId:           input.GroupID,
			Search:            input.Search,
			IncludeInherited:  input.IncludeInherited,
			Sort:              toSortEnum[pb.TerraformModuleSortableField](input.Sort, pb.TerraformModuleSortableField_value),
			PaginationOptions: buildPaginationOptions(input.Limit, input.Cursor),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list modules: %w", err)
		}

		modules := make([]terraformModule, len(resp.Modules))
		for i, m := range resp.Modules {
			modules[i] = toTerraformModule(m)
		}

		return nil, &listTerraformModulesOutput{
			Modules:  modules,
			PageInfo: buildPageInfo(resp.PageInfo),
		}, nil
	}

	return tool, handler
}

// getTerraformModuleInput is the input for getting a Terraform module.
type getTerraformModuleInput struct {
	ID string `json:"id" jsonschema:"required,Module ID or TRN (e.g. Ul8yZ... or trn:terraform_module:group/module-name/system)"`
}

// getTerraformModuleOutput is the output for getting a Terraform module.
type getTerraformModuleOutput struct {
	Module terraformModule `json:"module" jsonschema:"The Terraform module"`
}

// GetTerraformModule returns an MCP tool for getting a Terraform module.
func getTerraformModule(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getTerraformModuleInput, *getTerraformModuleOutput]) {
	tool := mcp.Tool{
		Name:        "get_terraform_module",
		Description: "Get details of a specific Terraform module by ID or TRN.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Terraform Module",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getTerraformModuleInput) (*mcp.CallToolResult, *getTerraformModuleOutput, error) {
		resp, err := tc.grpcClient.TerraformModulesClient.GetTerraformModuleByID(ctx, &pb.GetTerraformModuleByIDRequest{
			Id: input.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get module: %w", err)
		}

		return nil, &getTerraformModuleOutput{
			Module: toTerraformModule(resp),
		}, nil
	}

	return tool, handler
}

// createTerraformModuleInput is the input for creating a Terraform module.
type createTerraformModuleInput struct {
	Name          string `json:"name" jsonschema:"required,Module name"`
	System        string `json:"system" jsonschema:"required,Module system (e.g., aws, gcp, azure)"`
	GroupID       string `json:"group_id" jsonschema:"required,Group ID or TRN (e.g. Ul8yZ... or trn:group:group-path)"`
	RepositoryURL string `json:"repository_url,omitempty" jsonschema:"URL to the module's source repository"`
	Private       bool   `json:"private,omitempty" jsonschema:"Whether the module should be private (default: false)"`
}

// createTerraformModuleOutput is the output for creating a Terraform module.
type createTerraformModuleOutput struct {
	Module terraformModule `json:"module" jsonschema:"The created Terraform module"`
}

// CreateTerraformModule returns an MCP tool for creating a Terraform module.
func createTerraformModule(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*createTerraformModuleInput, *createTerraformModuleOutput]) {
	tool := mcp.Tool{
		Name:        "create_terraform_module",
		Description: "Create a new Terraform module in the Tharsis registry.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Terraform Module",
			DestructiveHint: ptr.Bool(false),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *createTerraformModuleInput) (*mcp.CallToolResult, *createTerraformModuleOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.GroupID, trn.ResourceTypeGroup); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.TerraformModulesClient.CreateTerraformModule(ctx, &pb.CreateTerraformModuleRequest{
			Name:          input.Name,
			System:        input.System,
			GroupId:       input.GroupID,
			RepositoryUrl: input.RepositoryURL,
			Private:       input.Private,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create module: %w", err)
		}

		return nil, &createTerraformModuleOutput{
			Module: toTerraformModule(resp),
		}, nil
	}

	return tool, handler
}

// updateTerraformModuleInput is the input for updating a Terraform module.
type updateTerraformModuleInput struct {
	ID            string  `json:"id" jsonschema:"required,Module ID or TRN (e.g. Ul8yZ... or trn:terraform_module:group/module-name/system)"`
	RepositoryURL *string `json:"repository_url,omitempty" jsonschema:"New repository URL"`
	Private       *bool   `json:"private,omitempty" jsonschema:"Whether the module should be private"`
}

// updateTerraformModuleOutput is the output for updating a Terraform module.
type updateTerraformModuleOutput struct {
	Module terraformModule `json:"module" jsonschema:"The updated Terraform module"`
}

// UpdateTerraformModule returns an MCP tool for updating a Terraform module.
func updateTerraformModule(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*updateTerraformModuleInput, *updateTerraformModuleOutput]) {
	tool := mcp.Tool{
		Name:        "update_terraform_module",
		Description: "Update an existing Terraform module's repository URL or privacy setting.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Update Terraform Module",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *updateTerraformModuleInput) (*mcp.CallToolResult, *updateTerraformModuleOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeTerraformModule); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.TerraformModulesClient.UpdateTerraformModule(ctx, &pb.UpdateTerraformModuleRequest{
			Id:            input.ID,
			RepositoryUrl: input.RepositoryURL,
			Private:       input.Private,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update module: %w", err)
		}

		return nil, &updateTerraformModuleOutput{
			Module: toTerraformModule(resp),
		}, nil
	}

	return tool, handler
}

// deleteTerraformModuleInput is the input for deleting a Terraform module.
type deleteTerraformModuleInput struct {
	ID string `json:"id" jsonschema:"required,Module ID or TRN (e.g. Ul8yZ... or trn:terraform_module:group/module-name/system)"`
}

// deleteTerraformModuleOutput is the output for deleting a Terraform module.
type deleteTerraformModuleOutput struct {
	Message string `json:"message" jsonschema:"Deletion confirmation message"`
	Success bool   `json:"success" jsonschema:"Whether deletion was successful"`
}

// DeleteTerraformModule returns an MCP tool for deleting a Terraform module.
func deleteTerraformModule(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*deleteTerraformModuleInput, *deleteTerraformModuleOutput]) {
	tool := mcp.Tool{
		Name:        "delete_terraform_module",
		Description: "Delete a Terraform module from the registry. Use with caution as this operation is irreversible.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Terraform Module",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *deleteTerraformModuleInput) (*mcp.CallToolResult, *deleteTerraformModuleOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeTerraformModule); err != nil {
			return nil, nil, err
		}

		if _, err := tc.grpcClient.TerraformModulesClient.DeleteTerraformModule(ctx, &pb.DeleteTerraformModuleRequest{
			Id: input.ID,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to delete module: %w", err)
		}

		return nil, &deleteTerraformModuleOutput{
			Message: fmt.Sprintf("Module %s deleted successfully", input.ID),
			Success: true,
		}, nil
	}

	return tool, handler
}
