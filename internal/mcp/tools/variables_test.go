package tools

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestSetVariable(t *testing.T) {
	namespaceID := "workspace-id"

	tests := []struct {
		name        string
		input       setVariableInput
		workspace   *sdktypes.Workspace
		variable    *sdktypes.NamespaceVariable
		aclError    error
		expectError bool
		validate    func(*testing.T, setVariableOutput)
	}{
		{
			name: "create new terraform variable",
			input: setVariableInput{
				NamespaceID: namespaceID,
				Key:         "region",
				Value:       "us-east-1",
				Category:    "terraform",
			},
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			validate: func(t *testing.T, output setVariableOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, "region")
			},
		},
		{
			name: "update existing environment variable",
			input: setVariableInput{
				NamespaceID: namespaceID,
				Key:         "ENV",
				Value:       "production",
				Category:    "environment",
			},
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			variable: &sdktypes.NamespaceVariable{
				Metadata: sdktypes.ResourceMetadata{ID: "var-id"},
				Key:      "ENV",
				Value:    ptr.String("staging"),
			},
			validate: func(t *testing.T, output setVariableOutput) {
				assert.True(t, output.Success)
			},
		},
		{
			name: "ACL denial",
			input: setVariableInput{
				NamespaceID: namespaceID,
				Key:         "region",
				Value:       "us-east-1",
				Category:    "terraform",
			},
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockVariable := tharsis.NewVariable(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockClient.On("Variables").Return(mockVariable)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)

				category := sdktypes.TerraformVariableCategory
				if tt.input.Category == "environment" {
					category = sdktypes.EnvironmentVariableCategory
				}
				variableID := fmt.Sprintf("trn:variable:%s/%s/%s", tt.workspace.FullPath, category, tt.input.Key)

				mockVariable.On("GetVariable", mock.Anything, &sdktypes.GetNamespaceVariableInput{ID: variableID}).Return(tt.variable, nil)

				if tt.variable != nil {
					mockVariable.On("UpdateVariable", mock.Anything, &sdktypes.UpdateNamespaceVariableInput{
						ID:    tt.variable.Metadata.ID,
						Key:   tt.variable.Key,
						Value: tt.input.Value,
					}).Return(&sdktypes.NamespaceVariable{}, nil)
				} else {
					mockVariable.On("CreateVariable", mock.Anything, &sdktypes.CreateNamespaceVariableInput{
						Key:           tt.input.Key,
						Value:         tt.input.Value,
						Category:      category,
						NamespacePath: tt.workspace.FullPath,
					}).Return(&sdktypes.NamespaceVariable{}, nil)
				}
			} else {
				// For ACL denial, still need to mock getNamespacePath
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, namespaceID, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := setVariable(tc)
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

func TestDeleteVariable(t *testing.T) {
	namespaceID := "workspace-id"

	tests := []struct {
		name        string
		input       deleteVariableInput
		workspace   *sdktypes.Workspace
		variable    *sdktypes.NamespaceVariable
		aclError    error
		expectError bool
		validate    func(*testing.T, deleteVariableOutput)
	}{
		{
			name: "successful variable deletion",
			input: deleteVariableInput{
				NamespaceID: namespaceID,
				Key:         "region",
				Category:    "terraform",
			},
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			variable: &sdktypes.NamespaceVariable{
				Metadata: sdktypes.ResourceMetadata{ID: "var-id"},
				Key:      "region",
			},
			validate: func(t *testing.T, output deleteVariableOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, "region")
			},
		},
		{
			name: "ACL denial",
			input: deleteVariableInput{
				NamespaceID: namespaceID,
				Key:         "region",
				Category:    "terraform",
			},
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockVariable := tharsis.NewVariable(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockClient.On("Variables").Return(mockVariable)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)

				category := sdktypes.TerraformVariableCategory
				if tt.input.Category == "environment" {
					category = sdktypes.EnvironmentVariableCategory
				}
				variableID := fmt.Sprintf("trn:variable:%s/%s/%s", tt.workspace.FullPath, category, tt.input.Key)

				mockVariable.On("GetVariable", mock.Anything, &sdktypes.GetNamespaceVariableInput{ID: variableID}).Return(tt.variable, nil)
				mockVariable.On("DeleteVariable", mock.Anything, &sdktypes.DeleteNamespaceVariableInput{ID: tt.variable.Metadata.ID}).Return(nil)
			} else {
				// For ACL denial, still need to mock getNamespacePath
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, namespaceID, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := deleteVariable(tc)
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

func TestSetTerraformVariablesFromFile(t *testing.T) {
	namespaceID := "workspace-id"

	tests := []struct {
		name        string
		fileContent string
		workspace   *sdktypes.Workspace
		aclError    error
		expectError bool
		validate    func(*testing.T, setTerraformVariablesFromFileOutput)
	}{
		{
			name: "successful terraform variables from file",
			fileContent: `region = "us-east-1"
instance_type = "t2.micro"
count = 3`,
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			validate: func(t *testing.T, output setTerraformVariablesFromFileOutput) {
				assert.True(t, output.Success)
				assert.Equal(t, 3, output.Count)
				assert.Contains(t, output.Message, "3 Terraform variables")
			},
		},
		{
			name:        "ACL denial",
			fileContent: `region = "us-east-1"`,
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := fmt.Sprintf("%s/test.tfvars", tmpDir)
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0600)
			assert.NoError(t, err)

			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockVariable := tharsis.NewVariable(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockClient.On("Variables").Return(mockVariable)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)
				mockVariable.On("SetVariables", mock.Anything, mock.MatchedBy(func(input *sdktypes.SetNamespaceVariablesInput) bool {
					return input.NamespacePath == tt.workspace.FullPath && input.Category == sdktypes.TerraformVariableCategory
				})).Return(nil)
			} else {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, namespaceID, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := setTerraformVariablesFromFile(tc)
			_, output, err := handler(t.Context(), nil, setTerraformVariablesFromFileInput{
				NamespaceID: namespaceID,
				FilePath:    filePath,
			})

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

func TestSetEnvironmentVariablesFromFile(t *testing.T) {
	namespaceID := "workspace-id"

	tests := []struct {
		name        string
		fileContent string
		workspace   *sdktypes.Workspace
		aclError    error
		expectError bool
		validate    func(*testing.T, setEnvironmentVariablesFromFileOutput)
	}{
		{
			name: "successful environment variables from file",
			fileContent: `ENV=production
DEBUG=false
PORT=8080`,
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			validate: func(t *testing.T, output setEnvironmentVariablesFromFileOutput) {
				assert.True(t, output.Success)
				assert.Equal(t, 3, output.Count)
				assert.Contains(t, output.Message, "3 environment variables")
			},
		},
		{
			name:        "ACL denial",
			fileContent: `ENV=production`,
			workspace: &sdktypes.Workspace{
				Metadata: sdktypes.ResourceMetadata{ID: namespaceID},
				FullPath: "group/workspace",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := fmt.Sprintf("%s/test.env", tmpDir)
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0600)
			assert.NoError(t, err)

			mockClient := tharsis.NewMockClient(t)
			mockWorkspaces := tharsis.NewWorkspaces(t)
			mockVariable := tharsis.NewVariable(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockClient.On("Variables").Return(mockVariable)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)
				mockVariable.On("SetVariables", mock.Anything, mock.MatchedBy(func(input *sdktypes.SetNamespaceVariablesInput) bool {
					return input.NamespacePath == tt.workspace.FullPath && input.Category == sdktypes.EnvironmentVariableCategory
				})).Return(nil)
			} else {
				mockClient.On("Workspaces").Return(mockWorkspaces)
				mockWorkspaces.On("GetWorkspace", mock.Anything, &sdktypes.GetWorkspaceInput{ID: &namespaceID}).Return(tt.workspace, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, namespaceID, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := setEnvironmentVariablesFromFile(tc)
			_, output, err := handler(t.Context(), nil, setEnvironmentVariablesFromFileInput{
				NamespaceID: namespaceID,
				FilePath:    filePath,
			})

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
