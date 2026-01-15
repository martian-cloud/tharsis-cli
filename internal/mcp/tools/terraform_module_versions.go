package tools

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/slug"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// terraformModuleVersion is the output type for Terraform module versions.
type terraformModuleVersion struct {
	ID          string   `json:"id" jsonschema:"Unique identifier for the module version"`
	TRN         string   `json:"trn" jsonschema:"Tharsis Resource Name (TRN) for the module version"`
	ModuleID    string   `json:"module_id" jsonschema:"ID of the parent module"`
	Version     string   `json:"version" jsonschema:"Semantic version string (e.g., 1.0.0)"`
	SHASum      string   `json:"sha_sum" jsonschema:"SHA256 checksum of the module package"`
	Status      string   `json:"status" jsonschema:"Upload and processing status: pending, upload_in_progress, uploaded, or errored"`
	Error       string   `json:"error,omitempty" jsonschema:"Error message if upload or processing failed"`
	Diagnostics string   `json:"diagnostics,omitempty" jsonschema:"Diagnostic information from module processing"`
	Submodules  []string `json:"submodules" jsonschema:"List of submodule paths found in the module"`
	Examples    []string `json:"examples" jsonschema:"List of example paths found in the module"`
	Latest      bool     `json:"latest" jsonschema:"Whether this is the latest version of the module"`
}

// toTerraformModuleVersion converts an SDK TerraformModuleVersion to the MCP output type.
func toTerraformModuleVersion(v *sdktypes.TerraformModuleVersion) terraformModuleVersion {
	return terraformModuleVersion{
		ID:          v.Metadata.ID,
		TRN:         v.Metadata.TRN,
		ModuleID:    v.ModuleID,
		Version:     v.Version,
		SHASum:      v.SHASum,
		Status:      v.Status,
		Error:       v.Error,
		Diagnostics: v.Diagnostics,
		Submodules:  v.Submodules,
		Examples:    v.Examples,
		Latest:      v.Latest,
	}
}

// listTerraformModuleVersionsInput is the input for listing Terraform module versions.
type listTerraformModuleVersionsInput struct {
	ModuleID string                                        `json:"module_id" jsonschema:"required,ID or TRN of the Terraform module (e.g. trn:terraform_module:group/module-name/system)"`
	Sort     *sdktypes.TerraformModuleVersionSortableField `json:"sort,omitempty" jsonschema:"Sort field (CREATED_AT_ASC, CREATED_AT_DESC, UPDATED_AT_ASC, UPDATED_AT_DESC)"`
	Limit    *int32                                        `json:"limit,omitempty" jsonschema:"Maximum number of versions to return"`
	Cursor   *string                                       `json:"cursor,omitempty" jsonschema:"Pagination cursor for next page"`
}

// listTerraformModuleVersionsOutput is the output for listing Terraform module versions.
type listTerraformModuleVersionsOutput struct {
	ModuleVersions []terraformModuleVersion `json:"module_versions" jsonschema:"List of Terraform module versions"`
	PageInfo       pageInfo                 `json:"page_info" jsonschema:"Pagination information"`
}

// ListTerraformModuleVersions returns an MCP tool for listing Terraform module versions.
func listTerraformModuleVersions(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[listTerraformModuleVersionsInput, listTerraformModuleVersionsOutput]) {
	tool := mcp.Tool{
		Name:        "list_terraform_module_versions",
		Description: "List versions of a Terraform module. Supports sorting and pagination.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Terraform Module Versions",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input listTerraformModuleVersionsInput) (*mcp.CallToolResult, listTerraformModuleVersionsOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, listTerraformModuleVersionsOutput{}, err
		}

		resp, err := client.TerraformModuleVersions().GetModuleVersions(ctx, &sdktypes.GetTerraformModuleVersionsInput{
			TerraformModuleID: input.ModuleID,
			Sort:              input.Sort,
			PaginationOptions: &sdktypes.PaginationOptions{
				Limit:  input.Limit,
				Cursor: input.Cursor,
			},
		})
		if err != nil {
			return nil, listTerraformModuleVersionsOutput{}, fmt.Errorf("failed to list module versions: %w", err)
		}

		versions := make([]terraformModuleVersion, len(resp.ModuleVersions))
		for i, v := range resp.ModuleVersions {
			versions[i] = toTerraformModuleVersion(&v)
		}

		return nil, listTerraformModuleVersionsOutput{
			ModuleVersions: versions,
			PageInfo:       buildPageInfo(resp.PageInfo),
		}, nil
	}

	return tool, handler
}

