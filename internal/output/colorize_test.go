package output

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestColorize(t *testing.T) {
	// Force color on for tests since CI has no TTY.
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = true })

	bold := color.New(color.Bold)
	green := PrimaryColor()

	t.Run("highlights product name", func(t *testing.T) {
		result := Colorize("tharsis workspace create", "tharsis")

		assert.Contains(t, result, green.Sprint("tharsis"))
	})

	t.Run("bolds section headers", func(t *testing.T) {
		result := Colorize("Command options:", "tharsis")

		assert.Contains(t, result, bold.Sprint("Command options:"))
	})

	t.Run("highlights flag tokens", func(t *testing.T) {
		result := Colorize("  -name\n      Resource name", "")

		assert.Contains(t, result, green.Sprint("-name"))
	})

	t.Run("highlights flags in examples", func(t *testing.T) {
		result := Colorize("tharsis mcp -toolsets auth", "tharsis")

		assert.Contains(t, result, green.Sprint("-toolsets"))
	})

	t.Run("bolds metadata labels", func(t *testing.T) {
		result := Colorize("      Values: a, b, c", "")

		assert.Contains(t, result, bold.Sprint("Values:"))
	})

	t.Run("highlights quoted command references", func(t *testing.T) {
		result := Colorize(`See "tharsis run" for details.`, "tharsis")

		assert.Contains(t, result, green.Sprint("tharsis run"))
	})

	t.Run("strips code block markers and preserves code", func(t *testing.T) {
		input := "Config:\n```json\n{\"key\": \"value\"}\n```"
		result := Colorize(input, "")

		assert.NotContains(t, result, "```")
		assert.Contains(t, result, "key")
	})

	t.Run("returns input unchanged when NoColor", func(t *testing.T) {
		color.NoColor = true
		defer func() { color.NoColor = false }()

		input := "tharsis workspace create\nCommand options:\n  -name"
		result := Colorize(input, "tharsis")

		// Code blocks should still be stripped.
		assert.Equal(t, input, result)
	})

	t.Run("strips code blocks even when NoColor", func(t *testing.T) {
		color.NoColor = true
		defer func() { color.NoColor = false }()

		result := Colorize("```json\n{\"key\": \"value\"}\n```", "")

		assert.NotContains(t, result, "```")
		assert.Contains(t, result, "key")
	})

	t.Run("preserves newline structure", func(t *testing.T) {
		input := "line1\n\nline3"
		result := Colorize(input, "")

		assert.Equal(t, input, result)
	})

	t.Run("handles empty input", func(t *testing.T) {
		assert.Equal(t, "", Colorize("", ""))
	})

	t.Run("handles empty product name", func(t *testing.T) {
		result := Colorize("tharsis workspace", "")

		assert.NotContains(t, result, "\x1b[")
	})
}
