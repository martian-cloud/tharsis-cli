package tools

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetApply(t *testing.T) {
	type testCase struct {
		name        string
		input       *getApplyInput
		mockSetup   func(*mocks.RunsClient)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "get apply successfully",
			input: &getApplyInput{ID: "trn:apply:a1"},
			mockSetup: func(rc *mocks.RunsClient) {
				rc.On("GetApplyByID", mock.Anything, &pb.GetApplyByIDRequest{Id: "trn:apply:a1"}).Return(&pb.Apply{
					Metadata:     &pb.ResourceMetadata{Id: "a1", Trn: "trn:apply:a1"},
					Status:       "finished",
					TriggeredBy:  "user1",
					ErrorMessage: nil,
				}, nil)
			},
		},
		{
			name:  "apply not found",
			input: &getApplyInput{ID: "nonexistent"},
			mockSetup: func(rc *mocks.RunsClient) {
				rc.On("GetApplyByID", mock.Anything, &pb.GetApplyByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRuns := mocks.NewRunsClient(t)

			if tc.mockSetup != nil {
				tc.mockSetup(mockRuns)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{RunsClient: mockRuns},
			}

			_, handler := getApply(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output.Apply)
		})
	}
}
