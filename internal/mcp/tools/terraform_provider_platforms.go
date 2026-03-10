package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
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
	CreatedBy         string `json:"created_by" jsonschema:"Username or service account that created this platform"`
}

// toTerraformProviderPlatform converts a proto TerraformProviderPlatform to the MCP output type.
func toTerraformProviderPlatform(p *pb.TerraformProviderPlatform) terraformProviderPlatform {
	return terraformProviderPlatform{
		ID:                p.Metadata.Id,
		TRN:               p.Metadata.Trn,
		ProviderVersionID: p.ProviderVersionId,
		OperatingSystem:   p.OperatingSystem,
		Architecture:      p.Architecture,
		SHASum:            p.ShaSum,
		Filename:          p.Filename,
		BinaryUploaded:    p.BinaryUploaded,
		CreatedBy:         p.CreatedBy,
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
func getTerraformProviderPlatform(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getTerraformProviderPlatformInput, *getTerraformProviderPlatformOutput]) {
	tool := mcp.Tool{
		Name:        "get_terraform_provider_platform",
		Description: "Get details of a specific Terraform provider platform by ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Terraform Provider Platform",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getTerraformProviderPlatformInput) (*mcp.CallToolResult, *getTerraformProviderPlatformOutput, error) {
		resp, err := tc.grpcClient.TerraformProvidersClient.GetTerraformProviderPlatformByID(ctx, &pb.GetTerraformProviderPlatformByIDRequest{
			Id: input.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get provider platform: %w", err)
		}

		return nil, &getTerraformProviderPlatformOutput{
			Platform: toTerraformProviderPlatform(resp),
		}, nil
	}

	return tool, handler
}
