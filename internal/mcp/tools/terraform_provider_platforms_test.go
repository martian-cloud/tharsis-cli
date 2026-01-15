package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGetTerraformProviderPlatform(t *testing.T) {
	type testCase struct {
		name           string
		input          getTerraformProviderPlatformInput
		setupMocks     func(*tharsis.MockClient, *tharsis.TerraformProviderPlatform)
		expectedOutput getTerraformProviderPlatformOutput
		expectError    bool
	}

	testCases := []testCase{
		{
			name: "successful get",
			input: getTerraformProviderPlatformInput{
				ID: "platform-123",
			},
			setupMocks: func(client *tharsis.MockClient, platforms *tharsis.TerraformProviderPlatform) {
				client.On("TerraformProviderPlatforms").Return(platforms)
				platforms.On("GetProviderPlatform", mock.Anything, &sdktypes.GetTerraformProviderPlatformInput{
					ID: "platform-123",
				}).Return(&sdktypes.TerraformProviderPlatform{
					Metadata: sdktypes.ResourceMetadata{
						ID:  "platform-123",
						TRN: "trn:terraform_provider_platform:group/provider/1.0.0/platform-123",
					},
					ProviderVersionID: "version-456",
					OperatingSystem:   "linux",
					Architecture:      "amd64",
					SHASum:            "abc123",
					Filename:          "terraform-provider-test_1.0.0_linux_amd64.zip",
					BinaryUploaded:    true,
				}, nil)
			},
			expectedOutput: getTerraformProviderPlatformOutput{
				Platform: terraformProviderPlatform{
					ID:                "platform-123",
					TRN:               "trn:terraform_provider_platform:group/provider/1.0.0/platform-123",
					ProviderVersionID: "version-456",
					OperatingSystem:   "linux",
					Architecture:      "amd64",
					SHASum:            "abc123",
					Filename:          "terraform-provider-test_1.0.0_linux_amd64.zip",
					BinaryUploaded:    true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockPlatforms := tharsis.NewTerraformProviderPlatform(t)
			tc.setupMocks(mockClient, mockPlatforms)

			toolContext := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
			}

			_, handler := getTerraformProviderPlatform(toolContext)
			_, output, err := handler(context.Background(), nil, tc.input)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}
