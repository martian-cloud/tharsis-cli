package run

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

func TestParseAnnotations(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantLen int
		wantErr bool
	}{
		{
			name:    "empty input returns nil",
			input:   nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "single valid annotation",
			input:   []string{`{"key":"commit","value":"abc123","link":"https://example.com/commit/abc123"}`},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "multiple valid annotations",
			input: []string{
				`{"key":"commit","value":"abc123"}`,
				`{"key":"ref","value":"main"}`,
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "invalid JSON returns error",
			input:   []string{"not-json"},
			wantErr: true,
		},
		{
			name:    "unknown field returns error",
			input:   []string{`{"nonsense":1}`},
			wantErr: true,
		},
		{
			name:    "wrong-case keys return error",
			input:   []string{`{"Key":"K","Value":"V"}`},
			wantErr: true,
		},
		{
			name:    "null returns error",
			input:   []string{"null"},
			wantErr: true,
		},
		{
			name:    "empty object returns error",
			input:   []string{`{}`},
			wantErr: true,
		},
		{
			name:    "missing value returns error",
			input:   []string{`{"key":"commit"}`},
			wantErr: true,
		},
		{
			name:    "missing key returns error",
			input:   []string{`{"value":"abc123"}`},
			wantErr: true,
		},
		{
			name:    "blank key returns error",
			input:   []string{`{"key":"","value":"abc123"}`},
			wantErr: true,
		},
		{
			name:    "blank value returns error",
			input:   []string{`{"key":"commit","value":""}`},
			wantErr: true,
		},
		{
			name: "one invalid annotation rejects the whole set",
			input: []string{
				`{"key":"commit","value":"abc123"}`,
				`{"key":"ref"}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAnnotations(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

// TestParseAnnotations_PreservesFields confirms key, value, and the optional link survive parsing.
func TestParseAnnotations_PreservesFields(t *testing.T) {
	link := "https://example.com/commit/abc123"
	got, err := parseAnnotations([]string{
		`{"key":"commit","value":"abc123","link":"https://example.com/commit/abc123"}`,
		`{"key":"ref","value":"main"}`,
	})
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "commit", got[0].Key)
	assert.Equal(t, "abc123", got[0].Value)
	require.NotNil(t, got[0].Link)
	assert.Equal(t, link, *got[0].Link)

	assert.Equal(t, "ref", got[1].Key)
	assert.Equal(t, "main", got[1].Value)
	assert.Nil(t, got[1].Link)
}

// TestCreateRun_AnnotationsReachRequest verifies the wiring from CreateRunInput.Annotations through
// to the CreateRunRequest sent to the API — parsing alone passing does not prove the field is
// actually forwarded. CreateRun is stubbed to fail so the flow stops right after the request is
// built, keeping the test focused on the request contents.
func TestCreateRun_AnnotationsReachRequest(t *testing.T) {
	mockWorkspaces := mocks.NewWorkspacesClient(t)
	mockRuns := mocks.NewRunsClient(t)

	mockWorkspaces.On("GetWorkspaceByID", mock.Anything, mock.Anything).
		Return(&pb.Workspace{Metadata: &pb.ResourceMetadata{Id: "ws-1"}}, nil)

	var captured *pb.CreateRunRequest
	mockRuns.On("CreateRun", mock.Anything, mock.MatchedBy(func(req *pb.CreateRunRequest) bool {
		captured = req
		return true
	})).Return(nil, errors.New("stop after request is built"))

	mgr := &Manager{
		grpcClient: &client.GRPCClient{
			WorkspacesClient: mockWorkspaces,
			RunsClient:       mockRuns,
		},
		logger: hclog.NewNullLogger(),
		ui:     terminal.NewNoopUI(),
	}

	_, err := mgr.CreateRun(context.Background(), &CreateRunInput{
		WorkspaceID:  "ws-1",
		ModuleSource: ptr.String("registry.terraform.io/namespace/module/aws"),
		Annotations: []string{
			`{"key":"commit","value":"abc123","link":"https://example.com/commit/abc123"}`,
			`{"key":"ref","value":"main"}`,
		},
	})
	require.Error(t, err)

	require.NotNil(t, captured, "CreateRun was never called")
	require.Len(t, captured.Annotations, 2)

	assert.Equal(t, "commit", captured.Annotations[0].Key)
	assert.Equal(t, "abc123", captured.Annotations[0].Value)
	require.NotNil(t, captured.Annotations[0].Link)
	assert.Equal(t, "https://example.com/commit/abc123", *captured.Annotations[0].Link)

	assert.Equal(t, "ref", captured.Annotations[1].Key)
	assert.Equal(t, "main", captured.Annotations[1].Value)
	assert.Nil(t, captured.Annotations[1].Link)
}

// TestCreateRun_NoAnnotationsLeavesFieldNil confirms that omitting --annotation sends no annotations
// rather than an empty slice.
func TestCreateRun_NoAnnotationsLeavesFieldNil(t *testing.T) {
	mockWorkspaces := mocks.NewWorkspacesClient(t)
	mockRuns := mocks.NewRunsClient(t)

	mockWorkspaces.On("GetWorkspaceByID", mock.Anything, mock.Anything).
		Return(&pb.Workspace{Metadata: &pb.ResourceMetadata{Id: "ws-1"}}, nil)

	var captured *pb.CreateRunRequest
	mockRuns.On("CreateRun", mock.Anything, mock.MatchedBy(func(req *pb.CreateRunRequest) bool {
		captured = req
		return true
	})).Return(nil, errors.New("stop after request is built"))

	mgr := &Manager{
		grpcClient: &client.GRPCClient{
			WorkspacesClient: mockWorkspaces,
			RunsClient:       mockRuns,
		},
		logger: hclog.NewNullLogger(),
		ui:     terminal.NewNoopUI(),
	}

	_, err := mgr.CreateRun(context.Background(), &CreateRunInput{
		WorkspaceID:  "ws-1",
		ModuleSource: ptr.String("registry.terraform.io/namespace/module/aws"),
	})
	require.Error(t, err)

	require.NotNil(t, captured, "CreateRun was never called")
	assert.Nil(t, captured.Annotations)
}
