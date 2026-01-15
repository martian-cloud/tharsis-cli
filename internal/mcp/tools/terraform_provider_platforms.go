package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// terraformProviderPlatform is the output type for Terraform provider platforms.
type terraformProviderPlatform struct {
	ID                string `json:"id" jsonschema:"Unique identifier for the provider platform"`
	TRN               string `json:"trn" jsonschema:"Tharsis Resource Name (TRN) for the provider platform"`
	ProviderVersionID string `json:"provider_version_id" jsonschema:"ID of the provider version this platform belongs to"`
	OperatingSystem   string `json:"operating_system" jsonschema:"Operating system for this platform"`
	Architecture      string `json:"architecture" jsonschema:"Architecture for this platform"`
	SHASum            string `json:"sha_sum" jsonschema:"SHA256 checksum of the provider binary"`
	Filename          string `json:"filename" jsonschema:"Filename of the provider binary"`
	BinaryUploaded    bool   `json:"binary_uploaded" jsonschema:"Whether the binary has been uploaded"`
}

// toTerraformProviderPlatform converts an SDK TerraformProviderPlatform to the MCP output type.
func toTerraformProviderPlatform(p *sdktypes.TerraformProviderPlatform) terraformProviderPlatform {
	return terraformProviderPlatform{
		ID:                p.Metadata.ID,
		TRN:               p.Metadata.TRN,
		ProviderVersionID: p.ProviderVersionID,
		OperatingSystem:   p.OperatingSystem,
		Architecture:      p.Architecture,
		SHASum:            p.SHASum,
		Filename:          p.Filename,
		BinaryUploaded:    p.BinaryUploaded,
	}
}

// getTerraformProviderPlatformInput is the input for getting a Terraform provider platform.
type getTerraformProviderPlatformInput struct {
	ID string `json:"id" jsonschema:"required,Provider platform ID or TRN (e.g. Ul8yZ... or trn:terraform_provider_platform:group/provider-name/version/platform-id)"`
}

// getTerraformProviderPlatformOutput is the output for getting a Terraform provider platform.
type getTerraformProviderPlatformOutput struct {
	Platform terraformProviderPlatform `json:"platform" jsonschema:"The Terraform provider platform"`
}

// GetTerraformProviderPlatform returns an MCP tool for getting a Terraform provider platform.
func getTerraformProviderPlatform(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getTerraformProviderPlatformInput, getTerraformProviderPlatformOutput]) {
	tool := mcp.Tool{
		Name:        "get_terraform_provider_platform",
		Description: "Get details of a specific Terraform provider platform by ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Terraform Provider Platform",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getTerraformProviderPlatformInput) (*mcp.CallToolResult, getTerraformProviderPlatformOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getTerraformProviderPlatformOutput{}, err
		}

		platform, err := client.TerraformProviderPlatforms().GetProviderPlatform(ctx, &sdktypes.GetTerraformProviderPlatformInput{
			ID: input.ID,
		})
		if err != nil {
			return nil, getTerraformProviderPlatformOutput{}, fmt.Errorf("failed to get provider platform: %w", err)
		}

		return nil, getTerraformProviderPlatformOutput{
			Platform: toTerraformProviderPlatform(platform),
		}, nil
	}

	return tool, handler
}
