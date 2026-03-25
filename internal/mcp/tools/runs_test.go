package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type runMocks struct {
	runs *mocks.RunsClient
	acl  *acl.MockChecker
}

func TestListRuns(t *testing.T) {
	type testCase struct {
		name          string
		input         *listRunsInput
		mockSetup     func(*runMocks)
		expectError   bool
		expectResults int
	}

	testCases := []testCase{
		{
			name:  "list runs successfully",
			input: &listRunsInput{WorkspaceID: ptr.String("trn:workspace:group/ws1"), Limit: ptr.Int32(10)},
			mockSetup: func(m *runMocks) {
				m.runs.On("GetRuns", mock.Anything, &pb.GetRunsRequest{
					WorkspaceId:       ptr.String("trn:workspace:group/ws1"),
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(&pb.GetRunsResponse{
					Runs: []*pb.Run{
						{Metadata: &pb.ResourceMetadata{Id: "r1", Trn: "trn:run:r1"}, Status: "applied", WorkspaceId: "ws1"},
					},
					PageInfo: &pb.PageInfo{HasNextPage: false},
				}, nil)
			},
			expectResults: 1,
		},
		{
			name:  "grpc error",
			input: &listRunsInput{WorkspaceID: ptr.String("trn:workspace:group/ws1"), Limit: ptr.Int32(10)},
			mockSetup: func(m *runMocks) {
				m.runs.On("GetRuns", mock.Anything, &pb.GetRunsRequest{
					WorkspaceId:       ptr.String("trn:workspace:group/ws1"),
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &runMocks{
				runs: mocks.NewRunsClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					RunsClient: testMocks.runs,
				},
			}

			_, handler := listRuns(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, output.Runs, tc.expectResults)
		})
	}
}

func TestGetRun(t *testing.T) {
	type testCase struct {
		name        string
		input       *getRunInput
		mockSetup   func(*runMocks)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name:  "get run successfully",
			input: &getRunInput{ID: "r1"},
			mockSetup: func(m *runMocks) {
				m.runs.On("GetRunByID", mock.Anything, &pb.GetRunByIDRequest{Id: "r1"}).Return(&pb.Run{
					Metadata:    &pb.ResourceMetadata{Id: "r1", Trn: "trn:run:r1"},
					Status:      "applied",
					WorkspaceId: "ws1",
				}, nil)
			},
			expectID: "r1",
		},
		{
			name:  "run not found",
			input: &getRunInput{ID: "nonexistent"},
			mockSetup: func(m *runMocks) {
				m.runs.On("GetRunByID", mock.Anything, &pb.GetRunByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &runMocks{
				runs: mocks.NewRunsClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					RunsClient: testMocks.runs,
				},
			}

			_, handler := getRun(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.Run.ID)
		})
	}
}

func TestCreateRun(t *testing.T) {
	type testCase struct {
		name        string
		input       *createRunInput
		mockSetup   func(*runMocks)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name: "create run successfully",
			input: &createRunInput{
				WorkspaceID: "ws1",
			},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.runs.On("CreateRun", mock.Anything, &pb.CreateRunRequest{
					WorkspaceId: "ws1",
				}).Return(&pb.Run{
					Metadata:    &pb.ResourceMetadata{Id: "r1", Trn: "trn:run:r1"},
					Status:      "pending",
					WorkspaceId: "ws1",
				}, nil)
			},
			expectID: "r1",
		},
		{
			name: "acl denial",
			input: &createRunInput{
				WorkspaceID: "ws1",
			},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "create run with module source",
			input: &createRunInput{
				WorkspaceID:   "ws1",
				ModuleSource:  ptr.String("registry.terraform.io/hashicorp/consul"),
				ModuleVersion: ptr.String("1.0.0"),
			},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.runs.On("CreateRun", mock.Anything, &pb.CreateRunRequest{
					WorkspaceId:   "ws1",
					ModuleSource:  ptr.String("registry.terraform.io/hashicorp/consul"),
					ModuleVersion: ptr.String("1.0.0"),
				}).Return(&pb.Run{
					Metadata:      &pb.ResourceMetadata{Id: "r1", Trn: "trn:run:r1"},
					Status:        "pending",
					WorkspaceId:   "ws1",
					ModuleSource:  ptr.String("registry.terraform.io/hashicorp/consul"),
					ModuleVersion: ptr.String("1.0.0"),
				}, nil)
			},
			expectID: "r1",
		},
		{
			name: "create run failed",
			input: &createRunInput{
				WorkspaceID: "nonexistent",
			},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeWorkspace).Return(nil)
				m.runs.On("CreateRun", mock.Anything, &pb.CreateRunRequest{
					WorkspaceId: "nonexistent",
				}).Return(nil, status.Error(codes.NotFound, "workspace not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &runMocks{
				runs: mocks.NewRunsClient(t),
				acl:  acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					RunsClient: testMocks.runs,
				},
				acl: testMocks.acl,
			}

			_, handler := createRun(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.Run.ID)
		})
	}
}

