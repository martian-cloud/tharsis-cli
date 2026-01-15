package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGetJobLogs(t *testing.T) {
	jobID := "test-job-id"

	tests := []struct {
		name        string
		input       getJobLogsInput
		logs        *sdktypes.JobLogs
		logsErr     error
		expectError bool
		validate    func(*testing.T, getJobLogsOutput)
	}{
		{
			name: "successful logs retrieval",
			input: getJobLogsInput{
				JobID: jobID,
			},
			logs: &sdktypes.JobLogs{
				Logs: "plan log output",
			},
			validate: func(t *testing.T, output getJobLogsOutput) {
				assert.Equal(t, jobID, output.JobID)
				assert.Equal(t, "plan log output", output.Logs)
				assert.Equal(t, 0, output.Start)
				assert.Equal(t, 15, output.Size)
				assert.False(t, output.HasMore)
			},
		},
		{
			name: "logs with pagination",
			input: getJobLogsInput{
				JobID: jobID,
				Start: ptr.Int(100),
				Limit: ptr.Int(10),
			},
			logs: &sdktypes.JobLogs{
				Logs: "12345678901",
			},
			validate: func(t *testing.T, output getJobLogsOutput) {
				assert.Equal(t, 100, output.Start)
				assert.Equal(t, 10, output.Size)
				assert.True(t, output.HasMore)
			},
		},
		{
			name: "job not found",
			input: getJobLogsInput{
				JobID: jobID,
			},
			logsErr:     assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockJob := tharsis.NewJob(t)

			mockClient.On("Jobs").Return(mockJob)

			start := int32(0)
			if tt.input.Start != nil {
				start = int32(*tt.input.Start)
			}
			limit := int32(defaultLogLimit)
			if tt.input.Limit != nil {
				limit = int32(*tt.input.Limit)
			}

			mockJob.On("GetJobLogs", mock.Anything, &sdktypes.GetJobLogsInput{
				JobID: jobID,
				Start: start,
				Limit: ptr.Int32(limit + 1),
			}).Return(tt.logs, tt.logsErr)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
			}

			_, handler := getJobLogs(tc)
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
