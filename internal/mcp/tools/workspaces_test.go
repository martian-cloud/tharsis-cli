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

type workspaceMocks struct {
	workspaces *mocks.WorkspacesClient
	groups     *mocks.GroupsClient
	acl        *acl.MockChecker
}

func TestListWorkspaces(t *testing.T) {
	type testCase struct {
		name          string
		input         *listWorkspacesInput
		mockSetup     func(*workspaceMocks)
		expectError   bool
		expectResults int
	}

	testCases := []testCase{
		{
			name:  "list workspaces successfully",
			input: &listWorkspacesInput{Limit: ptr.Int32(10)},
			mockSetup: func(m *workspaceMocks) {
				m.workspaces.On("GetWorkspaces", mock.Anything, &pb.GetWorkspacesRequest{
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(&pb.GetWorkspacesResponse{
					Workspaces: []*pb.Workspace{
						{Metadata: &pb.ResourceMetadata{Id: "ws1", Trn: "trn:workspace:group/ws1"}, Name: "workspace1", FullPath: "group/workspace1"},
					},
					PageInfo: &pb.PageInfo{HasNextPage: false},
				}, nil)
			},
			expectResults: 1,
		},
		{
			name:  "grpc error",
			input: &listWorkspacesInput{Limit: ptr.Int32(10)},
			mockSetup: func(m *workspaceMocks) {
				m.workspaces.On("GetWorkspaces", mock.Anything, &pb.GetWorkspacesRequest{
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &workspaceMocks{
				workspaces: mocks.NewWorkspacesClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					WorkspacesClient: testMocks.workspaces,
				},
			}

			_, handler := listWorkspaces(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, output.Workspaces, tc.expectResults)
		})
	}
}

func TestGetWorkspace(t *testing.T) {
	type testCase struct {
		name        string
		input       *getWorkspaceInput
		mockSetup   func(*workspaceMocks)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name:  "get workspace successfully",
			input: &getWorkspaceInput{ID: "ws1"},
			mockSetup: func(m *workspaceMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata:    &pb.ResourceMetadata{Id: "ws1", Trn: "trn:workspace:group/ws1"},
					Name:        "workspace1",
					FullPath:    "group/workspace1",
					Description: "test workspace",
				}, nil)
			},
			expectID: "ws1",
		},
		{
			name:  "workspace not found",
			input: &getWorkspaceInput{ID: "nonexistent"},
			mockSetup: func(m *workspaceMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &workspaceMocks{
				workspaces: mocks.NewWorkspacesClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					WorkspacesClient: testMocks.workspaces,
				},
			}

			_, handler := getWorkspace(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.Workspace.ID)
		})
	}
}

func TestCreateWorkspace(t *testing.T) {
	type testCase struct {
		name        string
		input       *createWorkspaceInput
		mockSetup   func(*workspaceMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "create workspace successfully",
			input: &createWorkspaceInput{
				GroupID:     "group1",
				Name:        "new-workspace",
				Description: "test workspace",
			},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "group1", trn.ResourceTypeGroup).Return(nil)
				m.workspaces.On("CreateWorkspace", mock.Anything, &pb.CreateWorkspaceRequest{
					GroupId:     "group1",
					Name:        "new-workspace",
					Description: "test workspace",
				}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1", Trn: "trn:workspace:group1/new-workspace"},
					Name:     "new-workspace",
					FullPath: "group1/new-workspace",
				}, nil)
			},
		},
		{
			name: "acl denial",
			input: &createWorkspaceInput{
				GroupID: "group1",
				Name:    "new-workspace",
			},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "group1", trn.ResourceTypeGroup).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &workspaceMocks{
				workspaces: mocks.NewWorkspacesClient(t),
				acl:        acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					GroupsClient:     testMocks.groups,
					WorkspacesClient: testMocks.workspaces,
				},
				acl: testMocks.acl,
			}

			_, handler := createWorkspace(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
			require.NotNil(t, output.Workspace)
		})
	}
}

func TestUpdateWorkspace(t *testing.T) {
	type testCase struct {
		name        string
		input       *updateWorkspaceInput
		mockSetup   func(*workspaceMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "update workspace successfully",
			input: &updateWorkspaceInput{
				ID:          "ws1",
				Description: ptr.String("updated description"),
			},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.workspaces.On("UpdateWorkspace", mock.Anything, &pb.UpdateWorkspaceRequest{
					Id:          "ws1",
					Description: ptr.String("updated description"),
				}).Return(&pb.Workspace{
					Metadata:    &pb.ResourceMetadata{Id: "ws1", Trn: "trn:workspace:group/ws1"},
					Name:        "workspace1",
					Description: "updated description",
				}, nil)
			},
		},
		{
			name: "acl denial",
			input: &updateWorkspaceInput{
				ID:          "ws1",
				Description: ptr.String("updated"),
			},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "workspace not found",
			input: &updateWorkspaceInput{
				ID:          "nonexistent",
				Description: ptr.String("updated"),
			},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeWorkspace).Return(nil)
				m.workspaces.On("UpdateWorkspace", mock.Anything, &pb.UpdateWorkspaceRequest{
					Id:          "nonexistent",
					Description: ptr.String("updated"),
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &workspaceMocks{
				groups:     mocks.NewGroupsClient(t),
				workspaces: mocks.NewWorkspacesClient(t),
				acl:        acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					GroupsClient:     testMocks.groups,
					WorkspacesClient: testMocks.workspaces,
				},
				acl: testMocks.acl,
			}

			_, handler := updateWorkspace(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
		})
	}
}

func TestDeleteWorkspace(t *testing.T) {
	type testCase struct {
		name        string
		input       *deleteWorkspaceInput
		mockSetup   func(*workspaceMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "delete workspace successfully",
			input: &deleteWorkspaceInput{ID: "ws1"},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.workspaces.On("DeleteWorkspace", mock.Anything, &pb.DeleteWorkspaceRequest{Id: "ws1"}).Return(nil, nil)
			},
		},
		{
			name:  "acl denial",
			input: &deleteWorkspaceInput{ID: "ws1"},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name:  "workspace not found",
			input: &deleteWorkspaceInput{ID: "nonexistent"},
			mockSetup: func(m *workspaceMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeWorkspace).Return(nil)
				m.workspaces.On("DeleteWorkspace", mock.Anything, &pb.DeleteWorkspaceRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &workspaceMocks{
				groups:     mocks.NewGroupsClient(t),
				workspaces: mocks.NewWorkspacesClient(t),
				acl:        acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					GroupsClient:     testMocks.groups,
					WorkspacesClient: testMocks.workspaces,
				},
				acl: testMocks.acl,
			}

			_, handler := deleteWorkspace(toolCtx)
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
