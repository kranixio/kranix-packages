package types

import (
	"context"
	"time"
)

// WorkloadMigrationRequest moves a running workload between runtime backends.
type WorkloadMigrationRequest struct {
	WorkloadID    string        `json:"workloadId"`
	Namespace     string        `json:"namespace,omitempty"`
	SourceBackend string        `json:"sourceBackend,omitempty"`
	TargetBackend string        `json:"targetBackend"`
	Spec          *WorkloadSpec `json:"spec,omitempty"`
	ZeroDowntime  bool          `json:"zeroDowntime,omitempty"`
	ReadyTimeout  string        `json:"readyTimeout,omitempty"` // e.g. 5m
}

// WorkloadMigrationResult reports the outcome of a backend migration.
type WorkloadMigrationResult struct {
	WorkloadID    string    `json:"workloadId"`
	SourceBackend string    `json:"sourceBackend"`
	TargetBackend string    `json:"targetBackend"`
	State         string    `json:"state"` // completed | failed
	CutoverAt     time.Time `json:"cutoverAt,omitempty"`
	Message       string    `json:"message,omitempty"`
}

// RuntimeMigrationOperations migrates workloads between backends with optional zero downtime.
type RuntimeMigrationOperations interface {
	MigrateWorkload(ctx context.Context, req WorkloadMigrationRequest) (*WorkloadMigrationResult, error)
}
