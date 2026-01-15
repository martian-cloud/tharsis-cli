package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	env "github.com/qiangxue/go-env"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
)

type mcpEnvConfig struct {
	Toolsets             string `env:"TOOLSETS"`
	Tools                string `env:"TOOLS"`
	ReadOnly             *bool  `env:"READ_ONLY"`
	NamespaceMutationACL string `env:"NAMESPACE_MUTATION_ACL"`
}

type mcpCommand struct {
	meta *Metadata
}

// NewMCPCommandFactory returns a factory function for creating MCP commands.
func NewMCPCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return &mcpCommand{meta: meta}, nil
	}
}

func (mc *mcpCommand) Run(args []string) int {
	mc.meta.Logger.Debugf("Starting the 'mcp' command with %d arguments:", len(args))
	for ix, arg := range args {
		mc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	defs := mc.buildMCPDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mc.meta.BinaryName+" mcp", defs, args)
	if err != nil {
		mc.meta.Logger.Error(output.FormatError("failed to parse mcp options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive mcp arguments: %s", cmdArgs)
		mc.meta.Logger.Error(output.FormatError(msg, nil), mc.HelpMCP())
		return 1
	}

	// Load environment variables first
	var envCfg mcpEnvConfig
	if err = env.New("THARSIS_MCP_", nil).Load(&envCfg); err != nil {
		mc.meta.UI.Error(output.FormatError("failed to load environment variables", err))
		return 1
	}

	toolsets := envCfg.Toolsets
	enabledTools := envCfg.Tools
	readOnly := envCfg.ReadOnly != nil && *envCfg.ReadOnly
	namespaceMutationACL := envCfg.NamespaceMutationACL

	// Command line args take precedence
	if opts, ok := cmdOpts["toolsets"]; ok {
		toolsets = strings.Join(opts, ",")
	}
	if opts, ok := cmdOpts["tools"]; ok {
		enabledTools = strings.Join(opts, ",")
	}
	if _, ok := cmdOpts["read-only"]; ok {
		readOnly = true
	}
	if opts, ok := cmdOpts["namespace-mutation-acl"]; ok {
		namespaceMutationACL = strings.Join(opts, ",")
	}

	// Enable all toolsets by default if none specified
	if toolsets == "" && enabledTools == "" {
		toolsets = strings.Join(tools.AvailableToolsets(), ",")
		// Only default to read-only if not explicitly set
		if envCfg.ReadOnly == nil && cmdOpts["read-only"] == nil {
			readOnly = true
		}
	}

	mc.meta.Logger.Debugw("MCP server configuration",
		"toolsets", toolsets,
		"tools", enabledTools,
		"read_only", readOnly,
		"namespace_mutation_acl", namespaceMutationACL,
	)

	currentSettings, err := mc.meta.ReadSettings()
	if err != nil {
		mc.meta.UI.Error(output.FormatError("failed to read settings", err))
		return 1
	}

	toolContext, err := tools.NewToolContext(
		currentSettings.CurrentProfile.TharsisURL,
		mc.meta.CurrentProfileName,
		mc.meta.HTTPClient,
		mc.meta.GetSDKClient,
		tools.WithACLPatterns(namespaceMutationACL),
	)
	if err != nil {
		mc.meta.UI.Error(output.FormatError("failed to create tool context", err))
		return 1
	}

	toolsetGroup, err := tools.BuildToolsetGroup(readOnly, toolContext)
	if err != nil {
		mc.meta.UI.Error(output.FormatError("failed to build toolset group", err))
		return 1
	}

	server, err := mcp.NewServer(&mcp.ServerConfig{
		Name:            "tharsis-cli",
		Title:           "Tharsis CLI MCP Server",
		Version:         mc.meta.Version,
		Logger:          mc.meta.Logger.Slog(),
		Instructions:    mcp.DefaultInstructions(),
		EnabledToolsets: toolsets,
		EnabledTools:    enabledTools,
		ReadOnly:        readOnly,
	}, toolsetGroup)
	if err != nil {
		mc.meta.UI.Error(output.FormatError("failed to create server", err))
		return 1
	}

	if err := server.Run(context.Background(), &sdkmcp.StdioTransport{}); err != nil {
		mc.meta.UI.Error(output.FormatError("server error", err))
		return 1
	}

	return 0
}

func (mcpCommand) buildMCPDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"toolsets": {
			Arguments: []string{"Toolsets"},
			Synopsis:  "Comma-separated list of toolsets to enable.",
		},
		"tools": {
			Arguments: []string{"Tools"},
			Synopsis:  "Comma-separated list of individual tools to enable.",
		},
		"read-only": {
			Arguments: []string{},
			Synopsis:  "Enable read-only mode (disables write tools).",
		},
		"namespace-mutation-acl": {
			Arguments: []string{"Patterns"},
			Synopsis:  "ACL patterns for namespace mutations (comma-separated).",
		},
	}
}

func (mc *mcpCommand) Synopsis() string {
	return "Starts the Tharsis MCP server."
}

func (mc *mcpCommand) Help() string {
	return mc.HelpMCP()
}

func (mc *mcpCommand) HelpMCP() string {
	return fmt.Sprintf(`
Usage: tharsis [global options] mcp [options]

   Starts the Tharsis MCP server, enabling AI assistants to interact with Tharsis
   resources through the Model Context Protocol. By default, all toolsets are
   enabled in read-only mode for safety.

%s

   Available toolsets: %s

   Environment variables (command-line options take precedence):
   - THARSIS_MCP_TOOLSETS
   - THARSIS_MCP_TOOLS
   - THARSIS_MCP_READ_ONLY
   - THARSIS_MCP_NAMESPACE_MUTATION_ACL

Access Control (ACL) Patterns:

   Control which namespaces (groups and workspaces) can be modified using simple
   wildcard patterns. ACL patterns apply to write operations (create, update,
   delete, apply) to prevent accidental changes to production resources. Read
   operations (get, list) are only restricted by user permissions.

   Patterns are case-insensitive and support:

   - Exact match: "prod" matches only "prod"
   - Wildcard: "prod/*" matches any path starting with "prod/" (all levels)
   - Prefix/suffix: "prod/team-*" matches "prod/team-alpha", "prod/team-beta"

   Tip: Wildcards match across all path levels. To match a specific resource,
   use exact paths like "prod/workspace" instead of "prod/*".

   Examples:
   - "prod" - Allow access to the "prod" group only
   - "prod/workspace" - Allow access to specific workspace
   - "prod/*" - Allow access to all resources under "prod" at any depth
   - "prod/team-*" - Allow access to resources matching "prod/team-*"
   - "dev,staging" - Allow access to "dev" and "staging" (comma-separated)

   Restrictions:
   - Wildcard-only patterns ("*") are not allowed
   - Patterns cannot start with a wildcard ("*/workspace")

MCP Client Configuration (mcp.json):

   {
     "mcpServers": {
       "tharsis-prod": {
         "command": "tharsis",
         "args": ["-p", "production", "mcp"],
         "env": {"THARSIS_MCP_READ_ONLY": "true"},
         "disabled": false,
         "autoApprove": []
       },
       "tharsis-dev": {
         "command": "tharsis",
         "args": ["-p", "development", "mcp"],
         "env": {"THARSIS_MCP_TOOLSETS": "auth,run"},
         "disabled": false,
         "autoApprove": []
       }
     }
   }

`, buildHelpText(mc.buildMCPDefs()), strings.Join(tools.AvailableToolsets(), ", "))
}
