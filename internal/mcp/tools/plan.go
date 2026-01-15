package tools

import sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"

// plan represents a Terraform plan in MCP responses.
type plan struct {
	ID                   string              `json:"id" jsonschema:"Unique identifier for this plan"`
	TRN                  string              `json:"trn" jsonschema:"Tharsis Resource Name"`
	Status               sdktypes.PlanStatus `json:"status" jsonschema:"Current status: queued, pending, running, finished, errored, or canceled"`
	CurrentJobID         *string             `json:"current_job_id,omitempty" jsonschema:"ID of the current job. Use with get_job_logs to retrieve plan logs"`
	ErrorMessage         *string             `json:"error_message,omitempty" jsonschema:"Error details if the plan failed"`
	HasChanges           bool                `json:"has_changes" jsonschema:"True if any resources will be added, changed, or destroyed"`
	ResourceAdditions    int                 `json:"resource_additions" jsonschema:"Number of new resources that will be created"`
	ResourceChanges      int                 `json:"resource_changes" jsonschema:"Number of existing resources that will be modified"`
	ResourceDestructions int                 `json:"resource_destructions" jsonschema:"Number of resources that will be deleted"`
}

// toPlan converts an SDK plan to MCP plan.
func toPlan(p *sdktypes.Plan) *plan {
	if p == nil {
		return nil
	}
	return &plan{
		ID:                   p.Metadata.ID,
		TRN:                  p.Metadata.TRN,
		Status:               p.Status,
		CurrentJobID:         p.CurrentJobID,
		ErrorMessage:         p.ErrorMessage,
		HasChanges:           p.HasChanges,
		ResourceAdditions:    p.ResourceAdditions,
		ResourceChanges:      p.ResourceChanges,
		ResourceDestructions: p.ResourceDestructions,
	}
}
