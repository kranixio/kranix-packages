package runtime

import (
	"context"
	"time"

	"github.com/kranix-io/kranix-packages/types"
)

// RuntimeDriver defines the interface for runtime backends.
// Implemented by kranix-runtime backends.
type RuntimeDriver interface {
	// Deploy deploys a workload to the runtime backend.
	Deploy(ctx context.Context, spec *types.WorkloadSpec) (*types.WorkloadStatus, error)
	// Destroy destroys a workload from the runtime backend.
	Destroy(ctx context.Context, workloadID string) error
	// Restart restarts a workload in the runtime backend.
	Restart(ctx context.Context, workloadID string) error
	// GetStatus retrieves the current status of a workload.
	GetStatus(ctx context.Context, workloadID string) (*types.WorkloadStatus, error)
	// ListWorkloads lists all workloads in a namespace.
	ListWorkloads(ctx context.Context, namespace string) ([]*types.WorkloadStatus, error)
	// StreamLogs streams logs from a pod.
	StreamLogs(ctx context.Context, podID string, opts *types.LogOptions) (<-chan string, error)
	// Ping checks if the runtime backend is reachable.
	Ping(ctx context.Context) error
	// Backend returns the name of the runtime backend.
	Backend() string
}

// RuntimeConfig defines configuration for runtime backends.
type RuntimeConfig struct {
	BackendType string            `json:"backendType"` // docker, kubernetes
	Endpoint    string            `json:"endpoint,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

// RuntimeStatus represents the status of a runtime backend.
type RuntimeStatus struct {
	Backend   string    `json:"backend"`
	Healthy   bool      `json:"healthy"`
	Version   string    `json:"version,omitempty"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
