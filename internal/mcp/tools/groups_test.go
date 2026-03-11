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

type groupMocks struct {
	groups *mocks.GroupsClient
	acl    *acl.MockChecker
}

func TestListGroups(t *testing.T) {
	type testCase struct {
		name          string
		input         *listGroupsInput
		mockSetup     func(*groupMocks)
		expectError   bool
		expectResults int
	}

	testCases := []testCase{
		{
			name:  "list groups successfully",
			input: &listGroupsInput{Limit: ptr.Int32(10)},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroups", mock.Anything, &pb.GetGroupsRequest{
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(&pb.GetGroupsResponse{
					Groups: []*pb.Group{
						{Metadata: &pb.ResourceMetadata{
							Id:  "g1",
							Trn: "trn:group:group1",
						},
							Name:     "group1",
							FullPath: "group1",
						},
					},
					PageInfo: &pb.PageInfo{HasNextPage: false},
				}, nil)
			},
			expectResults: 1,
		},
		{
			name:  "list groups with parent filter",
			input: &listGroupsInput{ParentID: ptr.String("parent1"), Limit: ptr.Int32(10)},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroups", mock.Anything, &pb.GetGroupsRequest{
					ParentId:          ptr.String("parent1"),
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(&pb.GetGroupsResponse{
					Groups: []*pb.Group{
						{Metadata: &pb.ResourceMetadata{
							Id:  "g1",
							Trn: "trn:group:parent1/child1",
						},
							Name:     "child1",
							FullPath: "parent1/child1",
						},
					},
					PageInfo: &pb.PageInfo{HasNextPage: false},
				}, nil)
			},
			expectResults: 1,
		},
		{
			name:  "grpc error",
			input: &listGroupsInput{Limit: ptr.Int32(10)},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroups", mock.Anything, &pb.GetGroupsRequest{
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &groupMocks{
				groups: mocks.NewGroupsClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					GroupsClient: testMocks.groups,
				},
			}

			_, handler := listGroups(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, output.Groups, tc.expectResults)
		})
	}
}

func TestGetGroup(t *testing.T) {
	type testCase struct {
		name        string
		input       *getGroupInput
		mockSetup   func(*groupMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "get group successfully",
			input: &getGroupInput{ID: "g1"},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "g1"}).Return(&pb.Group{
					Metadata: &pb.ResourceMetadata{
						Id:  "g1",
						Trn: "trn:group:group1",
					},
					Name:        "group1",
					FullPath:    "group1",
					Description: "test group",
				}, nil)
			},
		},
		{
			name:  "group not found",
			input: &getGroupInput{ID: "nonexistent"},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &groupMocks{
				groups: mocks.NewGroupsClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					GroupsClient: testMocks.groups,
				},
			}

			_, handler := getGroup(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output.Group)
		})
	}
}

