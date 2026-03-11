package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetJobLogs(t *testing.T) {
	type testCase struct {
		name        string
		input       *getJobLogsInput
		mockSetup   func(*mocks.JobsClient)
		expectError bool
		expectSize  int
		expectMore  bool
	}

	testCases := []testCase{
		{
			name:  "get job logs successfully",
			input: &getJobLogsInput{JobID: "j1"},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetJobLogs", mock.Anything, &pb.GetJobLogsRequest{
					JobId:       "j1",
					StartOffset: 0,
					Limit:       defaultLogLimit + 1,
				}).Return(&pb.GetJobLogsResponse{
					Logs: "log line 1\nlog line 2\n",
				}, nil)
			},
			expectSize: 22,
			expectMore: false,
		},
		{
			name:  "get job logs with custom start and limit",
			input: &getJobLogsInput{JobID: "j1", Start: ptr.Int(100), Limit: ptr.Int(500)},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetJobLogs", mock.Anything, &pb.GetJobLogsRequest{
					JobId:       "j1",
					StartOffset: 100,
					Limit:       501,
				}).Return(&pb.GetJobLogsResponse{
					Logs: "partial logs",
				}, nil)
			},
			expectSize: 12,
			expectMore: false,
		},
		{
			name:  "has more logs",
			input: &getJobLogsInput{JobID: "j1", Limit: ptr.Int(10)},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetJobLogs", mock.Anything, &pb.GetJobLogsRequest{
					JobId:       "j1",
					StartOffset: 0,
					Limit:       11,
				}).Return(&pb.GetJobLogsResponse{
					Logs: "12345678901",
				}, nil)
			},
			expectSize: 10,
			expectMore: true,
		},
		{
			name:        "limit exceeds maximum",
			input:       &getJobLogsInput{JobID: "j1", Limit: ptr.Int(maxLogLimit + 1)},
			expectError: true,
		},
		{
			name:  "job not found",
			input: &getJobLogsInput{JobID: "nonexistent"},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetJobLogs", mock.Anything, &pb.GetJobLogsRequest{
					JobId:       "nonexistent",
					StartOffset: 0,
					Limit:       defaultLogLimit + 1,
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockJobs := mocks.NewJobsClient(t)

			if tc.mockSetup != nil {
				tc.mockSetup(mockJobs)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{JobsClient: mockJobs},
			}

			_, handler := getJobLogs(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectSize, output.Size)
			assert.Equal(t, tc.expectMore, output.HasMore)
		})
	}
}

func TestGetLatestJob(t *testing.T) {
	type testCase struct {
		name        string
		input       *getLatestJobInput
		mockSetup   func(*mocks.JobsClient)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "get latest job for plan",
			input: &getLatestJobInput{PlanID: ptr.String("plan1")},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetLatestJobForPlan", mock.Anything, &pb.GetLatestJobForPlanRequest{
					PlanId: "plan1",
				}).Return(&pb.Job{
					Metadata:       &pb.ResourceMetadata{Id: "job1", Trn: "trn:job:job1"},
					WorkspaceId:    "ws1",
					RunId:          "run1",
					Type:           "plan",
					Status:         "finished",
					MaxJobDuration: 60,
				}, nil)
			},
		},
		{
			name:  "get latest job for apply",
			input: &getLatestJobInput{ApplyID: ptr.String("apply1")},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetLatestJobForApply", mock.Anything, &pb.GetLatestJobForApplyRequest{
					ApplyId: "apply1",
				}).Return(&pb.Job{
					Metadata:       &pb.ResourceMetadata{Id: "job2", Trn: "trn:job:job2"},
					WorkspaceId:    "ws1",
					RunId:          "run1",
					Type:           "apply",
					Status:         "running",
					MaxJobDuration: 120,
				}, nil)
			},
		},
		{
			name:        "neither plan nor apply provided",
			input:       &getLatestJobInput{},
			expectError: true,
		},
		{
			name:        "both plan and apply provided",
			input:       &getLatestJobInput{PlanID: ptr.String("plan1"), ApplyID: ptr.String("apply1")},
			expectError: true,
		},
		{
			name:  "plan not found",
			input: &getLatestJobInput{PlanID: ptr.String("nonexistent")},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetLatestJobForPlan", mock.Anything, &pb.GetLatestJobForPlanRequest{
					PlanId: "nonexistent",
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
		{
			name:  "apply not found",
			input: &getLatestJobInput{ApplyID: ptr.String("nonexistent")},
			mockSetup: func(jc *mocks.JobsClient) {
				jc.On("GetLatestJobForApply", mock.Anything, &pb.GetLatestJobForApplyRequest{
					ApplyId: "nonexistent",
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockJobs := mocks.NewJobsClient(t)

			if tc.mockSetup != nil {
				tc.mockSetup(mockJobs)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{JobsClient: mockJobs},
			}

			_, handler := getLatestJob(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output.Job)
		})
	}
}
