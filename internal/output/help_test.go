package output

import (
	"flag"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func init() {
	// Disable color for consistent test output
	color.NoColor = true
}

func TestCommandHelp(t *testing.T) {
	t.Run("highlights product name in usage", func(t *testing.T) {
		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli [options] command",
		}
		result := CommandHelp(info)
		// With color disabled, just check product name appears
		if !strings.Contains(result, "mycli") {
			t.Error("expected product name in usage line")
		}
	})

	t.Run("renders description", func(t *testing.T) {
		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli run",
			Description: "Run the application with specified config.",
		}
		result := CommandHelp(info)
		if !strings.Contains(result, "Run the application") {
			t.Error("expected description in output")
		}
	})

	t.Run("renders flags with defaults", func(t *testing.T) {
		flags := flag.NewFlagSet("test", flag.ContinueOnError)
		flags.String("config", "/etc/app.conf", "Path to config file")
		flags.Int("port", 8080, "Port to listen on")
		flags.Bool("debug", false, "Enable debug mode")

		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli serve",
			Flags:       flags,
		}
		result := CommandHelp(info)

		if !strings.Contains(result, "-config") {
			t.Error("expected -config flag")
		}
		if !strings.Contains(result, "/etc/app.conf") {
			t.Error("expected default value for config")
		}
		if !strings.Contains(result, "-port") {
			t.Error("expected -port flag")
		}
		if !strings.Contains(result, "Command options:") {
			t.Error("expected Command options header")
		}
	})

	t.Run("renders example section", func(t *testing.T) {
		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli deploy",
			Example:     "mycli deploy --env production",
		}
		result := CommandHelp(info)

		if !strings.Contains(result, "Example:") {
			t.Error("expected Example header")
		}
		if !strings.Contains(result, "--env production") {
			t.Error("expected example content")
		}
	})

	t.Run("highlights JSON code blocks", func(t *testing.T) {
		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli config",
			Description: "Config format:\n```json\n{\"key\": \"value\", \"count\": 42}\n```",
		}
		result := CommandHelp(info)

		// Code block markers should be stripped
		if strings.Contains(result, "```") {
			t.Error("code block markers should be removed")
		}
		// Content should remain
		if !strings.Contains(result, "key") {
			t.Error("expected JSON content to remain")
		}
	})

	t.Run("highlights HCL code blocks", func(t *testing.T) {
		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli init",
			Description: "Example config:\n```hcl\nresource \"test\" {\n  name = \"example\"\n}\n```",
		}
		result := CommandHelp(info)

		if strings.Contains(result, "```") {
			t.Error("code block markers should be removed")
		}
		if !strings.Contains(result, "resource") {
			t.Error("expected HCL content to remain")
		}
	})

	t.Run("auto-highlights shell examples", func(t *testing.T) {
		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli run",
			Example:     "mycli run --verbose",
		}
		result := CommandHelp(info)

		// Example should be present (highlighting tested via color codes when enabled)
		if !strings.Contains(result, "mycli run --verbose") {
			t.Error("expected shell example in output")
		}
	})

	t.Run("handles empty optional fields", func(t *testing.T) {
		info := CommandHelpInfo{
			ProductName: "mycli",
			Usage:       "mycli version",
		}
		result := CommandHelp(info)

		if !strings.Contains(result, "mycli version") {
			t.Error("expected usage in output")
		}
		// Should not have Example or Command options sections
		if strings.Contains(result, "Example:") {
			t.Error("should not have Example section when empty")
		}
	})
}
