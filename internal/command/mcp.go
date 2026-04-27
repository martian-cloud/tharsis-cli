package command

import (
	"errors"
	"log/slog"
	"regexp"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-hclog"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
)

// mcpCommand is the top-level structure for the mcp command.
type mcpCommand struct {
	*BaseCommand

	toolsets             *string
	enabledTools         *string
	readOnly             *bool
	namespaceMutationACL *string
}

var _ Command = (*mcpCommand)(nil)

func (c *mcpCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewMCPCommandFactory returns a mcpCommand struct.
func NewMCPCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &mcpCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *mcpCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("mcp"),
		WithClient(true),
		WithInputValidator(c.validate),
	); code != 0 {
		return code
	}

	// Enable all toolsets by default if none specified
	toolsets := ptr.ToString(c.toolsets)
	enabledTools := ptr.ToString(c.enabledTools)
	if toolsets == "" && enabledTools == "" {
		toolsets = strings.Join(tools.AvailableToolsets(), ",")

		if c.readOnly == nil {
			// Default to read-only for safety
			c.readOnly = ptr.Bool(true)
		}
	}

	c.Logger.Debug("MCP server configuration",
		"toolsets", toolsets,
		"tools", enabledTools,
		"read_only", c.readOnly,
		"namespace_mutation_acl", c.namespaceMutationACL,
	)

	currentSettings, err := c.getCurrentSettings()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get current settings")
		return 1
	}

	tokenGetter, err := currentSettings.CurrentProfile.NewTokenGetter(c.Context)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create token getter")
		return 1
	}

	tfeClient, err := tfe.NewRESTClient(currentSettings.CurrentProfile.Endpoint, tokenGetter, c.HTTPClient)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create tfe rest client")
		return 1
	}

	toolContext, err := tools.NewToolContext(
		currentSettings.CurrentProfile.Endpoint,
		c.CurrentProfileName,
		c.HTTPClient,
		c.grpcClient,
		tfeClient,
		tools.WithACLPatterns(ptr.ToString(c.namespaceMutationACL)),
	)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create tool context")
		return 1
	}

	var normalizedProfileName string
	if c.CurrentProfileName != "default" {
		// Normalize the profile name so it conforms to tool naming.
		normalizedProfileName = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(c.CurrentProfileName), "_") + "_"
	}

	server, err := mcp.NewServer(&mcp.ServerConfig{
		Name:            "tharsis-cli",
		Title:           "Tharsis CLI MCP Server",
		Version:         c.Version,
		Logger:          slog.New(slog.NewTextHandler(c.Logger.StandardWriter(&hclog.StandardLoggerOptions{}), nil)),
		Instructions:    mcp.DefaultInstructions(),
		EnabledToolsets: toolsets,
		EnabledTools:    enabledTools,
		Prefix:          normalizedProfileName,
		ReadOnly:        ptr.ToBool(c.readOnly),
	}, tools.BuildToolsetGroup(ptr.ToBool(c.readOnly), toolContext))
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create server")
		return 1
	}

	if err := server.Run(c.Context, &sdkmcp.StdioTransport{}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to start mcp server on stdio transport")
		return 1
	}

	return 0
}

func (*mcpCommand) Synopsis() string {
	return "Start the Tharsis MCP server."
}

func (*mcpCommand) Usage() string {
	return "tharsis [global options] mcp [options]"
}

func (*mcpCommand) Description() string {
	return `
   Starts the Tharsis MCP server, enabling AI assistants to interact
   with Tharsis resources through the Model Context Protocol.
   By default, all toolsets are enabled in read-only mode for safety.

   Available toolsets:
    ` + output.Wrap(strings.Join(tools.AvailableToolsets(), ", ")) + `

   Environment variables (command-line options take precedence):
     THARSIS_MCP_TOOLSETS               Comma-separated list of toolsets to enable
     THARSIS_MCP_TOOLS                  Comma-separated list of individual tools to enable
     THARSIS_MCP_READ_ONLY              Enable read-only mode (true/false)
     THARSIS_MCP_NAMESPACE_MUTATION_ACL ACL patterns for namespace mutations

   Access Control (ACL) Patterns:

   Control which namespaces (groups and workspaces) can be modified using
   simple wildcard patterns. ACL patterns apply to write operations (create,
   update, delete, apply) to prevent accidental changes to production resources.
   Read operations (get, list) are only restricted by user permissions.

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
`
}

func (*mcpCommand) Example() string {
	return "```bash" + `
# Start MCP server with production profile in read-only mode
tharsis -p production mcp

# Start with specific toolsets
tharsis mcp -toolsets auth,run

# Start with namespace ACL restrictions
tharsis mcp -namespace-mutation-acl "dev/*,staging/*"
` + "```" + `

MCP Client Configuration (mcp.json):
` + "```json" + `
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
` + "```" + `
`
}

func (c *mcpCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.toolsets,
		"toolsets",
		"Comma-separated list of toolsets to enable.",
		flag.EnvVar("THARSIS_MCP_TOOLSETS"),
	)
	f.StringVar(
		&c.enabledTools,
		"tools",
		"Comma-separated list of individual tools to enable.",
		flag.EnvVar("THARSIS_MCP_TOOLS"),
	)
	f.BoolVar(
		&c.readOnly,
		"read-only",
		"Enable read-only mode (disables write tools).",
		flag.EnvVar("THARSIS_MCP_READ_ONLY"),
	)
	f.StringVar(
		&c.namespaceMutationACL,
		"namespace-mutation-acl",
		"ACL patterns for namespace mutations (comma-separated).",
		flag.EnvVar("THARSIS_MCP_NAMESPACE_MUTATION_ACL"),
	)

	return f
}
