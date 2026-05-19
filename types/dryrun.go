package types

// DryRunResponse is returned when ?dryRun=true on a mutating endpoint.
type DryRunResponse struct {
	DryRun     bool                   `json:"dryRun"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource,omitempty"`
	ResourceID string                 `json:"resourceId,omitempty"`
	WouldApply map[string]interface{} `json:"wouldApply,omitempty"`
	Message    string                 `json:"message"`
}
