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

func TestGetTerraformProviderPlatform(t *testing.T) {
	type testCase struct {
		name        string
		input       *getTerraformProviderPlatformInput
		mockSetup   func(*mocks.TerraformProvidersClient)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name:  "get provider platform successfully",
			input: &getTerraformProviderPlatformInput{ID: "pp1"},
			mockSetup: func(tpc *mocks.TerraformProvidersClient) {
				tpc.On("GetTerraformProviderPlatformByID", mock.Anything, &pb.GetTerraformProviderPlatformByIDRequest{Id: "pp1"}).Return(&pb.TerraformProviderPlatform{
					Metadata:          &pb.ResourceMetadata{Id: "pp1", Trn: "trn:terraform_provider_platform:group/provider/version/platform"},
					ProviderVersionId: "pv1",
					OperatingSystem:   "linux",
					Architecture:      "amd64",
					ShaSum:            "abc123",
					Filename:          "terraform-provider-aws_v1.0.0_linux_amd64.zip",
					BinaryUploaded:    true,
				}, nil)
			},
			expectID: "pp1",
		},
		{
			name:  "provider platform not found",
			input: &getTerraformProviderPlatformInput{ID: "nonexistent"},
			mockSetup: func(tpc *mocks.TerraformProvidersClient) {
				tpc.On("GetTerraformProviderPlatformByID", mock.Anything, &pb.GetTerraformProviderPlatformByIDRequest{Id: "nonexistent"}).
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

			_, handler := getTerraformProviderPlatform(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.Platform.ID)
		})
	}
}