// getTerraformModuleVersionInput is the input for getting a Terraform module version.
type getTerraformModuleVersionInput struct {
	ID string `json:"id" jsonschema:"required,Module version ID or TRN (e.g. trn:terraform_module_version:group/module-name/system/1.0.0)"`
}

// getTerraformModuleVersionOutput is the output for getting a Terraform module version.
type getTerraformModuleVersionOutput struct {
	ModuleVersion terraformModuleVersion `json:"module_version" jsonschema:"The Terraform module version"`
}

// GetTerraformModuleVersion returns an MCP tool for getting a Terraform module version.
func getTerraformModuleVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getTerraformModuleVersionInput, getTerraformModuleVersionOutput]) {
	tool := mcp.Tool{
		Name:        "get_terraform_module_version",
		Description: "Get details of a specific Terraform module version by ID or TRN.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Terraform Module Version",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getTerraformModuleVersionInput) (*mcp.CallToolResult, getTerraformModuleVersionOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getTerraformModuleVersionOutput{}, err
		}

		version, err := client.TerraformModuleVersions().GetModuleVersion(ctx, &sdktypes.GetTerraformModuleVersionInput{
			ID: &input.ID,
		})
		if err != nil {
			return nil, getTerraformModuleVersionOutput{}, fmt.Errorf("failed to get module version: %w", err)
		}

		return nil, getTerraformModuleVersionOutput{
			ModuleVersion: toTerraformModuleVersion(version),
		}, nil
	}

	return tool, handler
}

// deleteTerraformModuleVersionInput is the input for deleting a Terraform module version.
type deleteTerraformModuleVersionInput struct {
	ID string `json:"id" jsonschema:"required,Module version ID or TRN (e.g. trn:terraform_module_version:group/module-name/system/1.0.0)"`
}

// deleteTerraformModuleVersionOutput is the output for deleting a Terraform module version.
type deleteTerraformModuleVersionOutput struct {
	Message string `json:"message" jsonschema:"Deletion confirmation message"`
	Success bool   `json:"success" jsonschema:"Whether deletion was successful"`
}

// DeleteTerraformModuleVersion returns an MCP tool for deleting a Terraform module version.
func deleteTerraformModuleVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[deleteTerraformModuleVersionInput, deleteTerraformModuleVersionOutput]) {
	tool := mcp.Tool{
		Name:        "delete_terraform_module_version",
		Description: "Delete a Terraform module version. Use with caution as this operation is irreversible.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Terraform Module Version",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input deleteTerraformModuleVersionInput) (*mcp.CallToolResult, deleteTerraformModuleVersionOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, deleteTerraformModuleVersionOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, input.ID, trn.ResourceTypeTerraformModuleVersion); err != nil {
			return nil, deleteTerraformModuleVersionOutput{}, err
		}

		if err := client.TerraformModuleVersions().DeleteModuleVersion(ctx, &sdktypes.DeleteTerraformModuleVersionInput{
			ID: input.ID,
		}); err != nil {
			return nil, deleteTerraformModuleVersionOutput{}, fmt.Errorf("failed to delete module version: %w", err)
		}

		return nil, deleteTerraformModuleVersionOutput{
			Message: fmt.Sprintf("Module version %s deleted successfully", input.ID),
			Success: true,
		}, nil
	}

	return tool, handler
}

