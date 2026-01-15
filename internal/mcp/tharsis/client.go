// Package tharsis provides a wrapper around the Tharsis SDK client.
package tharsis

import (
	sdk "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
)

//go:generate go tool mockery --name Client --inpackage --case underscore
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name Workspaces --output . --outpkg tharsis --case underscore --filename mock_workspaces.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name Run --output . --outpkg tharsis --case underscore --filename mock_run.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name Group --output . --outpkg tharsis --case underscore --filename mock_group.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name Job --output . --outpkg tharsis --case underscore --filename mock_job.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name ManagedIdentity --output . --outpkg tharsis --case underscore --filename mock_managed_identity.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name ConfigurationVersion --output . --outpkg tharsis --case underscore --filename mock_configuration_version.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name Variable --output . --outpkg tharsis --case underscore --filename mock_variable.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name StateVersion --output . --outpkg tharsis --case underscore --filename mock_state_version.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name Me --output . --outpkg tharsis --case underscore --filename mock_me.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name TerraformModule --output . --outpkg tharsis --case underscore --filename mock_terraform_module.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name TerraformModuleVersion --output . --outpkg tharsis --case underscore --filename mock_terraform_module_version.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name TerraformProvider --output . --outpkg tharsis --case underscore --filename mock_terraform_provider.go
//go:generate go tool mockery --srcpkg gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg --name TerraformProviderPlatform --output . --outpkg tharsis --case underscore --filename mock_terraform_provider_platform.go

// Client wraps the Tharsis SDK client.
type Client interface {
	Workspaces() sdk.Workspaces
	Runs() sdk.Run
	Groups() sdk.Group
	Jobs() sdk.Job
	ManagedIdentities() sdk.ManagedIdentity
	ConfigurationVersions() sdk.ConfigurationVersion
	Variables() sdk.Variable
	StateVersions() sdk.StateVersion
	Me() sdk.Me
	TerraformModules() sdk.TerraformModule
	TerraformModuleVersions() sdk.TerraformModuleVersion
	TerraformProviders() sdk.TerraformProvider
	TerraformProviderPlatforms() sdk.TerraformProviderPlatform
}

type wrappedClient struct {
	client *sdk.Client
}

// NewClient wraps a Tharsis SDK client.
func NewClient(c *sdk.Client) Client {
	return &wrappedClient{client: c}
}

func (c *wrappedClient) Workspaces() sdk.Workspaces {
	return c.client.Workspaces
}

func (c *wrappedClient) Runs() sdk.Run {
	return c.client.Run
}

func (c *wrappedClient) Groups() sdk.Group {
	return c.client.Group
}

func (c *wrappedClient) Jobs() sdk.Job {
	return c.client.Job
}

func (c *wrappedClient) ManagedIdentities() sdk.ManagedIdentity {
	return c.client.ManagedIdentity
}

func (c *wrappedClient) ConfigurationVersions() sdk.ConfigurationVersion {
	return c.client.ConfigurationVersion
}

func (c *wrappedClient) Variables() sdk.Variable {
	return c.client.Variable
}

func (c *wrappedClient) StateVersions() sdk.StateVersion {
	return c.client.StateVersion
}

func (c *wrappedClient) Me() sdk.Me {
	return c.client.Me
}

func (c *wrappedClient) TerraformModules() sdk.TerraformModule {
	return c.client.TerraformModule
}

func (c *wrappedClient) TerraformModuleVersions() sdk.TerraformModuleVersion {
	return c.client.TerraformModuleVersion
}

func (c *wrappedClient) TerraformProviders() sdk.TerraformProvider {
	return c.client.TerraformProvider
}

func (c *wrappedClient) TerraformProviderPlatforms() sdk.TerraformProviderPlatform {
	return c.client.TerraformProviderPlatform
}
