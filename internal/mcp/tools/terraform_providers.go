package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// terraformProvider is the output type for Terraform providers.
type terraformProvider struct {
	ID                string `json:"id" jsonschema:"Unique identifier for the provider"`
	TRN               string `json:"trn" jsonschema:"Tharsis Resource Name (TRN) for the provider"`
	Name              string `json:"name" jsonschema:"Provider name"`
	GroupPath         string `json:"group_path" jsonschema:"Path to the group containing this provider"`
	RegistryNamespace string `json:"registry_namespace" jsonschema:"Registry namespace for the provider"`
	RepositoryURL     string `json:"repository_url" jsonschema:"URL to the provider's source repository"`
	Private           bool   `json:"private" jsonschema:"Whether the provider is private"`
}

// toTerraformProvider converts an SDK TerraformProvider to the MCP output type.
func toTerraformProvider(p *sdktypes.TerraformProvider) terraformProvider {
	return terraformProvider{
		ID:                p.Metadata.ID,
		TRN:               p.Metadata.TRN,
		Name:              p.Name,
		GroupPath:         p.GroupPath,
		RegistryNamespace: p.RegistryNamespace,
		RepositoryURL:     p.RepositoryURL,
		Private:           p.Private,
	}
}

// getTerraformProviderInput is the input for getting a Terraform provider.
type getTerraformProviderInput struct {
	ID string `json:"id" jsonschema:"required,Provider ID or TRN (e.g. Ul8yZ... or trn:terraform_provider:group/provider-name)"`
}

// getTerraformProviderOutput is the output for getting a Terraform provider.
type getTerraformProviderOutput struct {
	Provider terraformProvider `json:"provider" jsonschema:"The Terraform provider"`
}

// GetTerraformProvider returns an MCP tool for getting a Terraform provider.
func getTerraformProvider(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getTerraformProviderInput, getTerraformProviderOutput]) {
	tool := mcp.Tool{
		Name:        "get_terraform_provider",
		Description: "Get details of a specific Terraform provider by ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Terraform Provider",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getTerraformProviderInput) (*mcp.CallToolResult, getTerraformProviderOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getTerraformProviderOutput{}, err
		}

		provider, err := client.TerraformProviders().GetProvider(ctx, &sdktypes.GetTerraformProviderInput{
			ID: input.ID,
		})
		if err != nil {
			return nil, getTerraformProviderOutput{}, fmt.Errorf("failed to get provider: %w", err)
		}

		return nil, getTerraformProviderOutput{
			Provider: toTerraformProvider(provider),
		}, nil
	}

	return tool, handler
}
