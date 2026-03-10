package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-slug"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// configurationVersion represents a Tharsis configuration version in MCP responses.
type configurationVersion struct {
	ID          string `json:"id" jsonschema:"The unique identifier of the configuration version"`
	Status      string `json:"status" jsonschema:"Status of the configuration version (e.g. pending uploaded errored)"`
	WorkspaceID string `json:"workspace_id" jsonschema:"The unique identifier of the workspace"`
	Speculative bool   `json:"speculative" jsonschema:"True if this is a speculative configuration version"`
	TRN         string `json:"trn" jsonschema:"Tharsis Resource Name"`
}

// toConfigurationVersion converts a proto configuration version to MCP configuration version.
func toConfigurationVersion(cv *pb.ConfigurationVersion) *configurationVersion {
	return &configurationVersion{
		ID:          cv.Metadata.Id,
		TRN:         cv.Metadata.Trn,
		Status:      cv.Status,
		WorkspaceID: cv.WorkspaceId,
		Speculative: cv.Speculative,
	}
}

// getConfigurationVersionInput is the input for the get_configuration_version tool.
type getConfigurationVersionInput struct {
	ID string `json:"id" jsonschema:"required,Configuration version ID or TRN (e.g. Ul8yZ... or trn:configuration_version:my-group/my-workspace/cv-id)"`
}

// getConfigurationVersionOutput is the output for the get_configuration_version tool.
type getConfigurationVersionOutput struct {
	ConfigurationVersion *configurationVersion `json:"configuration_version,omitempty" jsonschema:"The configuration version details"`
}

// GetConfigurationVersion returns an MCP tool for retrieving configuration version status.
func getConfigurationVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getConfigurationVersionInput, *getConfigurationVersionOutput]) {
	tool := mcp.Tool{
		Name:        "get_configuration_version",
		Description: "Get the status of a configuration version. Use this to check if an upload has completed.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Configuration Version",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getConfigurationVersionInput) (*mcp.CallToolResult, *getConfigurationVersionOutput, error) {
		resp, err := tc.grpcClient.ConfigurationVersionsClient.GetConfigurationVersionByID(ctx, &pb.GetConfigurationVersionByIDRequest{
			Id: input.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get configuration version %q: %w", input.ID, err)
		}

		return nil, &getConfigurationVersionOutput{
			ConfigurationVersion: toConfigurationVersion(resp),
		}, nil
	}

	return tool, handler
}

// createConfigurationVersionInput is the input for the create_configuration_version tool.
type createConfigurationVersionInput struct {
	WorkspaceID   string `json:"workspace_id" jsonschema:"required,Workspace ID or TRN (e.g. Ul8yZ... or trn:workspace:group/subgroup/workspace-name)"`
	DirectoryPath string `json:"directory_path" jsonschema:"required,Local directory path containing Terraform configuration files to upload"`
	Speculative   bool   `json:"speculative,omitempty" jsonschema:"True for a speculative configuration version (plan-only)"`
}

// createConfigurationVersionOutput is the output for the create_configuration_version tool.
type createConfigurationVersionOutput struct {
	ConfigurationVersion *configurationVersion `json:"configuration_version,omitempty" jsonschema:"The created configuration version"`
}

// CreateConfigurationVersion returns an MCP tool for creating and uploading a configuration version.
func createConfigurationVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*createConfigurationVersionInput, *createConfigurationVersionOutput]) {
	tool := mcp.Tool{
		Name:        "create_configuration_version",
		Description: "Create and upload a Terraform configuration version from a local directory. Use get_configuration_version to check upload status.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Configuration Version",
			DestructiveHint: ptr.Bool(false),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *createConfigurationVersionInput) (*mcp.CallToolResult, *createConfigurationVersionOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.WorkspaceID, trn.ResourceTypeWorkspace); err != nil {
			return nil, nil, err
		}

		cv, err := tc.grpcClient.ConfigurationVersionsClient.CreateConfigurationVersion(ctx, &pb.CreateConfigurationVersionRequest{
			WorkspaceId: input.WorkspaceID,
			Speculative: input.Speculative,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create configuration version in workspace %q: %w", input.WorkspaceID, err)
		}

		if err := tc.tfeClient.UploadConfigurationVersion(ctx, &tfe.UploadConfigurationVersionInput{
			WorkspaceID:     input.WorkspaceID,
			ConfigVersionID: cv.Metadata.Id,
			DirectoryPath:   input.DirectoryPath,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to upload configuration: %w", err)
		}

		return nil, &createConfigurationVersionOutput{
			ConfigurationVersion: toConfigurationVersion(cv),
		}, nil
	}

	return tool, handler
}

// downloadConfigurationVersionInput is the input for the download_configuration_version tool.
type downloadConfigurationVersionInput struct {
	ID string `json:"id" jsonschema:"required,Configuration version ID or TRN (e.g. Ul8yZ... or trn:configuration_version:my-group/my-workspace/cv-id)"`
}

// downloadConfigurationVersionOutput is the output for the download_configuration_version tool.
type downloadConfigurationVersionOutput struct {
	ConfigurationVersionID string `json:"configuration_version_id" jsonschema:"The unique identifier of the downloaded configuration version"`
	OutputPath             string `json:"output_path" jsonschema:"Temporary directory path where files were downloaded"`
}

// DownloadConfigurationVersion returns an MCP tool for downloading a configuration version.
func downloadConfigurationVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*downloadConfigurationVersionInput, *downloadConfigurationVersionOutput]) {
	tool := mcp.Tool{
		Name:        "download_configuration_version",
		Description: "Download a configuration version to a temporary directory and return the path.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Download Configuration Version",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *downloadConfigurationVersionInput) (*mcp.CallToolResult, *downloadConfigurationVersionOutput, error) {
		outputDir, err := os.MkdirTemp("", "config-version-*")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create temp directory for configuration version %q: %w", input.ID, err)
		}

		tarFile, err := os.CreateTemp("", "config-version-*.tar.gz")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create temp file for configuration version %q: %w", input.ID, err)
		}
		defer os.Remove(tarFile.Name())
		defer tarFile.Close()

		if err := tc.tfeClient.DownloadConfigurationVersion(ctx, &tfe.DownloadConfigurationVersionInput{
			ConfigVersionID: input.ID,
			Writer:          tarFile,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to download configuration version %q: %w", input.ID, err)
		}

		if _, err := tarFile.Seek(0, 0); err != nil {
			return nil, nil, fmt.Errorf("failed to seek tar file for configuration version %q: %w", input.ID, err)
		}

		if err := slug.Unpack(tarFile, outputDir); err != nil {
			return nil, nil, fmt.Errorf("failed to extract configuration version %q: %w", input.ID, err)
		}

		return nil, &downloadConfigurationVersionOutput{
			ConfigurationVersionID: input.ID,
			OutputPath:             outputDir,
		}, nil
	}

	return tool, handler
}
