package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestAssignManagedIdentity(t *testing.T) {
	workspacePath := "group/workspace"
	managedIdentityID := "mi-id"

	tests := []struct {
		name        string
		aclError    error
		expectError bool
		validate    func(*testing.T, assignManagedIdentityOutput)
	}{
		{
			name: "successful assignment",
			validate: func(t *testing.T, output assignManagedIdentityOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, managedIdentityID)
				assert.Contains(t, output.Message, workspacePath)
			},
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
			mockMI := tharsis.NewManagedIdentity(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, "trn:workspace:"+workspacePath, trn.ResourceTypeWorkspace).Return(tt.aclError)

			if tt.aclError == nil {
				mockClient.On("ManagedIdentities").Return(mockMI)
				mockMI.On("AssignManagedIdentityToWorkspace", mock.Anything, &sdktypes.AssignManagedIdentityInput{
					ManagedIdentityID: &managedIdentityID,
					WorkspacePath:     workspacePath,
				}).Return(&sdktypes.Workspace{}, nil)
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := assignManagedIdentity(tc)
			_, output, err := handler(t.Context(), nil, assignManagedIdentityInput{
				WorkspacePath:     workspacePath,
				ManagedIdentityID: managedIdentityID,
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

func TestUnassignManagedIdentity(t *testing.T) {
	workspacePath := "group/workspace"
	managedIdentityID := "mi-id"

	tests := []struct {
		name        string
		aclError    error
		expectError bool
		validate    func(*testing.T, unassignManagedIdentityOutput)
	}{
		{
			name: "successful unassignment",
			validate: func(t *testing.T, output unassignManagedIdentityOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, managedIdentityID)
				assert.Contains(t, output.Message, workspacePath)
			},
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
			mockMI := tharsis.NewManagedIdentity(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, "trn:workspace:"+workspacePath, trn.ResourceTypeWorkspace).Return(tt.aclError)

			if tt.aclError == nil {
				mockClient.On("ManagedIdentities").Return(mockMI)
				mockMI.On("UnassignManagedIdentityFromWorkspace", mock.Anything, &sdktypes.AssignManagedIdentityInput{
					ManagedIdentityID: &managedIdentityID,
					WorkspacePath:     workspacePath,
				}).Return(&sdktypes.Workspace{}, nil)
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := unassignManagedIdentity(tc)
			_, output, err := handler(t.Context(), nil, unassignManagedIdentityInput{
				WorkspacePath:     workspacePath,
				ManagedIdentityID: managedIdentityID,
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
