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

func TestGetPlan(t *testing.T) {
	type testCase struct {
		name        string
		input       *getPlanInput
		mockSetup   func(*mocks.RunsClient)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name:  "get plan successfully",
			input: &getPlanInput{ID: "p1"},
			mockSetup: func(rc *mocks.RunsClient) {
				rc.On("GetPlanByID", mock.Anything, &pb.GetPlanByIDRequest{Id: "p1"}).Return(&pb.Plan{
					Metadata:   &pb.ResourceMetadata{Id: "p1", Trn: "trn:plan:p1"},
					Status:     "finished",
					HasChanges: true,
				}, nil)
			},
			expectID: "p1",
		},
		{
			name:  "plan not found",
			input: &getPlanInput{ID: "nonexistent"},
			mockSetup: func(rc *mocks.RunsClient) {
				rc.On("GetPlanByID", mock.Anything, &pb.GetPlanByIDRequest{Id: "nonexistent"}).
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

			_, handler := getPlan(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.Plan.ID)
		})
	}
}
