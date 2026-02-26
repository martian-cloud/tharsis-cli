package command

import (
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestOutputNamespaceVariable_SensitiveValueMasking(t *testing.T) {
	tests := []struct {
		name           string
		variable       *sdktypes.NamespaceVariable
		toJSON         bool
		expectMasked   bool
		expectContains string
	}{
		{
			name: "sensitive variable with value should show value (retrieved with --show-sensitive)",
			variable: &sdktypes.NamespaceVariable{
				Key:           "db_password",
				Value:         stringPtr("secret123"),
				Category:      sdktypes.TerraformVariableCategory,
				NamespacePath: "test/group",
				Sensitive:     true,
			},
			toJSON:         false,
			expectMasked:   false,
			expectContains: "secret123",
		},
		{
			name: "non-sensitive variable should show value in human-readable output",
			variable: &sdktypes.NamespaceVariable{
				Key:           "region",
				Value:         stringPtr("us-east-1"),
				Category:      sdktypes.TerraformVariableCategory,
				NamespacePath: "test/group",
				Sensitive:     false,
			},
			toJSON:         false,
			expectMasked:   false,
			expectContains: "us-east-1",
		},
		{
			name: "sensitive variable with nil value should show masked (not retrieved with --show-sensitive)",
			variable: &sdktypes.NamespaceVariable{
				Key:           "empty_secret",
				Value:         nil,
				Category:      sdktypes.TerraformVariableCategory,
				NamespacePath: "test/group",
				Sensitive:     true,
			},
			toJSON:         false,
			expectMasked:   true,
			expectContains: "[SENSITIVE]",
		},
		{
			name: "non-sensitive variable with nil value should show masked",
			variable: &sdktypes.NamespaceVariable{
				Key:           "empty_var",
				Value:         nil,
				Category:      sdktypes.TerraformVariableCategory,
				NamespacePath: "test/group",
				Sensitive:     false,
			},
			toJSON:         false,
			expectMasked:   true,
			expectContains: "[SENSITIVE]",
		},
		{
			name: "sensitive variable with value should show value in JSON output",
			variable: &sdktypes.NamespaceVariable{
				Key:           "db_password",
				Value:         stringPtr("secret123"),
				Category:      sdktypes.TerraformVariableCategory,
				NamespacePath: "test/group",
				Sensitive:     true,
			},
			toJSON:         true,
			expectMasked:   false,
			expectContains: "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock UI to capture output
			ui := cli.NewMockUi()
			meta := &Metadata{
				UI: ui,
			}

			// Call the function
			exitCode := outputNamespaceVariable(meta, tt.toJSON, tt.variable)

			// Verify exit code
			assert.Equal(t, 0, exitCode, "Expected exit code 0")

			// Get the output
			output := ui.OutputWriter.String()

			// Verify the output contains expected string
			assert.Contains(t, output, tt.expectContains, "Output should contain expected string")

			// For human-readable output, verify masking behavior
			if !tt.toJSON {
				if tt.expectMasked {
					assert.Contains(t, output, "[SENSITIVE]", "Sensitive value should be masked")
					if tt.variable.Value != nil {
						assert.NotContains(t, output, *tt.variable.Value, "Actual value should not appear in output")
					}
				} else {
					assert.NotContains(t, output, "[SENSITIVE]", "Value should not be masked when explicitly retrieved")
				}
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
