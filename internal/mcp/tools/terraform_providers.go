package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// terraformProvider is the output type for Terraform providers.
type terraformProvider struct {
	ID            string `json:"id" jsonschema:"Unique identifier for the provider"`
	TRN           string `json:"trn" jsonschema:"Tharsis Resource Name (TRN) for the provider"`
	Name          string `json:"name" jsonschema:"Provider name"`
	GroupID       string `json:"group_id" jsonschema:"ID of the group containing this provider"`
	RepositoryURL string `json:"repository_url" jsonschema:"URL to the provider's source repository"`
	Private       bool   `json:"private" jsonschema:"Whether the provider is private"`
	CreatedBy     string `json:"created_by" jsonschema:"Username or service account that created this provider"`
}

// toTerraformProvider converts a proto TerraformProvider to the MCP output type.
func toTerraformProvider(p *pb.TerraformProvider) terraformProvider {
	return terraformProvider{
		ID:            p.Metadata.Id,
		TRN:           p.Metadata.Trn,
		Name:          p.Name,
		GroupID:       p.GroupId,
		RepositoryURL: p.RepositoryUrl,
		Private:       p.Private,
		CreatedBy:     p.CreatedBy,
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
func getTerraformProvider(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getTerraformProviderInput, *getTerraformProviderOutput]) {
	tool := mcp.Tool{
		Name:        "get_terraform_provider",
		Description: "Get details of a specific Terraform provider by ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Terraform Provider",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getTerraformProviderInput) (*mcp.CallToolResult, *getTerraformProviderOutput, error) {
		resp, err := tc.grpcClient.TerraformProvidersClient.GetTerraformProviderByID(ctx, &pb.GetTerraformProviderByIDRequest{
			Id: input.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get provider: %w", err)
		}

		return nil, &getTerraformProviderOutput{
			Provider: toTerraformProvider(resp),
		}, nil
	}

	return tool, handler
}