// uploadModuleVersionInput is the input for uploading a module version package.
type uploadModuleVersionInput struct {
	ModuleID      string `json:"module_id" jsonschema:"required,Module ID or TRN (e.g. trn:terraform_module:group/module-name/system)"`
	Version       string `json:"version" jsonschema:"required,Version string (e.g. 1.0.0)"`
	DirectoryPath string `json:"directory_path" jsonschema:"required,Local path to the module directory to package and upload"`
}

// uploadModuleVersionOutput is the output for uploading a module version package.
type uploadModuleVersionOutput struct {
	ModuleVersion terraformModuleVersion `json:"module_version" jsonschema:"The uploaded module version"`
	Message       string                 `json:"message" jsonschema:"Upload confirmation message"`
}

// UploadModuleVersion returns an MCP tool for uploading a module version package.
func uploadModuleVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[uploadModuleVersionInput, uploadModuleVersionOutput]) {
	tool := mcp.Tool{
		Name:        "upload_module_version",
		Description: "Create and upload a new Terraform module version. Automatically packages the directory, calculates SHA sum, creates the version, and initiates the upload.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Upload Module Version",
			DestructiveHint: ptr.Bool(false),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input uploadModuleVersionInput) (*mcp.CallToolResult, uploadModuleVersionOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, uploadModuleVersionOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, input.ModuleID, trn.ResourceTypeTerraformModule); err != nil {
			return nil, uploadModuleVersionOutput{}, err
		}

		// Validate and sanitize the directory path
		directoryPath := filepath.Clean(input.DirectoryPath)

		// Check if directory exists and is a directory
		dirInfo, err := os.Stat(directoryPath)
		if err != nil {
			return nil, uploadModuleVersionOutput{}, fmt.Errorf("failed to stat directory: %w", err)
		}

		if !dirInfo.IsDir() {
			return nil, uploadModuleVersionOutput{}, fmt.Errorf("path is not a directory: %s", directoryPath)
		}

		// Create temporary slug file
		slugFile, err := os.CreateTemp("", "terraform-slug.tgz")
		if err != nil {
			return nil, uploadModuleVersionOutput{}, fmt.Errorf("failed to create temporary package file: %w", err)
		}
		defer os.Remove(slugFile.Name())

		// Create the module package
		moduleSlug, err := slug.NewSlug(directoryPath, slugFile.Name())
		if err != nil {
			return nil, uploadModuleVersionOutput{}, fmt.Errorf("failed to create module package: %w", err)
		}

		// Create module version with SHA sum
		version, err := client.TerraformModuleVersions().CreateModuleVersion(ctx, &sdktypes.CreateTerraformModuleVersionInput{
			ModulePath: trn.ToPath(input.ModuleID),
			Version:    input.Version,
			SHASum:     hex.EncodeToString(moduleSlug.SHASum),
		})
		if err != nil {
			return nil, uploadModuleVersionOutput{}, fmt.Errorf("failed to create module version: %w", err)
		}

		// Open the package for reading
		reader, err := moduleSlug.Open()
		if err != nil {
			// Delete the version we just created
			_ = client.TerraformModuleVersions().DeleteModuleVersion(ctx, &sdktypes.DeleteTerraformModuleVersionInput{ID: version.Metadata.ID})
			return nil, uploadModuleVersionOutput{}, fmt.Errorf("failed to open module package: %w", err)
		}
		defer reader.Close()

		// Upload the module version
		if err := client.TerraformModuleVersions().UploadModuleVersion(ctx, version.Metadata.ID, reader); err != nil {
			// Delete the version on upload failure
			_ = client.TerraformModuleVersions().DeleteModuleVersion(ctx, &sdktypes.DeleteTerraformModuleVersionInput{ID: version.Metadata.ID})
			return nil, uploadModuleVersionOutput{}, fmt.Errorf("failed to upload module version: %w", err)
		}

		return nil, uploadModuleVersionOutput{
			ModuleVersion: toTerraformModuleVersion(version),
			Message:       fmt.Sprintf("Module version %s upload initiated. Use get_terraform_module_version to check the status (pending, upload_in_progress, uploaded, or errored).", input.Version),
		}, nil
	}

	return tool, handler
}
