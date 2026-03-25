package tools

import (
	"testing"

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

type managedIdentityMocks struct {
	managedIdentities *mocks.ManagedIdentitiesClient
	acl               *acl.MockChecker
}

func TestAssignManagedIdentity(t *testing.T) {
	type testCase struct {
		name        string
		input       *assignManagedIdentityInput
		mockSetup   func(*managedIdentityMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "assign managed identity successfully",
			input: &assignManagedIdentityInput{
				WorkspaceID:       "ws1",
				ManagedIdentityID: "mi1",
			},
			mockSetup: func(m *managedIdentityMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.managedIdentities.On("AssignManagedIdentityToWorkspace", mock.Anything, &pb.AssignManagedIdentityToWorkspaceRequest{
					ManagedIdentityId: "mi1",
					WorkspaceId:       "ws1",
				}).Return(nil, nil)
			},
		},
		{
			name: "acl denial",
			input: &assignManagedIdentityInput{
				WorkspaceID:       "ws1",
				ManagedIdentityID: "mi1",
			},
			mockSetup: func(m *managedIdentityMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "assign call unsuccessful",
			input: &assignManagedIdentityInput{
				WorkspaceID:       "nonexistent",
				ManagedIdentityID: "mi1",
			},
			mockSetup: func(m *managedIdentityMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeWorkspace).Return(nil)
				m.managedIdentities.On("AssignManagedIdentityToWorkspace", mock.Anything, &pb.AssignManagedIdentityToWorkspaceRequest{
					ManagedIdentityId: "mi1",
					WorkspaceId:       "nonexistent",
				}).Return(nil, status.Error(codes.NotFound, "workspace not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &managedIdentityMocks{
				managedIdentities: mocks.NewManagedIdentitiesClient(t),
				acl:               acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					ManagedIdentitiesClient: testMocks.managedIdentities,
				},
				acl: testMocks.acl,
			}

			_, handler := assignManagedIdentity(toolCtx)
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

func TestUnassignManagedIdentity(t *testing.T) {
	type testCase struct {
		name        string
		input       *unassignManagedIdentityInput
		mockSetup   func(*managedIdentityMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "unassign managed identity successfully",
			input: &unassignManagedIdentityInput{
				WorkspaceID:       "ws1",
				ManagedIdentityID: "mi1",
			},
			mockSetup: func(m *managedIdentityMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.managedIdentities.On("RemoveManagedIdentityFromWorkspace", mock.Anything, &pb.RemoveManagedIdentityFromWorkspaceRequest{
					ManagedIdentityId: "mi1",
					WorkspaceId:       "ws1",
				}).Return(nil, nil)
			},
		},
		{
			name: "acl denial",
			input: &unassignManagedIdentityInput{
				WorkspaceID:       "ws1",
				ManagedIdentityID: "mi1",
			},
			mockSetup: func(m *managedIdentityMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "unassign call unsuccessful",
			input: &unassignManagedIdentityInput{
				WorkspaceID:       "ws1",
				ManagedIdentityID: "mi1",
			},
			mockSetup: func(m *managedIdentityMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.managedIdentities.On("RemoveManagedIdentityFromWorkspace", mock.Anything, &pb.RemoveManagedIdentityFromWorkspaceRequest{
					ManagedIdentityId: "mi1",
					WorkspaceId:       "ws1",
				}).Return(nil, status.Error(codes.NotFound, "assignment not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &managedIdentityMocks{
				managedIdentities: mocks.NewManagedIdentitiesClient(t),
				acl:               acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					ManagedIdentitiesClient: testMocks.managedIdentities,
				},
				acl: testMocks.acl,
			}

			_, handler := unassignManagedIdentity(toolCtx)
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
