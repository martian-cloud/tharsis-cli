package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestListGroups(t *testing.T) {
	type testCase struct {
		name     string
		input    listGroupsInput
		groups   []sdktypes.Group
		pageInfo sdktypes.PageInfo
		validate func(*testing.T, listGroupsOutput)
	}

	tests := []testCase{
		{
			name:  "list groups without filter",
			input: listGroupsInput{},
			groups: []sdktypes.Group{
				{
					Metadata:    sdktypes.ResourceMetadata{ID: "group-1"},
					Name:        "group-1",
					FullPath:    "parent/group-1",
					Description: "First group",
				},
				{
					Metadata:    sdktypes.ResourceMetadata{ID: "group-2"},
					Name:        "group-2",
					FullPath:    "parent/group-2",
					Description: "Second group",
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			validate: func(t *testing.T, output listGroupsOutput) {
				assert.Len(t, output.Groups, 2)
				assert.Equal(t, "group-1", output.Groups[0].ID)
				assert.Equal(t, "parent/group-1", output.Groups[0].Path)
				assert.False(t, output.PageInfo.HasNextPage)
			},
		},
		{
			name: "list groups with parent filter",
			input: listGroupsInput{
				ParentPath: ptr.String("parent"),
			},
			groups: []sdktypes.Group{
				{
					Metadata: sdktypes.ResourceMetadata{ID: "group-1"},
					Name:     "group-1",
					FullPath: "parent/group-1",
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			validate: func(t *testing.T, output listGroupsOutput) {
				assert.Len(t, output.Groups, 1)
			},
		},
		{
			name: "list groups with pagination",
			input: listGroupsInput{
				Limit:  ptr.Int32(10),
				Cursor: ptr.String("cursor-123"),
			},
			groups: []sdktypes.Group{
				{
					Metadata: sdktypes.ResourceMetadata{ID: "group-3"},
					Name:     "group-3",
					FullPath: "parent/group-3",
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: true,
				Cursor:      "cursor-456",
			},
			validate: func(t *testing.T, output listGroupsOutput) {
				assert.Len(t, output.Groups, 1)
				assert.True(t, output.PageInfo.HasNextPage)
				assert.Equal(t, "cursor-456", output.PageInfo.Cursor)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockGroup := tharsis.NewGroup(t)
			mockACL := acl.NewMockChecker(t)

			mockClient.On("Groups").Return(mockGroup)
			mockGroup.On("GetGroups", mock.Anything, mock.Anything).Return(
				&sdktypes.GetGroupsOutput{
					Groups:   tt.groups,
					PageInfo: &tt.pageInfo,
				}, nil)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := listGroups(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestGetGroup(t *testing.T) {
	groupID := "test-group-id"

	type testCase struct {
		name        string
		group       *sdktypes.Group
		aclError    error
		expectError bool
		validate    func(*testing.T, getGroupOutput)
	}

	tests := []testCase{
		{
			name: "successful group retrieval",
			group: &sdktypes.Group{
				Metadata: sdktypes.ResourceMetadata{
					ID: groupID,
				},
				FullPath:    "parent/test-group",
				Name:        "test-group",
				Description: "Test group description",
			},
			validate: func(t *testing.T, output getGroupOutput) {
				assert.Equal(t, groupID, output.Group.ID)
				assert.Equal(t, "parent/test-group", output.Group.Path)
				assert.Equal(t, "test-group", output.Group.Name)
				assert.Equal(t, "Test group description", output.Group.Description)
			},
		},
		{
			name: "group with empty description",
			group: &sdktypes.Group{
				Metadata: sdktypes.ResourceMetadata{
					ID: groupID,
				},
				FullPath: "org/team",
				Name:     "team",
			},
			validate: func(t *testing.T, output getGroupOutput) {
				assert.Equal(t, "team", output.Group.Name)
				assert.Equal(t, "", output.Group.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockGroup := tharsis.NewGroup(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Groups").Return(mockGroup)
				mockGroup.On("GetGroup", mock.Anything, &sdktypes.GetGroupInput{ID: &groupID}).Return(tt.group, nil)
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := getGroup(tc)
			_, output, err := handler(t.Context(), nil, getGroupInput{ID: groupID})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestCreateGroup(t *testing.T) {
	tests := []struct {
		name        string
		input       createGroupInput
		group       *sdktypes.Group
		aclError    error
		expectError bool
		validate    func(*testing.T, createGroupOutput)
	}{
		{
			name: "successful group creation",
			input: createGroupInput{
				Name:       "new-group",
				ParentPath: "parent",
			},
			group: &sdktypes.Group{
				Metadata: sdktypes.ResourceMetadata{ID: "group-id"},
				FullPath: "parent/new-group",
				Name:     "new-group",
			},
			validate: func(t *testing.T, output createGroupOutput) {
				assert.Equal(t, "group-id", output.Group.ID)
				assert.Equal(t, "parent/new-group", output.Group.Path)
				assert.Equal(t, "new-group", output.Group.Name)
			},
		},
		{
			name: "ACL denial",
			input: createGroupInput{
				Name:       "new-group",
				ParentPath: "parent",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockGroup := tharsis.NewGroup(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Groups").Return(mockGroup)
				mockGroup.On("CreateGroup", mock.Anything, &sdktypes.CreateGroupInput{
					Name:        tt.input.Name,
					ParentPath:  &tt.input.ParentPath,
					Description: tt.input.Description,
				}).Return(tt.group, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, "trn:group:"+tt.input.ParentPath, trn.ResourceTypeGroup).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := createGroup(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestUpdateGroup(t *testing.T) {
	groupID := "group-id"

	tests := []struct {
		name        string
		input       updateGroupInput
		group       *sdktypes.Group
		aclError    error
		expectError bool
		validate    func(*testing.T, updateGroupOutput)
	}{
		{
			name: "successful group update",
			input: updateGroupInput{
				ID:          groupID,
				Description: "Updated description",
			},
			group: &sdktypes.Group{
				Metadata:    sdktypes.ResourceMetadata{ID: groupID},
				FullPath:    "parent/group",
				Name:        "group",
				Description: "Updated description",
			},
			validate: func(t *testing.T, output updateGroupOutput) {
				assert.Equal(t, groupID, output.Group.ID)
				assert.Equal(t, "Updated description", output.Group.Description)
			},
		},
		{
			name: "ACL denial",
			input: updateGroupInput{
				ID:          groupID,
				Description: "Updated description",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockGroup := tharsis.NewGroup(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, groupID, trn.ResourceTypeGroup).Return(tt.aclError)

			if tt.aclError == nil {
				mockClient.On("Groups").Return(mockGroup)
				mockGroup.On("UpdateGroup", mock.Anything, &sdktypes.UpdateGroupInput{
					ID:          &tt.input.ID,
					Description: tt.input.Description,
				}).Return(tt.group, nil)
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := updateGroup(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestDeleteGroup(t *testing.T) {
	groupID := "test-group-id"

	type testCase struct {
		name        string
		topLevel    bool
		aclError    error
		expectError bool
		validate    func(*testing.T, deleteGroupOutput)
	}

	tests := []testCase{
		{
			name: "successful group deletion",
			validate: func(t *testing.T, output deleteGroupOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, "deleted successfully")
			},
		},
		{
			name:        "prevent top-level group deletion",
			topLevel:    true,
			expectError: true,
		},
		{
			name:        "ACL denial",
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockGroup := tharsis.NewGroup(t)
			mockACL := acl.NewMockChecker(t)

			mockClient.On("Groups").Return(mockGroup)

			// Mock GetGroup with appropriate path (always called first)
			groupPath := "parent/child"
			if tt.topLevel {
				groupPath = "toplevel"
			}
			mockGroup.On("GetGroup", mock.Anything, &sdktypes.GetGroupInput{ID: &groupID}).
				Return(&sdktypes.Group{
					Metadata: sdktypes.ResourceMetadata{ID: groupID},
					FullPath: groupPath,
					Name:     "child",
				}, nil)

			// Only expect ACL and DeleteGroup calls if not top-level and no ACL error
			if !tt.topLevel {
				mockACL.On("Authorize", mock.Anything, mockClient, groupID, trn.ResourceTypeGroup).Return(tt.aclError)

				if tt.aclError == nil {
					mockGroup.On("DeleteGroup", mock.Anything, &sdktypes.DeleteGroupInput{ID: &groupID}).
						Return(nil)
				}
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := deleteGroup(tc)
			_, output, err := handler(t.Context(), nil, deleteGroupInput{ID: groupID})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}
