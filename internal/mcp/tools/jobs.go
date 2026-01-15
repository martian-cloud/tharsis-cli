package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	// defaultLogLimit is the default number of bytes to retrieve when fetching job logs (10 KiB)
	// Conservative limit to preserve LLM context window space for conversation
	defaultLogLimit = 10 * 1024
	// maxLogLimit is the maximum number of bytes that can be retrieved in a single request (50 KiB)
	// Conservative limit to avoid overwhelming LLM context windows
	maxLogLimit = 50 * 1024
)

// getJobLogsInput is the input for the get_job_logs tool.
type getJobLogsInput struct {
	JobID string `json:"job_id" jsonschema:"required,Job ID from get_run response (plan_job_id or apply_job_id)"`
	Start *int   `json:"start,omitempty" jsonschema:"Starting byte offset. Defaults to 0"`
	Limit *int   `json:"limit,omitempty" jsonschema:"Maximum number of bytes to return. Defaults to 10000 max 50000"`
}

// getJobLogsOutput is the output for the get_job_logs tool.
type getJobLogsOutput struct {
	JobID   string `json:"job_id" jsonschema:"The job ID these logs belong to"`
	Logs    string `json:"logs" jsonschema:"The log content"`
	Start   int    `json:"start" jsonschema:"Starting byte offset of returned logs"`
	Size    int    `json:"size" jsonschema:"Size of returned logs in bytes"`
	HasMore bool   `json:"has_more" jsonschema:"True if there are more logs available"`
}

// GetJobLogs returns an MCP tool for retrieving job logs.
func getJobLogs(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getJobLogsInput, getJobLogsOutput]) {
	tool := mcp.Tool{
		Name:        "get_job_logs",
		Description: "Retrieve logs from a Terraform job. Get the job ID from get_run response (plan_job_id or apply_job_id). Returns a limited number of bytes for efficiency. Use start parameter to paginate through logs. Check has_more in response to determine if additional pages exist.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Job Logs",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getJobLogsInput) (*mcp.CallToolResult, getJobLogsOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getJobLogsOutput{}, fmt.Errorf("failed to get tharsis client: %w", err)
		}

		// Set defaults
		start := int32(0)
		if input.Start != nil {
			start = int32(*input.Start)
		}
		limit := int32(defaultLogLimit)
		if input.Limit != nil {
			if *input.Limit > maxLogLimit {
				return nil, getJobLogsOutput{}, fmt.Errorf("limit %d exceeds maximum allowed limit of %d bytes", *input.Limit, maxLogLimit)
			}
			limit = int32(*input.Limit)
		}

		// Request one extra byte to detect if there's more data
		requestLimit := limit + 1
		logsResp, err := client.Jobs().GetJobLogs(ctx, &sdktypes.GetJobLogsInput{
			JobID: input.JobID,
			Start: start,
			Limit: &requestLimit,
		})
		if err != nil {
			return nil, getJobLogsOutput{}, fmt.Errorf("failed to get logs for job %q: %w", input.JobID, err)
		}

		// If we got more than limit, there's more data available
		logs := logsResp.Logs
		hasMore := len(logs) > int(limit)
		if hasMore {
			logs = logs[:limit]
		}

		return nil, getJobLogsOutput{
			JobID:   input.JobID,
			Logs:    logs,
			Start:   int(start),
			Size:    len(logs),
			HasMore: hasMore,
		}, nil
	}

	return tool, handler
}
