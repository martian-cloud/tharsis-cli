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

func TestListWorkspaces(t *testing.T) {
	type testCase struct {
		name        string
		input       listWorkspacesInput
		workspaces  []sdktypes.Workspace
		pageInfo    sdktypes.PageInfo
		expectError bool
		validate    func(*testing.T, listWorkspacesOutput)
	}

	tests := []testCase{
		{
			name:  "list workspaces without filter",
			input: listWorkspacesInput{},
			workspaces: []sdktypes.Workspace{
				{
					Metadata:           sdktypes.ResourceMetadata{ID: "ws-1"},
					Name:               "workspace-1",
					FullPath:           "group1/workspace-1",
					GroupPath:          "group1",
					Description:        "First workspace",
					TerraformVersion:   "1.5.0",
					MaxJobDuration:     60,
					PreventDestroyPlan: false,
					Labels:             map[string]string{"env": "dev"},
				},
				{
					Metadata:           sdktypes.ResourceMetadata{ID: "ws-2"},
					Name:               "workspace-2",
					FullPath:           "group2/workspace-2",
					GroupPath:          "group2",
					Description:        "Second workspace",
					TerraformVersion:   "1.6.0",
					MaxJobDuration:     120,
					PreventDestroyPlan: true,
					Labels:             map[string]string{"env": "prod"},
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			validate: func(t *testing.T, output listWorkspacesOutput) {
				assert.Len(t, output.Workspaces, 2)
				assert.Equal(t, "ws-1", output.Workspaces[0].ID)
				assert.Equal(t, "workspace-1", output.Workspaces[0].Name)
				assert.Equal(t, "group1/workspace-1", output.Workspaces[0].Path)
				assert.Equal(t, "ws-2", output.Workspaces[1].ID)
				assert.False(t, output.PageInfo.HasNextPage)
			},
		},
		{
			name: "list workspaces with group filter",
			input: listWorkspacesInput{
				GroupPath: ptr.String("group1"),
			},
			workspaces: []sdktypes.Workspace{
				{
					Metadata:           sdktypes.ResourceMetadata{ID: "ws-1"},
					Name:               "workspace-1",
					FullPath:           "group1/workspace-1",
					GroupPath:          "group1",
					Description:        "Workspace in group1",
					TerraformVersion:   "1.5.0",
					MaxJobDuration:     60,
					PreventDestroyPlan: false,
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			validate: func(t *testing.T, output listWorkspacesOutput) {
				assert.Len(t, output.Workspaces, 1)
				assert.Equal(t, "group1", output.Workspaces[0].GroupPath)
			},
		},
		{
			name: "list workspaces with pagination",
			input: listWorkspacesInput{
				Limit:  ptr.Int32(10),
				Cursor: ptr.String("cursor-123"),
			},
			workspaces: []sdktypes.Workspace{
				{
					Metadata:         sdktypes.ResourceMetadata{ID: "ws-3"},
					Name:             "workspace-3",
					FullPath:         "group1/workspace-3",
					GroupPath:        "group1",
					Description:      "Third workspace",
					TerraformVersion: "1.5.0",
					MaxJobDuration:   60,
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: true,
				Cursor:      "cursor-456",
			},
			validate: func(t *testing.T, output listWorkspacesOutput) {
				assert.Len(t, output.Workspaces, 1)
				assert.True(t, output.PageInfo.HasNextPage)
				assert.Equal(t, "cursor-456", output.PageInfo.Cursor)
			},
		},
		{
			name:       "empty result",
			input:      listWorkspacesInput{},
			workspaces: []sdktypes.Workspace{},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			validate: func(t *testing.T, output listWorkspacesOutput) {
				assert.Len(t, output.Workspaces, 0)
				assert.False(t, output.PageInfo.HasNextPage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockACL := acl.NewMockChecker(t)

			mockClient.On("Workspaces").Return(mockWorkspaces)
			mockWorkspaces.On("GetWorkspaces", mock.Anything, mock.Anything).Return(
				&sdktypes.GetWorkspacesOutput{
					Workspaces: tt.workspaces,
					PageInfo:   &tt.pageInfo,
				}, nil)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := listWorkspaces(tc)
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

func TestGetWorkspace(t *testing.T) {
	workspaceID := "test-workspace-id"

	type testCase struct {
		name        string
		workspace   *sdktypes.Workspace
		aclError    error
		expectError bool
		validate    func(*testing.T, getWorkspaceOutput)
	}

	tests := []testCase{
		{
			name: "successful workspace retrieval",
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{
					ID: workspaceID,
				},
				FullPath:         "test-group/test-workspace",
				Name:             "test-workspace",
				GroupPath:        "test-group",
				Description:      "Test workspace",
				TerraformVersion: "1.5.0",
				MaxJobDuration:   60,
			},
			validate: func(t *testing.T, output getWorkspaceOutput) {
				assert.Equal(t, workspaceID, output.Workspace.ID)
				assert.Equal(t, "test-group/test-workspace", output.Workspace.Path)
				assert.Equal(t, "test-workspace", output.Workspace.Name)
				assert.Equal(t, "test-group", output.Workspace.GroupPath)
				assert.Equal(t, "Test workspace", output.Workspace.Description)
				assert.Equal(t, "1.5.0", output.Workspace.TerraformVersion)
				assert.Equal(t, int32(60), output.Workspace.MaxJobDuration)
			},
		},
		{
			name: "workspace with labels",
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{
					ID: workspaceID,
				},
				FullPath:           "org/prod-workspace",
				Name:               "prod-workspace",
				GroupPath:          "org",
				PreventDestroyPlan: true,
				Labels:             map[string]string{"env": "production", "team": "platform"},
			},
			validate: func(t *testing.T, output getWorkspaceOutput) {
				assert.Equal(t, "prod-workspace", output.Workspace.Name)
				assert.True(t, output.Workspace.PreventDestroyPlan)
				assert.Len(t, output.Workspace.Labels, 2)
				assert.Equal(t, "production", output.Workspace.Labels["env"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &workspaceID}).Return(tt.workspace, nil)
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := getWorkspace(tc)
			_, output, err := handler(t.Context(), nil, getWorkspaceInput{ID: workspaceID})

			if tt.expectError || tt.aclError != nil {
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

func TestCreateWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		input       createWorkspaceInput
		workspace   *sdktypes.Workspace
		aclError    error
		expectError bool
		validate    func(*testing.T, createWorkspaceOutput)
	}{
		{
			name: "successful workspace creation",
			input: createWorkspaceInput{
				Name:      "new-workspace",
				GroupPath: "test-group",
			},
			workspace: &sdktypes.Workspace{
				Metadata:  sdktypes.ResourceMetadata{ID: "ws-id"},
				FullPath:  "test-group/new-workspace",
				Name:      "new-workspace",
				GroupPath: "test-group",
			},
			validate: func(t *testing.T, output createWorkspaceOutput) {
				assert.Equal(t, "ws-id", output.Workspace.ID)
				assert.Equal(t, "test-group/new-workspace", output.Workspace.Path)
				assert.Equal(t, "new-workspace", output.Workspace.Name)
			},
		},
		{
			name: "ACL denies group access",
			input: createWorkspaceInput{
				Name:      "new-workspace",
				GroupPath: "restricted-group",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)

			var labels []sdktypes.WorkspaceLabelInput
			for k, v := range tt.input.Labels {
				labels = append(labels, sdktypes.WorkspaceLabelInput{Key: k, Value: v})
			}

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("CreateWorkspace", mock.Anything, &sdktypes.CreateWorkspaceInput{
					Name:               tt.input.Name,
					GroupPath:          tt.input.GroupPath,
					Description:        tt.input.Description,
					TerraformVersion:   tt.input.TerraformVersion,
					MaxJobDuration:     tt.input.MaxJobDuration,
					PreventDestroyPlan: tt.input.PreventDestroyPlan,
					Labels:             labels,
				}).Return(tt.workspace, nil)
			}

			mockACL := acl.NewMockChecker(t)
			mockACL.On("Authorize", mock.Anything, mockClient, "trn:group:"+tt.input.GroupPath, trn.ResourceTypeGroup).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := createWorkspace(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			if tt.expectError || tt.aclError != nil {
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

func TestUpdateWorkspace(t *testing.T) {
	workspaceID := "ws-id"

	tests := []struct {
		name        string
		input       updateWorkspaceInput
		workspace   *sdktypes.Workspace
		aclError    error
		expectError bool
		validate    func(*testing.T, updateWorkspaceOutput)
	}{
		{
			name: "successful workspace update",
			input: updateWorkspaceInput{
				ID:          workspaceID,
				Description: "Updated description",
			},
			workspace: &sdktypes.Workspace{
				Metadata:    sdktypes.ResourceMetadata{ID: workspaceID},
				FullPath:    "group/workspace",
				Name:        "workspace",
				GroupPath:   "group",
				Description: "Updated description",
			},
			validate: func(t *testing.T, output updateWorkspaceOutput) {
				assert.Equal(t, workspaceID, output.Workspace.ID)
				assert.Equal(t, "Updated description", output.Workspace.Description)
			},
		},
		{
			name: "ACL denies access",
			input: updateWorkspaceInput{
				ID: workspaceID,
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockACL := acl.NewMockChecker(t)

			var labels []sdktypes.WorkspaceLabelInput
			for k, v := range tt.input.Labels {
				labels = append(labels, sdktypes.WorkspaceLabelInput{Key: k, Value: v})
			}

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("UpdateWorkspace", mock.Anything, &sdktypes.UpdateWorkspaceInput{
					ID:                 &tt.input.ID,
					Description:        tt.input.Description,
					TerraformVersion:   tt.input.TerraformVersion,
					MaxJobDuration:     tt.input.MaxJobDuration,
					PreventDestroyPlan: tt.input.PreventDestroyPlan,
					Labels:             labels,
				}).Return(tt.workspace, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, tt.input.ID, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := updateWorkspace(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			if tt.expectError || tt.aclError != nil {
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

func TestDeleteWorkspace(t *testing.T) {
	workspaceID := "ws-id"

	tests := []struct {
		name        string
		aclError    error
		expectError bool
		validate    func(*testing.T, deleteWorkspaceOutput)
	}{
		{
			name: "successful workspace deletion",
			validate: func(t *testing.T, output deleteWorkspaceOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, "deleted successfully")
			},
		},
		{
			name:        "ACL denies access",
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("DeleteWorkspace", mock.Anything, &sdktypes.DeleteWorkspaceInput{ID: &workspaceID}).Return(nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, workspaceID, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := deleteWorkspace(tc)
			_, output, err := handler(t.Context(), nil, deleteWorkspaceInput{ID: workspaceID})

			if tt.expectError || tt.aclError != nil {
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

func TestGetWorkspaceOutputs(t *testing.T) {
	workspaceID := "ws-id"

	tests := []struct {
		name        string
		workspace   *sdktypes.Workspace
		stateVer    *sdktypes.StateVersion
		aclError    error
		expectError bool
		validate    func(*testing.T, getWorkspaceOutputsOutput)
	}{
		{
			name: "workspace with no state version",
			workspace: &sdktypes.Workspace{
				Metadata:            sdktypes.ResourceMetadata{ID: workspaceID},
				CurrentStateVersion: nil,
			},
			validate: func(t *testing.T, output getWorkspaceOutputsOutput) {
				assert.Empty(t, output.Outputs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &workspaceID}).Return(tt.workspace, nil)

				if tt.stateVer != nil {
					mockStateVersion := tharsis.NewStateVersion(t)
					mockClient.On("StateVersions").Return(mockStateVersion)
					mockStateVersion.On("GetStateVersion", mock.Anything, &sdktypes.GetStateVersionInput{
						ID: tt.workspace.CurrentStateVersion.Metadata.ID,
					}).Return(tt.stateVer, nil)
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

			_, handler := getWorkspaceOutputs(tc)
			_, output, err := handler(t.Context(), nil, getWorkspaceOutputsInput{ID: workspaceID})

			if tt.expectError || tt.aclError != nil {
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
