package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

const (
	// defaultLogLimit is the default number of bytes to retrieve when fetching job logs (10 KiB)
	// Conservative limit to preserve LLM context window space for conversation
	defaultLogLimit = 10 * 1024
	// maxLogLimit is the maximum number of bytes that can be retrieved in a single request (50 KiB)
	// Conservative limit to avoid overwhelming LLM context windows
	maxLogLimit = 50 * 1024
)

// job represents a Tharsis job in MCP responses.
type job struct {
	ID             string `json:"id" jsonschema:"The unique identifier of the job"`
	TRN            string `json:"trn" jsonschema:"Tharsis Resource Name"`
	WorkspaceID    string `json:"workspace_id" jsonschema:"The workspace ID this job belongs to"`
	RunID          string `json:"run_id" jsonschema:"The run ID this job belongs to"`
	Type           string `json:"type" jsonschema:"Job type (plan or apply)"`
	Status         string `json:"status" jsonschema:"Job status (queued pending running finished canceled)"`
	MaxJobDuration int32  `json:"max_job_duration" jsonschema:"Maximum job duration in minutes"`
}

// toJob converts a proto job to MCP job.
func toJob(j *pb.Job) *job {
	return &job{
		ID:             j.Metadata.Id,
		TRN:            j.Metadata.Trn,
		WorkspaceID:    j.WorkspaceId,
		RunID:          j.RunId,
		Type:           j.Type,
		Status:         j.Status,
		MaxJobDuration: j.MaxJobDuration,
	}
}

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
func getJobLogs(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getJobLogsInput, *getJobLogsOutput]) {
	tool := mcp.Tool{
		Name:        "get_job_logs",
		Description: "Retrieve logs from a Terraform job. Get the job ID from get_run response (plan_job_id or apply_job_id). Returns a limited number of bytes for efficiency. Use start parameter to paginate through logs. Check has_more in response to determine if additional pages exist.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Job Logs",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getJobLogsInput) (*mcp.CallToolResult, *getJobLogsOutput, error) {
		start := int32(ptr.ToInt(input.Start))
		limit := int32(defaultLogLimit)
		if input.Limit != nil {
			if *input.Limit > maxLogLimit {
				return nil, nil, fmt.Errorf("limit %d exceeds maximum allowed limit of %d bytes", *input.Limit, maxLogLimit)
			}
			limit = int32(*input.Limit)
		}

		// Request one extra byte to detect if there's more data
		requestLimit := limit + 1
		logsResp, err := tc.grpcClient.JobsClient.GetJobLogs(ctx, &pb.GetJobLogsRequest{
			JobId:       input.JobID,
			StartOffset: start,
			Limit:       requestLimit,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get logs for job %q: %w", input.JobID, err)
		}

		// If we got more than limit, there's more data available
		logs := logsResp.Logs
		hasMore := len(logs) > int(limit)
		if hasMore {
			logs = logs[:limit]
		}

		return nil, &getJobLogsOutput{
			JobID:   input.JobID,
			Logs:    logs,
			Start:   int(start),
			Size:    len(logs),
			HasMore: hasMore,
		}, nil
	}

	return tool, handler
}

// getLatestJobInput is the input for the get_latest_job tool.
type getLatestJobInput struct {
	PlanID  *string `json:"plan_id,omitempty" jsonschema:"Plan ID to get the latest job for (e.g. Ul8yZ... or trn:plan:my-group/my-workspace/plan-id)"`
	ApplyID *string `json:"apply_id,omitempty" jsonschema:"Apply ID to get the latest job for (e.g. Ul8yZ... or trn:apply:my-group/my-workspace/apply-id)"`
}

// getLatestJobOutput is the output for the get_latest_job tool.
type getLatestJobOutput struct {
	Job *job `json:"job,omitempty" jsonschema:"The latest job details"`
}

// GetLatestJob returns an MCP tool for retrieving the latest job for a plan or apply.
func getLatestJob(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getLatestJobInput, *getLatestJobOutput]) {
	tool := mcp.Tool{
		Name:        "get_latest_job",
		Description: "Get the latest job for a plan or apply. Provide either plan_id or apply_id. Use this to get the job ID for retrieving logs.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Latest Job",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getLatestJobInput) (*mcp.CallToolResult, *getLatestJobOutput, error) {
		if input.PlanID == nil && input.ApplyID == nil {
			return nil, nil, fmt.Errorf("either plan_id or apply_id must be provided")
		}
		if input.PlanID != nil && input.ApplyID != nil {
			return nil, nil, fmt.Errorf("only one of plan_id or apply_id can be provided")
		}

		var jobResp *pb.Job
		var err error

		if input.PlanID != nil {
			jobResp, err = tc.grpcClient.JobsClient.GetLatestJobForPlan(ctx, &pb.GetLatestJobForPlanRequest{
				PlanId: *input.PlanID,
			})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get latest job for plan %q: %w", *input.PlanID, err)
			}
		} else {
			jobResp, err = tc.grpcClient.JobsClient.GetLatestJobForApply(ctx, &pb.GetLatestJobForApplyRequest{
				ApplyId: *input.ApplyID,
			})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get latest job for apply %q: %w", *input.ApplyID, err)
			}
		}

		return nil, &getLatestJobOutput{
			Job: toJob(jobResp),
		}, nil
	}

	return tool, handler
}
