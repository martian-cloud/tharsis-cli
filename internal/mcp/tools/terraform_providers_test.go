package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetTerraformProvider(t *testing.T) {
	type testCase struct {
		name        string
		input       *getTerraformProviderInput
		mockSetup   func(*mocks.TerraformProvidersClient)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name:  "get provider successfully",
			input: &getTerraformProviderInput{ID: "p1"},
			mockSetup: func(tpc *mocks.TerraformProvidersClient) {
				tpc.On("GetTerraformProviderByID", mock.Anything, &pb.GetTerraformProviderByIDRequest{Id: "p1"}).Return(&pb.TerraformProvider{
					Metadata:      &pb.ResourceMetadata{Id: "p1", Trn: "trn:terraform_provider:group/provider"},
					Name:          "provider",
					RepositoryUrl: "https://github.com/org/provider",
				}, nil)
			},
			expectID: "p1",
		},
		{
			name:  "provider not found",
			input: &getTerraformProviderInput{ID: "nonexistent"},
			mockSetup: func(tpc *mocks.TerraformProvidersClient) {
				tpc.On("GetTerraformProviderByID", mock.Anything, &pb.GetTerraformProviderByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockProviders := mocks.NewTerraformProvidersClient(t)

			if tc.mockSetup != nil {
				tc.mockSetup(mockProviders)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{TerraformProvidersClient: mockProviders},
			}

			_, handler := getTerraformProvider(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.Provider.ID)
		})
	}
}
