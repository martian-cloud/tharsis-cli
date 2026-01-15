// Package prompts provides MCP workflow prompts for Tharsis CLI.
package prompts

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/mcp/prompts"
)

// DiagnoseRunPrompt returns a prompt for diagnosing a run.
func DiagnoseRunPrompt() (mcp.Prompt, mcp.PromptHandler) {
	return prompts.NewWorkflowPrompt(
		"diagnose_run",
		"Diagnose run {run_id}",
	).
		AddRequiredArgument("run_id", "The ID of the run to diagnose").
		AddStep("get_run", "retrieve run details and error messages for {run_id}").
		AddStep("get_job_logs", "retrieve the full logs to see detailed error context").
		AddStep("", "analyze the run and explain any errors or issues to the user").
		Build()
}

// FixRunPrompt returns a prompt for fixing a failed run.
func FixRunPrompt() (mcp.Prompt, mcp.PromptHandler) {
	return prompts.NewWorkflowPrompt(
		"fix_run",
		"Fix the failed run {run_id}",
	).
		AddRequiredArgument("run_id", "The ID of the failed run to fix").
		AddStep("get_run", "retrieve run details to determine if it was created from configuration version or module source").
		AddStep("download_configuration_version", "if configuration version: download the configuration using the version ID from the run").
		AddStep("", "if configuration version: fix the Terraform files in the downloaded directory").
		AddStep("create_configuration_version", "if configuration version: upload the fixed configuration to the workspace").
		AddStep("", "if module source: suggest fixes like updating module version, changing variables, or providing local config to override").
		AddStep("create_run", "create a new run with the fix applied").
		AddStep("get_run", "poll to verify the fix worked").
		Build()
}

// PlanRunPrompt returns a prompt for creating a speculative plan.
func PlanRunPrompt() (mcp.Prompt, mcp.PromptHandler) {
	return prompts.NewWorkflowPrompt(
		"plan_run",
		"Create a speculative plan to preview changes for workspace {workspace_path}",
	).
		AddRequiredArgument("workspace_path", "Full path to the workspace (e.g., group/subgroup/workspace-name)").
		AddStep("", "ask user if they want to deploy from a local directory (configuration version) or a module source").
		AddStep("create_configuration_version", "if using local directory: upload the Terraform files as speculative").
		AddStep("get_configuration_version", "if using local directory: poll until upload status is 'uploaded'").
		AddStep("create_run", "create a speculative run (set speculative=true) - use configuration version ID if local, or module_source/module_version if module").
		AddStep("get_run", "poll until run status becomes 'planned' or 'errored'").
		AddStep("get_job_logs", "retrieve the plan logs with job_type='plan' to see what changes would be made").
		Build()
}

// ApplyRunPrompt returns a prompt for planning and applying changes.
func ApplyRunPrompt() (mcp.Prompt, mcp.PromptHandler) {
	return prompts.NewWorkflowPrompt(
		"apply_run",
		"Plan and apply Terraform changes to workspace {workspace_path}",
	).
		AddRequiredArgument("workspace_path", "Full path to the workspace (e.g., group/subgroup/workspace-name)").
		AddStep("", "ask user if they want to deploy from a local directory (configuration version) or a module source").
		AddStep("create_configuration_version", "if using local directory: upload the Terraform files").
		AddStep("get_configuration_version", "if using local directory: poll until upload status is 'uploaded'").
		AddStep("create_run", "create a run - use configuration version ID if local, or module_source/module_version if module").
		AddStep("get_run", "poll until run status becomes 'planned'").
		AddStep("get_job_logs", "retrieve the plan logs with job_type='plan' to show what changes will be made").
		AddStep("", "ask user for explicit confirmation before applying").
		AddStep("apply_run", "apply the changes after user confirms").
		AddStep("get_run", "poll until run status becomes 'applied' or 'errored'").
		AddStep("get_job_logs", "retrieve the apply logs with job_type='apply' to show apply results").
		Build()
}

// SetupWorkspacePrompt returns a prompt for setting up a workspace.
func SetupWorkspacePrompt() (mcp.Prompt, mcp.PromptHandler) {
	return prompts.NewWorkflowPrompt(
		"setup_workspace",
		"Create and configure workspace {workspace_path}",
	).
		AddRequiredArgument("workspace_path", "Full path for the new workspace (e.g., group/subgroup/workspace-name)").
		AddStep("create_workspace", "create the workspace at {workspace_path}").
		AddStep("", "ask user if they want to set variables individually or from a file").
		AddStep("set_variable", "set individual variables if user provides them one by one").
		AddStep("set_terraform_variables_from_file", "set Terraform variables from .tfvars file if user provides a file path").
		AddStep("set_environment_variables_from_file", "set environment variables from file if user provides a file path").
		AddStep("", "ask user for any managed identities to assign").
		AddStep("assign_managed_identity", "assign each managed identity to the workspace if provided").
		Build()
}

// PublishModulePrompt returns a prompt for publishing a module.
func PublishModulePrompt() (mcp.Prompt, mcp.PromptHandler) {
	return prompts.NewWorkflowPrompt(
		"publish_module",
		"Publish version {version} of a Terraform module from {directory_path} to {module_path} in the registry, creating the module if needed",
	).
		AddRequiredArgument("module_path", "Full path to the module in the registry (e.g., group/module-name/system)").
		AddRequiredArgument("version", "Version to publish (e.g., 1.0.0)").
		AddRequiredArgument("directory_path", "Local directory path containing the module files").
		AddStep("get_terraform_module", "try to get the module using TRN format trn:terraform_module:{module_path}").
		AddStep("create_terraform_module", "if module doesn't exist, create it (extract name, system, and group_path from module_path)").
		AddStep("upload_module_version", "upload the module version using the module ID from get or create step").
		AddStep("", "confirm successful publication with module version details").
		Build()
}

// InspectWorkspacePrompt returns a prompt for inspecting a workspace.
func InspectWorkspacePrompt() (mcp.Prompt, mcp.PromptHandler) {
	return prompts.NewWorkflowPrompt(
		"inspect_workspace",
		"Get comprehensive status of workspace {workspace_path}",
	).
		AddRequiredArgument("workspace_path", "Full path to the workspace (e.g., group/subgroup/workspace-name)").
		AddStep("get_workspace", "retrieve workspace details and configuration").
		AddStep("get_workspace_outputs", "retrieve current Terraform outputs").
		AddStep("", "summarize workspace configuration and current state").
		Build()
}
