package run

import (
	"context"
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
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
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
			grpcClient := &client.Client{
				WorkspacesClient:            mocks.NewWorkspacesClient(t),
				ConfigurationVersionsClient: mocks.NewConfigurationVersionsClient(t),
				RunsClient:                  mocks.NewRunsClient(t),
				JobsClient:                  mocks.NewJobsClient(t),
			}

			mgr := &Manager{
				grpcClient: grpcClient,
				tfeClient:  tfe.NewMockRESTClient(t),
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
		setupMocks func(*client.Client)
	}

	testCases := []testCase{
		{
			name:  "apply run fails",
			runID: "run1",
			setupMocks: func(c *client.Client) {
				mockRuns := c.RunsClient.(*mocks.RunsClient)
				mockRuns.On("ApplyRun", mock.Anything, &pb.ApplyRunRequest{RunId: "run1"}).
					Return(nil, assert.AnError)
			},
		},
		{
			name:  "get job fails",
			runID: "run1",
			setupMocks: func(c *client.Client) {
				mockRuns := c.RunsClient.(*mocks.RunsClient)
				mockJobs := c.JobsClient.(*mocks.JobsClient)
				mockRuns.On("ApplyRun", mock.Anything, &pb.ApplyRunRequest{RunId: "run1"}).
					Return(&pb.Run{ApplyId: "apply1"}, nil)
				mockJobs.On("GetLatestJobForApply", mock.Anything, &pb.GetLatestJobForApplyRequest{ApplyId: "apply1"}).
					Return(nil, assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			grpcClient := &client.Client{
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