func TestCreateGroup(t *testing.T) {
	type testCase struct {
		name        string
		input       *createGroupInput
		mockSetup   func(*groupMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "create child group successfully",
			input: &createGroupInput{
				ParentID:    "trn:group:parent1",
				Name:        "child-group",
				Description: "test child group",
			},
			mockSetup: func(m *groupMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:group:parent1", trn.ResourceTypeGroup).Return(nil)
				m.groups.On("CreateGroup", mock.Anything, &pb.CreateGroupRequest{
					ParentId:    ptr.String("trn:group:parent1"),
					Name:        "child-group",
					Description: "test child group",
				}).Return(&pb.Group{
					Metadata: &pb.ResourceMetadata{Id: "g2", Trn: "trn:group:parent1/child-group"},
					Name:     "child-group",
					FullPath: "parent1/child-group",
				}, nil)
			},
		},
		{
			name: "create failed",
			input: &createGroupInput{
				ParentID:    "trn:group:parent1",
				Name:        "child-group",
				Description: "test child group",
			},
			mockSetup: func(m *groupMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:group:parent1", trn.ResourceTypeGroup).Return(nil)
				m.groups.On("CreateGroup", mock.Anything, &pb.CreateGroupRequest{
					ParentId:    ptr.String("trn:group:parent1"),
					Name:        "child-group",
					Description: "test child group",
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
		{
			name: "acl denial",
			input: &createGroupInput{
				ParentID: "p1",
				Name:     "child-group",
			},
			mockSetup: func(m *groupMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "p1", trn.ResourceTypeGroup).Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &groupMocks{
				groups: mocks.NewGroupsClient(t),
				acl:    acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				acl: testMocks.acl,
				grpcClient: &client.Client{
					GroupsClient: testMocks.groups,
				},
			}

			_, handler := createGroup(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output.Group)
		})
	}
}

func TestUpdateGroup(t *testing.T) {
	type testCase struct {
		name        string
		input       *updateGroupInput
		mockSetup   func(*groupMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "update group successfully",
			input: &updateGroupInput{
				ID:          "trn:group:g1/g2",
				Description: ptr.String("updated description"),
			},
			mockSetup: func(m *groupMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:group:g1/g2", trn.ResourceTypeGroup).Return(nil)
				m.groups.On("UpdateGroup", mock.Anything, &pb.UpdateGroupRequest{
					Id:          "trn:group:g1/g2",
					Description: ptr.String("updated description"),
				}).Return(&pb.Group{
					Metadata:    &pb.ResourceMetadata{Id: "g2", Trn: "trn:group:g1/g2"},
					Name:        "group1",
					Description: "updated description",
				}, nil)
			},
		},
		{
			name: "update failed",
			input: &updateGroupInput{
				ID:          "trn:group:g1/g2",
				Description: ptr.String("updated"),
			},
			mockSetup: func(m *groupMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:group:g1/g2", trn.ResourceTypeGroup).Return(nil)
				m.groups.On("UpdateGroup", mock.Anything, &pb.UpdateGroupRequest{
					Id:          "trn:group:g1/g2",
					Description: ptr.String("updated"),
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
		{
			name: "acl denial",
			input: &updateGroupInput{
				ID:          "trn:group:g1/g2",
				Description: ptr.String("updated"),
			},
			mockSetup: func(m *groupMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:group:g1/g2", trn.ResourceTypeGroup).Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &groupMocks{
				groups: mocks.NewGroupsClient(t),
				acl:    acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				acl: testMocks.acl,
				grpcClient: &client.Client{
					GroupsClient: testMocks.groups,
				},
			}

			_, handler := updateGroup(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output.Group)
			assert.Equal(t, output.Group.Description, *tc.input.Description)
		})
	}
}

func TestDeleteGroup(t *testing.T) {
	type testCase struct {
		name        string
		input       *deleteGroupInput
		mockSetup   func(*groupMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "delete group successfully",
			input: &deleteGroupInput{ID: "trn:group:g1/g2"},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:g1/g2"}).Return(&pb.Group{
					Metadata: &pb.ResourceMetadata{
						Id:  "g2",
						Trn: "trn:group:g1/g2",
					},
					FullPath: "g1/g2",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:group:g1/g2", trn.ResourceTypeGroup).Return(nil)
				m.groups.On("DeleteGroup", mock.Anything, &pb.DeleteGroupRequest{Id: "trn:group:g1/g2"}).Return(nil, nil)
			},
		},
		{
			name:  "denies root group deletion",
			input: &deleteGroupInput{ID: "trn:group:g1"},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:g1"}).Return(&pb.Group{
					Metadata: &pb.ResourceMetadata{
						Id:  "g1",
						Trn: "trn:group:g1",
					},
					FullPath: "g1",
				}, nil)
			},
			expectError: true,
		},
		{
			name:  "group not found",
			input: &deleteGroupInput{ID: "nonexistent"},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
		{
			name:  "acl denial",
			input: &deleteGroupInput{ID: "trn:group:g1/g2"},
			mockSetup: func(m *groupMocks) {
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:g1/g2"}).Return(&pb.Group{
					Metadata: &pb.ResourceMetadata{
						Id:  "g2",
						Trn: "trn:group:g1/g2",
					},
					FullPath: "g1/g2",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:group:g1/g2", trn.ResourceTypeGroup).Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &groupMocks{
				groups: mocks.NewGroupsClient(t),
				acl:    acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				acl: testMocks.acl,
				grpcClient: &client.Client{
					GroupsClient: testMocks.groups,
				},
			}

			_, handler := deleteGroup(toolCtx)
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
