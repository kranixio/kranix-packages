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
