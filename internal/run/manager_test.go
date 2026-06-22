package run

import (
	"context"
	"io"
	"os"
	"path/filepath"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestCreateRun(t *testing.T) {
	type testCase struct {
		name        string
		input       *CreateRunInput
		expectError bool
	}

	testCases := []testCase{
		{
			name: "both directory and module source",
			input: &CreateRunInput{
				WorkspaceID:   "ws1",
				DirectoryPath: ptr.String("./"),
				ModuleSource:  ptr.String("module"),
			},
			expectError: true,
		},
		{
			name: "module version without module source",
			input: &CreateRunInput{
				WorkspaceID:   "ws1",
				ModuleVersion: ptr.String("1.0.0"),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			grpcClient := &client.GRPCClient{
				WorkspacesClient:            mocks.NewWorkspacesClient(t),
				ConfigurationVersionsClient: mocks.NewConfigurationVersionsClient(t),
				RunsClient:                  mocks.NewRunsClient(t),
				JobsClient:                  mocks.NewJobsClient(t),
			}

			mgr := &Manager{
				grpcClient: grpcClient,
				tfeClient:  client.NewMockRESTClient(t),
				logger:     hclog.NewNullLogger(),
				ui:         terminal.NewNoopUI(),
			}

			_, err := mgr.CreateRun(context.Background(), tc.input)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplyRun(t *testing.T) {
	// Note: Success cases with log streaming are not tested here due to complexity of mocking gRPC streams.
	// Those scenarios are covered by integration tests.
	type testCase struct {
		name       string
		runID      string
		setupMocks func(*client.GRPCClient)
	}

	testCases := []testCase{
		{
			name:  "apply run fails",
			runID: "run1",
			setupMocks: func(c *client.GRPCClient) {
				mockRuns := c.RunsClient.(*mocks.RunsClient)
				mockRuns.On("ApplyRun", mock.Anything, &pb.ApplyRunRequest{RunId: "run1"}).
					Return(nil, assert.AnError)
			},
		},
		{
			name:  "get job fails",
			runID: "run1",
			setupMocks: func(c *client.GRPCClient) {
				mockRuns := c.RunsClient.(*mocks.RunsClient)
				mockJobs := c.JobsClient.(*mocks.JobsClient)
				mockRuns.On("ApplyRun", mock.Anything, &pb.ApplyRunRequest{RunId: "run1"}).
					Return(&pb.Run{ApplyId: "apply1", Status: "apply_queued"}, nil)
				mockRuns.On("GetApplyByID", mock.Anything, &pb.GetApplyByIDRequest{Id: "apply1"}).
					Return(&pb.Apply{Status: "queued"}, nil)
				mockJobs.On("GetLatestJobForApply", mock.Anything, &pb.GetLatestJobForApplyRequest{ApplyId: "apply1"}).
					Return(nil, assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			grpcClient := &client.GRPCClient{
				RunsClient: mocks.NewRunsClient(t),
				JobsClient: mocks.NewJobsClient(t),
			}
			tc.setupMocks(grpcClient)

			mgr := &Manager{
				grpcClient: grpcClient,
				logger:     hclog.NewNullLogger(),
				ui:         terminal.NewNoopUI(),
			}

			_, err := mgr.ApplyRun(context.Background(), tc.runID)
			require.Error(t, err)
		})
	}
}

func TestPlanJobReady(t *testing.T) {
	type testCase struct {
		status      string
		expectReady bool
		expectError bool
	}

	testCases := []testCase{
		{status: "created", expectReady: false},
		{status: "pending", expectReady: false},
		{status: "queued", expectReady: true},
		{status: "running", expectReady: true},
		{status: "finished", expectReady: true},
		{status: "errored", expectReady: true},
		{status: "canceled", expectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			ready, err := planJobReady(tc.status)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectReady, ready)
		})
	}
}

func TestApplyJobReady(t *testing.T) {
	type testCase struct {
		status      string
		expectReady bool
		expectError bool
	}

	testCases := []testCase{
		{status: "created", expectReady: false},
		{status: "pending", expectReady: false},
		{status: "queued", expectReady: true},
		{status: "running", expectReady: true},
		{status: "finished", expectReady: true},
		{status: "errored", expectReady: true},
		{status: "canceled", expectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			ready, err := applyJobReady(tc.status)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectReady, ready)
		})
	}
}

// fakeRunEventStream is a minimal grpc.ServerStreamingClient[pb.RunEvent] for tests.
type fakeRunEventStream struct {
	events []*pb.RunEvent
	idx    int
}

func (s *fakeRunEventStream) Recv() (*pb.RunEvent, error) {
	if s.idx >= len(s.events) {
		return nil, io.EOF
	}
	event := s.events[s.idx]
	s.idx++
	return event, nil
}

func (*fakeRunEventStream) Header() (metadata.MD, error) { return nil, nil }
func (*fakeRunEventStream) Trailer() metadata.MD         { return nil }
func (*fakeRunEventStream) CloseSend() error             { return nil }
func (*fakeRunEventStream) Context() context.Context     { return context.Background() }
func (*fakeRunEventStream) SendMsg(any) error            { return nil }
func (*fakeRunEventStream) RecvMsg(any) error            { return nil }

func TestWaitForRunJob(t *testing.T) {
	// statusReturner returns the supplied statuses in order across calls, repeating the last.
	statusReturner := func(statuses ...string) func(context.Context) (string, error) {
		i := 0
		return func(context.Context) (string, error) {
			s := statuses[i]
			if i < len(statuses)-1 {
				i++
			}
			return s, nil
		}
	}

	newManager := func(t *testing.T) (*Manager, *mocks.RunsClient) {
		runs := mocks.NewRunsClient(t)
		return &Manager{
			grpcClient: &client.GRPCClient{RunsClient: runs},
			logger:     hclog.NewNullLogger(),
			ui:         terminal.NewNoopUI(),
		}, runs
	}

	t.Run("returns when initial status already shows a job", func(t *testing.T) {
		mgr, _ := newManager(t) // SubscribeToRunEvents must not be called
		err := mgr.waitForRunJob(context.Background(), "ws-1", "run-1",
			statusReturner("queued"), planJobReady)
		require.NoError(t, err)
	})

	t.Run("returns error when initial status is a final state without a job", func(t *testing.T) {
		mgr, _ := newManager(t)
		err := mgr.waitForRunJob(context.Background(), "ws-1", "run-1",
			statusReturner("canceled"), planJobReady)
		require.Error(t, err)
	})

	t.Run("propagates getStatus error", func(t *testing.T) {
		mgr, _ := newManager(t)
		err := mgr.waitForRunJob(context.Background(), "ws-1", "run-1",
			func(context.Context) (string, error) {
				return "", status.Error(codes.NotFound, "not found")
			}, planJobReady)
		require.Error(t, err)
	})

	t.Run("returns error when subscribe fails", func(t *testing.T) {
		mgr, runs := newManager(t)
		runs.On("SubscribeToRunEvents", mock.Anything, mock.Anything).
			Return(nil, status.Error(codes.Unimplemented, "not supported"))
		err := mgr.waitForRunJob(context.Background(), "ws-1", "run-1",
			statusReturner("pending", "queued"), planJobReady)
		require.Error(t, err)
	})

	t.Run("waits for a run event then proceeds", func(t *testing.T) {
		mgr, runs := newManager(t)
		runs.On("SubscribeToRunEvents", mock.Anything, mock.Anything).
			Return(&fakeRunEventStream{events: []*pb.RunEvent{{Run: &pb.Run{Status: "plan_queued"}}}}, nil)
		err := mgr.waitForRunJob(context.Background(), "ws-1", "run-1",
			statusReturner("pending", "queued"), planJobReady)
		require.NoError(t, err)
	})
}

func TestProcessDirectoryPath(t *testing.T) {
	type testCase struct {
		name        string
		setupDir    func(string)
		testPath    func(string) string
		isDestroy   bool
		expectError bool
	}

	testCases := []testCase{
		{
			name: "valid directory with tf file",
			setupDir: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte("resource {}"), 0600))
			},
			testPath: func(dir string) string { return dir },
		},
		{
			name: "valid directory with tf.json file",
			setupDir: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf.json"), []byte("{}"), 0600))
			},
			testPath: func(dir string) string { return dir },
		},
		{
			name:      "destroy mode allows empty directory",
			setupDir:  func(string) {},
			testPath:  func(dir string) string { return dir },
			isDestroy: true,
		},
		{
			name: "no config files in non-destroy mode",
			setupDir: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("docs"), 0600))
			},
			testPath:    func(dir string) string { return dir },
			expectError: true,
		},
		{
			name: "config file in subdirectory",
			setupDir: func(dir string) {
				subdir := filepath.Join(dir, "modules")
				require.NoError(t, os.Mkdir(subdir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(subdir, "main.tf"), []byte("resource {}"), 0600))
			},
			testPath: func(dir string) string { return dir },
		},
		{
			name:        "non-existent directory",
			setupDir:    func(string) {},
			testPath:    func(dir string) string { return filepath.Join(dir, "nonexistent") },
			expectError: true,
		},
		{
			name: "path is not a directory",
			setupDir: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0600))
			},
			testPath:    func(dir string) string { return filepath.Join(dir, "file.txt") },
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tc.setupDir(tmpDir)

			err := processDirectoryPath(tc.testPath(tmpDir), tc.isDestroy)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReAnsi(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "plain text",
			input:  "no colors here",
			expect: "no colors here",
		},
		{
			name:   "bold",
			input:  "\x1b[1mBold\x1b[0m",
			expect: "Bold",
		},
		{
			name:   "terraform plan output",
			input:  "\x1b[0m\x1b[1m\x1b[32mApply complete! Resources: 1 added, 0 changed, 0 destroyed.\x1b[0m",
			expect: "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.",
		},
		{
			name:   "256 color",
			input:  "\x1b[38;5;196mred text\x1b[0m",
			expect: "red text",
		},
		{
			name:   "24-bit true color",
			input:  "\x1b[38;2;255;0;0mred\x1b[0m",
			expect: "red",
		},
		{
			name:   "cursor movement",
			input:  "\x1b[2K\x1b[1A\x1b[2Krefreshing...",
			expect: "refreshing...",
		},
		{
			name:   "mixed content",
			input:  "Plan: \x1b[32m1 to add\x1b[0m, \x1b[33m0 to change\x1b[0m, \x1b[31m0 to destroy\x1b[0m.",
			expect: "Plan: 1 to add, 0 to change, 0 to destroy.",
		},
		{
			name:   "empty string",
			input:  "",
			expect: "",
		},
		{
			name:   "hyperlink",
			input:  "\x1b]8;;https://example.com\x07click here\x1b]8;;\x07",
			expect: "click here",
		},
		{
			name:   "terraform error with bold red",
			input:  "\x1b[0m\x1b[1m\x1b[31mError: \x1b[0m\x1b[0m\x1b[1mInsufficient features blocks\x1b[0m",
			expect: "Error: Insufficient features blocks",
		},
		{
			name:   "only escape codes",
			input:  "\x1b[0m\x1b[1m\x1b[0m",
			expect: "",
		},
		{
			name:   "tab and newline preserved",
			input:  "\x1b[32m+\x1b[0m resource\n\t\x1b[33m~\x1b[0m attribute",
			expect: "+ resource\n\t~ attribute",
		},
		{
			name:   "CSI with question mark (show/hide cursor)",
			input:  "\x1b[?25lhidden cursor\x1b[?25h",
			expect: "hidden cursor",
		},
		{
			name:   "multiple resets in sequence",
			input:  "\x1b[0m\x1b[0m\x1b[0mtext\x1b[0m\x1b[0m",
			expect: "text",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, terminal.StripAnsi(test.input))
		})
	}
}
