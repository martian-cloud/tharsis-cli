package tools

import (
	"fmt"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/mcp/tools"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/prompts"
)

// Toolset metadata for Tharsis CLI
var (
	ToolsetMetadataAuth = tools.ToolsetMetadata{
		Name:        "auth",
		Description: "Authentication tools such as SSO login",
	}
	ToolsetMetadataRuns = tools.ToolsetMetadata{
		Name:        "run",
		Description: "Run management tools for creating and managing Terraform runs",
	}
	ToolsetMetadataJobs = tools.ToolsetMetadata{
		Name:        "job",
		Description: "Job tools for retrieving logs and job information",
	}
	ToolsetMetadataConfigurationVersions = tools.ToolsetMetadata{
		Name:        "configuration_version",
		Description: "Configuration version tools for uploading and downloading Terraform configurations",
	}
	ToolsetMetadataWorkspaces = tools.ToolsetMetadata{
		Name:        "workspace",
		Description: "Workspace management tools for creating, updating, and deleting workspaces",
	}
	ToolsetMetadataGroups = tools.ToolsetMetadata{
		Name:        "group",
		Description: "Group management tools for creating, updating, and deleting groups",
	}
	ToolsetMetadataVariables = tools.ToolsetMetadata{
		Name:        "variable",
		Description: "Tools for managing Terraform and environment variables on workspaces",
	}
	ToolsetMetadataManagedIdentities = tools.ToolsetMetadata{
		Name:        "managed_identity",
		Description: "Tools for assigning and unassigning managed identities to workspaces",
	}
	ToolsetMetadataDocumentation = tools.ToolsetMetadata{
		Name:        "documentation",
		Description: "Tools for searching and retrieving Tharsis documentation",
	}
	ToolsetMetadataTerraformModules = tools.ToolsetMetadata{
		Name:        "terraform_module",
		Description: "Tools for managing Terraform modules in the Tharsis registry",
	}
	ToolsetMetadataTerraformModuleVersions = tools.ToolsetMetadata{
		Name:        "terraform_module_version",
		Description: "Tools for managing Terraform module versions in the Tharsis registry",
	}
	ToolsetMetadataTerraformProviders = tools.ToolsetMetadata{
		Name:        "terraform_provider",
		Description: "Tools for managing Terraform providers in the Tharsis registry",
	}
	ToolsetMetadataTerraformProviderPlatforms = tools.ToolsetMetadata{
		Name:        "terraform_provider_platform",
		Description: "Tools for managing Terraform provider platforms in the Tharsis registry",
	}
)

// AvailableToolsets returns a list of all available toolsets.
func AvailableToolsets() []string {
	return []string{
		ToolsetMetadataAuth.Name,
		ToolsetMetadataRuns.Name,
		ToolsetMetadataJobs.Name,
		ToolsetMetadataConfigurationVersions.Name,
		ToolsetMetadataWorkspaces.Name,
		ToolsetMetadataGroups.Name,
		ToolsetMetadataVariables.Name,
		ToolsetMetadataManagedIdentities.Name,
		ToolsetMetadataDocumentation.Name,
		ToolsetMetadataTerraformModules.Name,
		ToolsetMetadataTerraformModuleVersions.Name,
		ToolsetMetadataTerraformProviders.Name,
		ToolsetMetadataTerraformProviderPlatforms.Name,
	}
}

