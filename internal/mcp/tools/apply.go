package tools

import sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"

// apply represents a Terraform apply in MCP responses.
type apply struct {
	ID           string               `json:"id" jsonschema:"Unique identifier for this apply"`
	TRN          string               `json:"trn" jsonschema:"Tharsis Resource Name"`
	Status       sdktypes.ApplyStatus `json:"status" jsonschema:"Current status: created, queued, pending, running, finished, errored, or canceled"`
	CurrentJobID *string              `json:"current_job_id,omitempty" jsonschema:"ID of the current job. Use with get_job_logs to retrieve apply logs"`
	ErrorMessage *string              `json:"error_message,omitempty" jsonschema:"Error details if the apply failed"`
}

// toApply converts an SDK apply to MCP apply.
func toApply(a *sdktypes.Apply) *apply {
	if a == nil {
		return nil
	}
	return &apply{
		ID:           a.Metadata.ID,
		TRN:          a.Metadata.TRN,
		Status:       a.Status,
		CurrentJobID: a.CurrentJobID,
		ErrorMessage: a.ErrorMessage,
	}
}
