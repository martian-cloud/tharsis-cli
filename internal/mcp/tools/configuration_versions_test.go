package tools

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGetConfigurationVersion(t *testing.T) {
	cvID := "test-cv-id"

	tests := []struct {
		name        string
		cv          *sdktypes.ConfigurationVersion
		aclError    error
		expectError bool
		validate    func(*testing.T, getConfigurationVersionOutput)
	}{
		{
			name: "successful configuration version retrieval",
			cv: &sdktypes.ConfigurationVersion{
				Metadata:    sdktypes.ResourceMetadata{ID: cvID},
				Status:      "uploaded",
				Speculative: false,
			},
			validate: func(t *testing.T, output getConfigurationVersionOutput) {
				assert.Equal(t, cvID, output.ConfigurationVersion.ID)
				assert.Equal(t, "uploaded", output.ConfigurationVersion.Status)
				assert.False(t, output.ConfigurationVersion.Speculative)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockCV := tharsis.NewConfigurationVersion(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("ConfigurationVersions").Return(mockCV)
				mockCV.On("GetConfigurationVersion", mock.Anything, &sdktypes.GetConfigurationVersionInput{ID: cvID}).Return(tt.cv, nil)
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := getConfigurationVersion(tc)
			_, output, err := handler(t.Context(), nil, getConfigurationVersionInput{
				ID: cvID,
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

func TestCreateConfigurationVersion(t *testing.T) {
	workspacePath := "group/workspace"

	tests := []struct {
		name        string
		input       createConfigurationVersionInput
		cv          *sdktypes.ConfigurationVersion
		aclError    error
		expectError bool
		validate    func(*testing.T, createConfigurationVersionOutput)
	}{
		{
			name: "successful configuration version creation",
			input: createConfigurationVersionInput{
				WorkspacePath: workspacePath,
				DirectoryPath: "/tmp/test",
			},
			cv: &sdktypes.ConfigurationVersion{
				Metadata: sdktypes.ResourceMetadata{ID: "cv-id"},
				Status:   "pending",
			},
			validate: func(t *testing.T, output createConfigurationVersionOutput) {
				assert.Equal(t, "cv-id", output.ConfigurationVersion.ID)
				assert.Equal(t, "pending", output.ConfigurationVersion.Status)
			},
		},
		{
			name: "ACL denial",
			input: createConfigurationVersionInput{
				WorkspacePath: workspacePath,
				DirectoryPath: "/tmp/test",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockCV := tharsis.NewConfigurationVersion(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("ConfigurationVersions").Return(mockCV)
				mockCV.On("CreateConfigurationVersion", mock.Anything, &sdktypes.CreateConfigurationVersionInput{
					WorkspacePath: tt.input.WorkspacePath,
					Speculative:   tt.input.Speculative,
				}).Return(tt.cv, nil)
				mockCV.On("UploadConfigurationVersion", mock.Anything, &sdktypes.UploadConfigurationVersionInput{
					WorkspacePath:          tt.input.WorkspacePath,
					ConfigurationVersionID: tt.cv.Metadata.ID,
					DirectoryPath:          tt.input.DirectoryPath,
				}).Return(nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, "trn:workspace:"+workspacePath, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := createConfigurationVersion(tc)
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

func TestDownloadConfigurationVersion(t *testing.T) {
	cvID := "test-cv-id"

	mockClient := tharsis.NewMockClient(t)
	mockCV := tharsis.NewConfigurationVersion(t)

	mockClient.On("ConfigurationVersions").Return(mockCV)
	mockCV.On("DownloadConfigurationVersion", mock.Anything, &sdktypes.GetConfigurationVersionInput{ID: cvID}, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			writer := args.Get(2).(io.Writer)

			// Create a tar.gz with test content
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			tarWriter := tar.NewWriter(gzWriter)

			// Add a test file
			content := []byte("resource \"test\" \"example\" {}")
			header := &tar.Header{
				Name: "main.tf",
				Mode: 0600,
				Size: int64(len(content)),
			}
			tarWriter.WriteHeader(header)
			tarWriter.Write(content)

			tarWriter.Close()
			gzWriter.Close()

			writer.Write(buf.Bytes())
		})

	tc := &ToolContext{
		tharsisURL:  "https://test.tharsis.io",
		profileName: "test",
		clientGetter: func() (tharsis.Client, error) {
			return mockClient, nil
		},
	}

	_, handler := downloadConfigurationVersion(tc)
	_, output, err := handler(t.Context(), nil, downloadConfigurationVersionInput{
		ID: cvID,
	})

	assert.NoError(t, err)
	assert.Equal(t, cvID, output.ConfigurationVersionID)
	assert.NotEmpty(t, output.OutputPath)
	assert.DirExists(t, output.OutputPath)

	// Verify the test file was extracted
	testFilePath := filepath.Join(output.OutputPath, "main.tf")
	assert.FileExists(t, testFilePath)

	content, err := os.ReadFile(testFilePath)
	assert.NoError(t, err)
	assert.Equal(t, "resource \"test\" \"example\" {}", string(content))
}
