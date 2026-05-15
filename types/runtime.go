package types

import "context"

// RuntimeDriver defines the interface that all runtime backends must implement.
type RuntimeDriver interface {
	// Workload operations
	Deploy(ctx context.Context, spec *WorkloadSpec) (*WorkloadStatus, error)
	Destroy(ctx context.Context, workloadID string) error
	Restart(ctx context.Context, workloadID string) error

	// Observation
	GetStatus(ctx context.Context, workloadID string) (*WorkloadStatus, error)
	ListWorkloads(ctx context.Context, namespace string) ([]*WorkloadStatus, error)
	StreamLogs(ctx context.Context, podID string, opts *LogOptions) (<-chan string, error)

	// Lifecycle
	Ping(ctx context.Context) error
	Backend() string
}
