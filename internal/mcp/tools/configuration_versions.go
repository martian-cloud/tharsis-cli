package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-slug"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// configurationVersion represents a Tharsis configuration version in MCP responses.
type configurationVersion struct {
	ID          string `json:"id" jsonschema:"The unique identifier of the configuration version"`
	Status      string `json:"status" jsonschema:"Status of the configuration version (e.g. pending uploaded errored)"`
	WorkspaceID string `json:"workspace_id" jsonschema:"The unique identifier of the workspace"`
	Speculative bool   `json:"speculative" jsonschema:"True if this is a speculative configuration version"`
	TRN         string `json:"trn" jsonschema:"Tharsis Resource Name"`
}

// toConfigurationVersion converts an SDK configuration version to MCP configuration version.
func toConfigurationVersion(cv *sdktypes.ConfigurationVersion) configurationVersion {
	return configurationVersion{
		ID:          cv.Metadata.ID,
		TRN:         cv.Metadata.TRN,
		Status:      cv.Status,
		WorkspaceID: cv.WorkspaceID,
		Speculative: cv.Speculative,
	}
}

// getConfigurationVersionInput is the input for the get_configuration_version tool.
type getConfigurationVersionInput struct {
	ID string `json:"id" jsonschema:"required,Configuration version ID or TRN (e.g. Ul8yZ... or trn:configuration_version:my-group/my-workspace/cv-id)"`
}

// getConfigurationVersionOutput is the output for the get_configuration_version tool.
type getConfigurationVersionOutput struct {
	ConfigurationVersion configurationVersion `json:"configuration_version" jsonschema:"The configuration version details"`
}

// GetConfigurationVersion returns an MCP tool for retrieving configuration version status.
func getConfigurationVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getConfigurationVersionInput, getConfigurationVersionOutput]) {
	tool := mcp.Tool{
		Name:        "get_configuration_version",
		Description: "Get the status of a configuration version. Use this to check if an upload has completed.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Configuration Version",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getConfigurationVersionInput) (*mcp.CallToolResult, getConfigurationVersionOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getConfigurationVersionOutput{}, fmt.Errorf("failed to get tharsis client: %w", err)
		}

		cv, err := client.ConfigurationVersions().GetConfigurationVersion(ctx, &sdktypes.GetConfigurationVersionInput{
			ID: input.ID,
		})
		if err != nil {
			return nil, getConfigurationVersionOutput{}, fmt.Errorf("failed to get configuration version %q: %w", input.ID, err)
		}

		return nil, getConfigurationVersionOutput{
			ConfigurationVersion: toConfigurationVersion(cv),
		}, nil
	}

	return tool, handler
}

// createConfigurationVersionInput is the input for the create_configuration_version tool.
type createConfigurationVersionInput struct {
	WorkspacePath string `json:"workspace_path" jsonschema:"required,Full path to the workspace (e.g. group/subgroup/workspace-name)"`
	DirectoryPath string `json:"directory_path" jsonschema:"required,Local directory path containing Terraform configuration files to upload"`
	Speculative   *bool  `json:"speculative,omitempty" jsonschema:"True for a speculative configuration version (plan-only)"`
}

// createConfigurationVersionOutput is the output for the create_configuration_version tool.
type createConfigurationVersionOutput struct {
	ConfigurationVersion configurationVersion `json:"configuration_version" jsonschema:"The created configuration version"`
}

// CreateConfigurationVersion returns an MCP tool for creating and uploading a configuration version.
func createConfigurationVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[createConfigurationVersionInput, createConfigurationVersionOutput]) {
	tool := mcp.Tool{
		Name:        "create_configuration_version",
		Description: "Create and upload a Terraform configuration version from a local directory. Use get_configuration_version to check upload status.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Configuration Version",
			DestructiveHint: ptr.Bool(false),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input createConfigurationVersionInput) (*mcp.CallToolResult, createConfigurationVersionOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, createConfigurationVersionOutput{}, fmt.Errorf("failed to get tharsis client: %w", err)
		}

		if err = tc.acl.Authorize(ctx, client, trn.ToTRN(input.WorkspacePath, trn.ResourceTypeWorkspace), trn.ResourceTypeWorkspace); err != nil {
			return nil, createConfigurationVersionOutput{}, err
		}

		cv, err := client.ConfigurationVersions().CreateConfigurationVersion(ctx, &sdktypes.CreateConfigurationVersionInput{
			WorkspacePath: input.WorkspacePath,
			Speculative:   input.Speculative,
		})
		if err != nil {
			return nil, createConfigurationVersionOutput{}, fmt.Errorf("failed to create configuration version in workspace %q: %w", input.WorkspacePath, err)
		}

		err = client.ConfigurationVersions().UploadConfigurationVersion(ctx, &sdktypes.UploadConfigurationVersionInput{
			WorkspacePath:          input.WorkspacePath,
			ConfigurationVersionID: cv.Metadata.ID,
			DirectoryPath:          input.DirectoryPath,
		})
		if err != nil {
			return nil, createConfigurationVersionOutput{}, fmt.Errorf("failed to upload configuration from %q: %w", input.DirectoryPath, err)
		}

		return nil, createConfigurationVersionOutput{
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
func downloadConfigurationVersion(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[downloadConfigurationVersionInput, downloadConfigurationVersionOutput]) {
	tool := mcp.Tool{
		Name:        "download_configuration_version",
		Description: "Download a configuration version to a temporary directory and return the path.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Download Configuration Version",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input downloadConfigurationVersionInput) (*mcp.CallToolResult, downloadConfigurationVersionOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, downloadConfigurationVersionOutput{}, fmt.Errorf("failed to get tharsis client: %w", err)
		}

		outputDir, err := os.MkdirTemp("", "config-version-*")
		if err != nil {
			return nil, downloadConfigurationVersionOutput{}, fmt.Errorf("failed to create temp directory for configuration version %q: %w", input.ID, err)
		}

		tarFile, err := os.CreateTemp("", "config-version-*.tar.gz")
		if err != nil {
			return nil, downloadConfigurationVersionOutput{}, fmt.Errorf("failed to create temp file for configuration version %q: %w", input.ID, err)
		}
		defer os.Remove(tarFile.Name())
		defer tarFile.Close()

		err = client.ConfigurationVersions().DownloadConfigurationVersion(ctx, &sdktypes.GetConfigurationVersionInput{
			ID: input.ID,
		}, tarFile)
		if err != nil {
			return nil, downloadConfigurationVersionOutput{}, fmt.Errorf("failed to download configuration version %q: %w", input.ID, err)
		}

		if _, err := tarFile.Seek(0, 0); err != nil {
			return nil, downloadConfigurationVersionOutput{}, fmt.Errorf("failed to seek tar file for configuration version %q: %w", input.ID, err)
		}

		if err := slug.Unpack(tarFile, outputDir); err != nil {
			return nil, downloadConfigurationVersionOutput{}, fmt.Errorf("failed to extract configuration version %q: %w", input.ID, err)
		}

		return nil, downloadConfigurationVersionOutput{
			ConfigurationVersionID: input.ID,
			OutputPath:             outputDir,
		}, nil
	}

	return tool, handler
}
