package types

import "time"

// WorkloadRevision is a retained spec snapshot for instant rollback.
type WorkloadRevision struct {
	ID           string        `json:"id"`
	RecordedAt   time.Time     `json:"recordedAt"`
	Spec         WorkloadSpec  `json:"spec"`
	Tags         *WorkloadTags `json:"tags,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	ChangeReason string        `json:"changeReason,omitempty"`
}

// RollbackHistoryStatus exposes revision bookkeeping on the workload.
type RollbackHistoryStatus struct {
	MaxVersions int    `json:"maxVersions,omitempty"`
	Count       int    `json:"count,omitempty"`
	ActiveID    string `json:"activeRevisionId,omitempty"`
}

// RollbackRequest reverts a workload to a stored revision.
type RollbackRequest struct {
	RevisionID string `json:"revisionId,omitempty"` // omit to revert to previous version
}

// RollbackResult is returned after a successful rollback.
type RollbackResult struct {
	WorkloadID     string   `json:"workloadId"`
	Namespace      string   `json:"namespace,omitempty"`
	RevisionID     string   `json:"revisionId"`
	PreviousSpec   string   `json:"previousImage,omitempty"`
	RestoredSpec   string   `json:"restoredImage,omitempty"`
	Status         string   `json:"status"`
	Message        string   `json:"message,omitempty"`
}

// RevisionListResponse lists available rollback revisions.
type RevisionListResponse struct {
	WorkloadID string             `json:"workloadId"`
	Namespace  string             `json:"namespace,omitempty"`
	Revisions  []WorkloadRevision `json:"revisions"`
	Count      int                `json:"count"`
}