// BuildToolsetGroup creates and configures all toolsets for the CLI MCP server.
func BuildToolsetGroup(readOnly bool, tc *ToolContext) (*tools.ToolsetGroup, error) {
	group := tools.NewToolsetGroup(readOnly)

	// Authentication tools
	auth := tools.NewToolset(ToolsetMetadataAuth).
		AddReadTools(
			tools.NewServerTool(loginWithSSO(tc)),
			tools.NewServerTool(getConnectionInfo(tc)),
		)

	// Run tools
	runs := tools.NewToolset(ToolsetMetadataRuns).
		AddReadTools(
			tools.NewServerTool(listRuns(tc)),
			tools.NewServerTool(getRun(tc)),
		).
		AddWriteTools(
			tools.NewServerTool(createRun(tc)),
			tools.NewServerTool(applyRun(tc)),
			tools.NewServerTool(cancelRun(tc)),
		).
		AddPrompts(
			// Deployment workflows
			tools.NewServerPrompt(prompts.ApplyRunPrompt()),
			tools.NewServerPrompt(prompts.PlanRunPrompt()),
			// Troubleshooting workflows
			tools.NewServerPrompt(prompts.DiagnoseRunPrompt()),
			tools.NewServerPrompt(prompts.FixRunPrompt()),
		)

	// Job tools
	jobs := tools.NewToolset(ToolsetMetadataJobs).
		AddReadTools(
			tools.NewServerTool(getJobLogs(tc)),
		)

	// Configuration version tools
	configVersions := tools.NewToolset(ToolsetMetadataConfigurationVersions).
		AddReadTools(
			tools.NewServerTool(getConfigurationVersion(tc)),
			tools.NewServerTool(downloadConfigurationVersion(tc)),
		).
		AddWriteTools(
			tools.NewServerTool(createConfigurationVersion(tc)),
		)

	// Workspace tools
	workspaces := tools.NewToolset(ToolsetMetadataWorkspaces).
		AddReadTools(
			tools.NewServerTool(listWorkspaces(tc)),
			tools.NewServerTool(getWorkspace(tc)),
			tools.NewServerTool(getWorkspaceOutputs(tc)),
		).
		AddWriteTools(
			tools.NewServerTool(createWorkspace(tc)),
			tools.NewServerTool(updateWorkspace(tc)),
			tools.NewServerTool(deleteWorkspace(tc)),
		).
		AddPrompts(
			tools.NewServerPrompt(prompts.SetupWorkspacePrompt()),
			tools.NewServerPrompt(prompts.InspectWorkspacePrompt()),
		)

	// Group tools
	groups := tools.NewToolset(ToolsetMetadataGroups).
		AddReadTools(
			tools.NewServerTool(listGroups(tc)),
			tools.NewServerTool(getGroup(tc)),
		).
		AddWriteTools(
			tools.NewServerTool(createGroup(tc)),
			tools.NewServerTool(updateGroup(tc)),
			tools.NewServerTool(deleteGroup(tc)),
		)

	// Variable tools
	variables := tools.NewToolset(ToolsetMetadataVariables).
		AddWriteTools(
			tools.NewServerTool(setVariable(tc)),
			tools.NewServerTool(setTerraformVariablesFromFile(tc)),
			tools.NewServerTool(setEnvironmentVariablesFromFile(tc)),
			tools.NewServerTool(deleteVariable(tc)),
		)

	// Managed identity tools
	managedIdentities := tools.NewToolset(ToolsetMetadataManagedIdentities).
		AddWriteTools(
			tools.NewServerTool(assignManagedIdentity(tc)),
			tools.NewServerTool(unassignManagedIdentity(tc)),
		)

	// Documentation tools
	docService, err := tools.NewDocumentSearchService(tc.httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create documentation search service: %w", err)
	}

	documentation := tools.NewToolset(ToolsetMetadataDocumentation).
		AddReadTools(
			tools.NewServerTool(tools.SearchDocumentation(docService)),
			tools.NewServerTool(tools.GetDocumentationPage(docService)),
		)

	// Terraform module tools
	terraformModules := tools.NewToolset(ToolsetMetadataTerraformModules).
		AddReadTools(
			tools.NewServerTool(listTerraformModules(tc)),
			tools.NewServerTool(getTerraformModule(tc)),
		).
		AddWriteTools(
			tools.NewServerTool(createTerraformModule(tc)),
			tools.NewServerTool(updateTerraformModule(tc)),
			tools.NewServerTool(deleteTerraformModule(tc)),
		).
		AddPrompts(
			tools.NewServerPrompt(prompts.PublishModulePrompt()),
		)

	// Terraform module version tools
	terraformModuleVersions := tools.NewToolset(ToolsetMetadataTerraformModuleVersions).
		AddReadTools(
			tools.NewServerTool(listTerraformModuleVersions(tc)),
			tools.NewServerTool(getTerraformModuleVersion(tc)),
		).
		AddWriteTools(
			tools.NewServerTool(uploadModuleVersion(tc)),
			tools.NewServerTool(deleteTerraformModuleVersion(tc)),
		)

	// Terraform provider tools
	terraformProviders := tools.NewToolset(ToolsetMetadataTerraformProviders).
		AddReadTools(
			tools.NewServerTool(getTerraformProvider(tc)),
		)

	// Terraform provider platform tools
	terraformProviderPlatforms := tools.NewToolset(ToolsetMetadataTerraformProviderPlatforms).
		AddReadTools(
			tools.NewServerTool(getTerraformProviderPlatform(tc)),
		)

	group.AddToolset(auth)
	group.AddToolset(runs)
	group.AddToolset(jobs)
	group.AddToolset(configVersions)
	group.AddToolset(workspaces)
	group.AddToolset(groups)
	group.AddToolset(variables)
	group.AddToolset(managedIdentities)
	group.AddToolset(documentation)
	group.AddToolset(terraformModules)
	group.AddToolset(terraformModuleVersions)
	group.AddToolset(terraformProviders)
	group.AddToolset(terraformProviderPlatforms)

	return group, nil
}
