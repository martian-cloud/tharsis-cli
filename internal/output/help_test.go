package output

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

func init() {
	// Disable color for consistent test output
	color.NoColor = true
}

func TestCommandHelp(t *testing.T) {
	t.Run("renders usage", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli [options] command",
		})
		assert.Contains(t, result, "mycli [options] command")
	})

	t.Run("renders description", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli run",
			Description: "Run the application with specified config.",
		})
		assert.Contains(t, result, "Run the application")
	})

	t.Run("renders flags with defaults", func(t *testing.T) {
		fs := flag.NewSet("Command options")
		var config *string
		var port *int
		var debug *bool
		fs.StringVar(&config, "config", "Path to config file", flag.Default("/etc/app.conf"))
		fs.IntVar(&port, "port", "Port to listen on", flag.Default(8080))
		fs.BoolVar(&debug, "debug", "Enable debug mode", flag.Default(false))

		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli serve",
			Flags: fs,
		})

		assert.Contains(t, result, "Command options:")
		assert.Contains(t, result, "-config")
		assert.Contains(t, result, "Default: /etc/app.conf")
		assert.Contains(t, result, "-port")
		assert.Contains(t, result, "Default: 8080")
		assert.Contains(t, result, "-debug")
		assert.Contains(t, result, "Default: false")
	})

	t.Run("renders flag aliases", func(t *testing.T) {
		fs := flag.NewSet("Options")
		var name *string
		fs.StringVar(&name, "name", "Resource name", flag.Aliases("n"))

		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli create",
			Flags: fs,
		})

		assert.Contains(t, result, "-name, -n")
	})

	t.Run("renders required flag with asterisk", func(t *testing.T) {
		fs := flag.NewSet("Options")
		var id *string
		fs.StringVar(&id, "id", "Resource ID", flag.Required())

		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli get",
			Flags: fs,
		})

		assert.Contains(t, result, "-id *")
	})

	t.Run("renders env var", func(t *testing.T) {
		fs := flag.NewSet("Options")
		var token *string
		fs.StringVar(&token, "token", "Auth token", flag.EnvVar("MY_TOKEN"))

		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli auth",
			Flags: fs,
		})

		assert.Contains(t, result, "Env: MY_TOKEN")
	})

	t.Run("renders combined metadata", func(t *testing.T) {
		fs := flag.NewSet("Options")
		var name *string
		fs.StringVar(&name, "name", "Resource name",
			flag.Default("default"), flag.EnvVar("NAME"), flag.Aliases("n"),
		)

		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli create",
			Flags: fs,
		})

		assert.Contains(t, result, "-name, -n")
		assert.Contains(t, result, "Default: default")
		assert.Contains(t, result, "Env: NAME")
	})

	t.Run("shows deprecated flags with message", func(t *testing.T) {
		fs := flag.NewSet("Options")
		var old *string
		var current *string
		fs.StringVar(&old, "old-flag", "Deprecated flag", flag.Deprecated("use --current-flag"))
		fs.StringVar(&current, "current-flag", "Current flag")

		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli run",
			Flags: fs,
		})

		assert.Contains(t, result, "-old-flag")
		assert.Contains(t, result, "use --current-flag")
		assert.Contains(t, result, "-current-flag")
	})

	t.Run("uses flag set name as title", func(t *testing.T) {
		fs := flag.NewSet("Global options")
		var verbose *bool
		fs.BoolVar(&verbose, "verbose", "Enable verbose output")

		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli",
			Flags: fs,
		})

		assert.Contains(t, result, "Global options:")
	})

	t.Run("renders example section", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli deploy",
			Example:     "mycli deploy --env production",
		})

		assert.Contains(t, result, "Example:")
		assert.Contains(t, result, "--env production")
	})

	t.Run("highlights JSON code blocks", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			Usage:       "mycli config",
			Description: "Config format:\n```json\n{\"key\": \"value\"}\n```",
		})

		assert.NotContains(t, result, "```")
		assert.Contains(t, result, "key")
	})

	t.Run("highlights HCL code blocks", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			Usage:       "mycli init",
			Description: "Example:\n```hcl\nresource \"test\" {\n  name = \"example\"\n}\n```",
		})

		assert.NotContains(t, result, "```")
		assert.Contains(t, result, "resource")
	})

	t.Run("handles nil flags", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli version",
		})

		assert.Contains(t, result, "mycli version")
		assert.NotContains(t, result, "options:")
	})

	t.Run("omits empty sections", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			Usage: "mycli version",
		})

		assert.NotContains(t, result, "Example:")
	})

	t.Run("highlights quoted command references in description", func(t *testing.T) {
		result := CommandHelp(CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli help",
			Description: `Use "mycli run" to execute.`,
		})

		assert.Contains(t, result, "mycli run")
	})
}
