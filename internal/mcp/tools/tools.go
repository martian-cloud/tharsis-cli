// Package tools provides MCP tool implementations for Tharsis CLI.
package tools

//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name CallerClient --output mocks --outpkg mocks --filename mock_caller_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name ConfigurationVersionsClient --output mocks --outpkg mocks --filename mock_configuration_versions_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name GroupsClient --output mocks --outpkg mocks --filename mock_groups_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name JobsClient --output mocks --outpkg mocks --filename mock_jobs_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name ManagedIdentitiesClient --output mocks --outpkg mocks --filename mock_managed_identities_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name NamespaceVariablesClient --output mocks --outpkg mocks --filename mock_namespace_variables_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name RunsClient --output mocks --outpkg mocks --filename mock_runs_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name StateVersionsClient --output mocks --outpkg mocks --filename mock_state_versions_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name TerraformModulesClient --output mocks --outpkg mocks --filename mock_terraform_modules_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name TerraformModuleVersionsClient --output mocks --outpkg mocks --filename mock_terraform_module_versions_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name TerraformProvidersClient --output mocks --outpkg mocks --filename mock_terraform_providers_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name TerraformProviderPlatformsClient --output mocks --outpkg mocks --filename mock_terraform_provider_platforms_client.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen --name WorkspacesClient --output mocks --outpkg mocks --filename mock_workspaces_client.go

import (
	"fmt"
	"net/http"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
)

// ToolContext holds dependencies for tool execution.
type ToolContext struct {
	tharsisURL  string
	profileName string
	grpcClient  *client.Client
	tfeClient   tfe.RESTClient
	httpClient  *http.Client
	acl         acl.Checker
}

// ToolContextOption configures a ToolContext.
type ToolContextOption func(*ToolContext) error

// WithACLPatterns sets ACL patterns for the tool context.
func WithACLPatterns(patterns string) ToolContextOption {
	return func(tc *ToolContext) error {
		checker, err := acl.NewChecker(patterns)
		if err != nil {
			return fmt.Errorf("failed to initialize ACL checker: %w", err)
		}
		tc.acl = checker
		return nil
	}
}

// NewToolContext creates a new tool context.
func NewToolContext(
	tharsisURL,
	profileName string,
	httpClient *http.Client,
	client *client.Client,
	restClient tfe.RESTClient,
	opts ...ToolContextOption,
) (*ToolContext, error) {
	tc := &ToolContext{
		tharsisURL:  tharsisURL,
		profileName: profileName,
		grpcClient:  client,
		tfeClient:   restClient,
		httpClient:  httpClient,
	}

	for _, opt := range opts {
		if err := opt(tc); err != nil {
			return nil, err
		}
	}

	if tc.acl == nil {
		checker, err := acl.NewChecker("")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize ACL checker: %w", err)
		}
		tc.acl = checker
	}

	return tc, nil
}
