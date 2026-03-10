package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// plan represents a Terraform plan in MCP responses.
type plan struct {
	ID           string  `json:"id" jsonschema:"Unique identifier for this plan"`
	TRN          string  `json:"trn" jsonschema:"Tharsis Resource Name"`
	Status       string  `json:"status" jsonschema:"Current status: queued, pending, running, finished, errored, or canceled"`
	ErrorMessage *string `json:"error_message,omitempty" jsonschema:"Error details if the plan failed"`
	HasChanges   bool    `json:"has_changes" jsonschema:"True if any resources will be added, changed, or destroyed"`
}

// toPlan converts a proto plan to MCP plan.
func toPlan(p *pb.Plan) *plan {
	return &plan{
		ID:           p.Metadata.Id,
		TRN:          p.Metadata.Trn,
		Status:       p.Status,
		ErrorMessage: p.ErrorMessage,
		HasChanges:   p.HasChanges,
	}
}

// getPlanInput is the input for getting a plan.
type getPlanInput struct {
	ID string `json:"id" jsonschema:"required,Plan ID or TRN (e.g. Ul8yZ... or trn:plan:...)"`
}

// getPlanOutput is the output for getting a plan.
type getPlanOutput struct {
	Plan *plan `json:"plan" jsonschema:"The plan details"`
}

// GetPlan returns an MCP tool for getting a plan.
func getPlan(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getPlanInput, *getPlanOutput]) {
	tool := mcp.Tool{
		Name:        "get_plan",
		Description: "Get details of a Terraform plan by ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Plan",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getPlanInput) (*mcp.CallToolResult, *getPlanOutput, error) {
		resp, err := tc.grpcClient.RunsClient.GetPlanByID(ctx, &pb.GetPlanByIDRequest{Id: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get plan: %w", err)
		}

		return nil, &getPlanOutput{
			Plan: toPlan(resp),
		}, nil
	}

	return tool, handler
}
