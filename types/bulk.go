package types

// BulkOperation is deploy, restart, or delete.
type BulkOperation string

const (
	BulkOpDeploy  BulkOperation = "deploy"
	BulkOpRestart BulkOperation = "restart"
	BulkOpDelete  BulkOperation = "delete"
)

// BulkWorkloadItem is one target in a bulk request.
type BulkWorkloadItem struct {
	ID   string       `json:"id,omitempty"`
	Spec WorkloadSpec `json:"spec,omitempty"`
}

// BulkWorkloadRequest batches workload mutations.
type BulkWorkloadRequest struct {
	Operation       BulkOperation      `json:"operation"`
	Workloads       []BulkWorkloadItem `json:"workloads"`
	ContinueOnError bool               `json:"continueOnError,omitempty"`
}

// BulkWorkloadResult is the per-item outcome.
type BulkWorkloadResult struct {
	ID      string `json:"id,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BulkWorkloadResponse aggregates bulk results.
type BulkWorkloadResponse struct {
	Operation string               `json:"operation"`
	Results   []BulkWorkloadResult `json:"results"`
	Succeeded int                  `json:"succeeded"`
	Failed    int                  `json:"failed"`
}
