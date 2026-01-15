package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGetTerraformProvider(t *testing.T) {
	providerID := "provider-id"

	type testCase struct {
		name     string
		input    getTerraformProviderInput
		provider *sdktypes.TerraformProvider
		validate func(*testing.T, getTerraformProviderOutput)
	}

	tests := []testCase{
		{
			name: "get provider by ID",
			input: getTerraformProviderInput{
				ID: providerID,
			},
			provider: &sdktypes.TerraformProvider{
				Metadata:          sdktypes.ResourceMetadata{ID: providerID, TRN: "trn:terraform_provider:group/aws"},
				Name:              "aws",
				GroupPath:         "group",
				RegistryNamespace: "hashicorp",
				RepositoryURL:     "https://github.com/hashicorp/terraform-provider-aws",
				Private:           false,
			},
			validate: func(t *testing.T, output getTerraformProviderOutput) {
				assert.Equal(t, providerID, output.Provider.ID)
				assert.Equal(t, "aws", output.Provider.Name)
				assert.Equal(t, "https://github.com/hashicorp/terraform-provider-aws", output.Provider.RepositoryURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockProvider := tharsis.NewTerraformProvider(t)

			mockClient.On("TerraformProviders").Return(mockProvider)
			mockProvider.On("GetProvider", mock.Anything, mock.Anything).Return(tt.provider, nil)

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
			}

			_, handler := getTerraformProvider(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}
