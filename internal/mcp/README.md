# MCP Package

This package provides Model Context Protocol (MCP) server functionality for the Tharsis CLI, enabling AI assistants to interact with Tharsis resources.

## Structure

- **acl/** - Access control for restricting tool operations to specific groups/workspaces via pattern matching
- **prompts/** - MCP workflow prompts for guided multi-step operations (deployments, diagnostics, etc.)
- **tharsis/** - Tharsis SDK client wrapper and mocks for testing
- **tools/** - MCP tool implementations, toolset configuration, and the main `ToolContext`

## Usage

The MCP server is started via `tharsis mcp` command. Configuration is done through environment variables:

- `THARSIS_MCP_TOOLSETS` - Comma-separated list of toolsets to enable
- `THARSIS_MCP_TOOLS` - Comma-separated list of individual tools to enable
- `THARSIS_MCP_READ_ONLY` - Enable read-only mode (disables write tools)
- `THARSIS_MCP_NAMESPACE_MUTATION_ACL` - ACL patterns for namespace mutations (groups and workspaces)

## Available Toolsets

- `auth` - Authentication (SSO login, connection info)
- `run` - Run management (create, apply, cancel)
- `job` - Job logs retrieval
- `configuration_version` - Configuration version management
- `workspace` - Workspace CRUD operations
- `group` - Group CRUD operations
- `variable` - Terraform and environment variable management
- `managed_identity` - Managed identity assignment
- `documentation` - Tharsis documentation search
- `terraform_module` - Module registry management
- `terraform_module_version` - Module version management
- `terraform_provider` - Provider registry
- `terraform_provider_platform` - Provider platform details
