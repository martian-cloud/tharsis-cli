package command

import (
	"flag"
	"log/slog"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-hclog"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	env "github.com/qiangxue/go-env"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
)

type mcpConfig struct {
	Toolsets             string `env:"TOOLSETS"`
	Tools                string `env:"TOOLS"`
	ReadOnly             *bool  `env:"READ_ONLY"`
	NamespaceMutationACL string `env:"NAMESPACE_MUTATION_ACL"`
}

// mcpCommand is the top-level structure for the mcp command.
type mcpCommand struct {
	*BaseCommand

	toolsets             *string
	enabledTools         *string
	readOnly             *bool
	namespaceMutationACL *string
}

var _ Command = (*mcpCommand)(nil)

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
	); code != 0 {
		return code
	}

	// Load environment variables
	var cfg mcpConfig
	if err := env.New("THARSIS_MCP_", nil).Load(&cfg); err != nil {
		c.UI.ErrorWithSummary(err, "failed to load environment variables")
		return 1
	}

	// Command line args override environment variables
	if c.toolsets != nil {
		cfg.Toolsets = *c.toolsets
	}

	if c.enabledTools != nil {
		cfg.Tools = *c.enabledTools
	}

	if c.readOnly != nil {
		cfg.ReadOnly = c.readOnly
	}

	if c.namespaceMutationACL != nil {
		cfg.NamespaceMutationACL = *c.namespaceMutationACL
	}

	// Enable all toolsets by default if none specified
	if cfg.Toolsets == "" && cfg.Tools == "" {
		cfg.Toolsets = strings.Join(tools.AvailableToolsets(), ",")

		if cfg.ReadOnly == nil {
			// Default to read-only for safety
			cfg.ReadOnly = ptr.Bool(true)
		}
	}

	c.Logger.Debug("MCP server configuration",
		"toolsets", cfg.Toolsets,
		"tools", cfg.Tools,
		"read_only", cfg.ReadOnly,
		"namespace_mutation_acl", cfg.NamespaceMutationACL,
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
		tools.WithACLPatterns(cfg.NamespaceMutationACL),
	)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create tool context")
		return 1
	}

	toolsetGroup, err := tools.BuildToolsetGroup(ptr.ToBool(cfg.ReadOnly), toolContext)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to build toolset group")
		return 1
	}

	server, err := mcp.NewServer(&mcp.ServerConfig{
		Name:            "tharsis-cli",
		Title:           "Tharsis CLI MCP Server",
		Version:         c.Version,
		Logger:          slog.New(slog.NewTextHandler(c.Logger.StandardWriter(&hclog.StandardLoggerOptions{}), nil)),
		Instructions:    mcp.DefaultInstructions(),
		EnabledToolsets: cfg.Toolsets,
		EnabledTools:    cfg.Tools,
		ReadOnly:        ptr.ToBool(cfg.ReadOnly),
	}, toolsetGroup)
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
   The mcp command starts the Tharsis MCP server, enabling AI assistants
   to interact with Tharsis resources through the Model Context Protocol.
   By default, all toolsets are enabled in read-only mode for safety.

   Available toolsets: ` + strings.Join(tools.AvailableToolsets(), ", ") + `

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
	return `
# Start MCP server with production profile in read-only mode
tharsis -p production mcp

# Start with specific toolsets
tharsis mcp --toolsets auth,run

# Start with namespace ACL restrictions
tharsis mcp --namespace-mutation-acl "dev/*,staging/*"

# MCP Client Configuration (mcp.json):
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
`
}

func (c *mcpCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"toolsets",
		"Comma-separated list of toolsets to enable.",
		func(s string) error {
			c.toolsets = &s
			return nil
		},
	)
	f.Func(
		"tools",
		"Comma-separated list of individual tools to enable.",
		func(s string) error {
			c.enabledTools = &s
			return nil
		},
	)
	f.BoolFunc(
		"read-only",
		"Enable read-only mode (disables write tools).",
		func(s string) error {
			v, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			c.readOnly = &v
			return nil
		},
	)
	f.Func(
		"namespace-mutation-acl",
		"ACL patterns for namespace mutations (comma-separated).",
		func(s string) error {
			c.namespaceMutationACL = &s
			return nil
		},
	)

	return f
}