func TestApplyRun(t *testing.T) {
	type testCase struct {
		name        string
		input       *applyRunInput
		mockSetup   func(*runMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "apply run successfully",
			input: &applyRunInput{ID: "r1"},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "r1", trn.ResourceTypeRun).Return(nil)
				m.runs.On("ApplyRun", mock.Anything, &pb.ApplyRunRequest{RunId: "r1"}).Return(&pb.Run{
					Metadata:    &pb.ResourceMetadata{Id: "r1", Trn: "trn:run:r1"},
					Status:      "apply_queued",
					WorkspaceId: "ws1",
				}, nil)
			},
		},
		{
			name:  "acl denial",
			input: &applyRunInput{ID: "r1"},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "r1", trn.ResourceTypeRun).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name:  "apply run failed",
			input: &applyRunInput{ID: "nonexistent"},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeRun).Return(nil)
				m.runs.On("ApplyRun", mock.Anything, &pb.ApplyRunRequest{RunId: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &runMocks{
				runs: mocks.NewRunsClient(t),
				acl:  acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					RunsClient: testMocks.runs,
				},
				acl: testMocks.acl,
			}

			_, handler := applyRun(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, output)
		})
	}
}

func TestCancelRun(t *testing.T) {
	type testCase struct {
		name        string
		input       *cancelRunInput
		mockSetup   func(*runMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "cancel run successfully",
			input: &cancelRunInput{ID: "r1"},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "r1", trn.ResourceTypeRun).Return(nil)
				m.runs.On("CancelRun", mock.Anything, &pb.CancelRunRequest{Id: "r1"}).Return(&pb.Run{
					Metadata:    &pb.ResourceMetadata{Id: "r1", Trn: "trn:run:r1"},
					Status:      "canceled",
					WorkspaceId: "ws1",
				}, nil)
			},
		},
		{
			name:  "acl denial",
			input: &cancelRunInput{ID: "r1"},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "r1", trn.ResourceTypeRun).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name:  "force cancel run successfully",
			input: &cancelRunInput{ID: "r1", Force: ptr.Bool(true)},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "r1", trn.ResourceTypeRun).Return(nil)
				m.runs.On("CancelRun", mock.Anything, &pb.CancelRunRequest{Id: "r1", Force: ptr.Bool(true)}).Return(&pb.Run{
					Metadata:      &pb.ResourceMetadata{Id: "r1", Trn: "trn:run:r1"},
					Status:        "canceled",
					WorkspaceId:   "ws1",
					ForceCanceled: true,
				}, nil)
			},
		},
		{
			name:  "run not found",
			input: &cancelRunInput{ID: "nonexistent"},
			mockSetup: func(m *runMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeRun).Return(nil)
				m.runs.On("CancelRun", mock.Anything, &pb.CancelRunRequest{Id: "nonexistent"}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &runMocks{
				runs: mocks.NewRunsClient(t),
				acl:  acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					RunsClient: testMocks.runs,
				},
				acl: testMocks.acl,
			}

			_, handler := cancelRun(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, output.Success)
		})
	}
}
