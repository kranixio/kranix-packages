package types

import (
	"context"
	"time"
)

// VolumeSpec defines a persistent volume for a workload.
type VolumeSpec struct {
	Name         string `json:"name"`
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
	MountPath    string `json:"mountPath"`
	AccessMode   string `json:"accessMode,omitempty"` // ReadWriteOnce | ReadWriteMany
	AutoCleanup  bool   `json:"autoCleanup,omitempty"`
}

// NetworkBandwidthSpec limits network egress (and optional ingress) per workload.
type NetworkBandwidthSpec struct {
	EgressLimit   string `json:"egressLimit,omitempty"`   // e.g. 10Mbit, 1Gbit
	IngressLimit  string `json:"ingressLimit,omitempty"`  // e.g. 5Mbit
	Enabled       bool   `json:"enabled,omitempty"`
}

// CheckpointRequest captures running container state.
type CheckpointRequest struct {
	WorkloadID string `json:"workloadId"`
	Namespace  string `json:"namespace,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

// CheckpointResult is the outcome of a checkpoint operation.
type CheckpointResult struct {
	WorkloadID   string    `json:"workloadId"`
	CheckpointID string    `json:"checkpointId"`
	State        string    `json:"state"` // paused | checkpointed
	CreatedAt    time.Time `json:"createdAt"`
	Message      string    `json:"message,omitempty"`
}

// RestoreRequest resumes a workload from a checkpoint.
type RestoreRequest struct {
	WorkloadID   string `json:"workloadId"`
	Namespace    string `json:"namespace,omitempty"`
	CheckpointID string `json:"checkpointId,omitempty"`
}

// RestoreResult is the outcome of a restore operation.
type RestoreResult struct {
	WorkloadID string    `json:"workloadId"`
	State      string    `json:"state"`
	RestoredAt time.Time `json:"restoredAt"`
	Message    string    `json:"message,omitempty"`
}

// VolumeLifecycleResult reports provisioned volumes for a workload.
type VolumeLifecycleResult struct {
	WorkloadID string        `json:"workloadId"`
	Volumes    []VolumeState `json:"volumes"`
}

// VolumeState tracks one attached volume.
type VolumeState struct {
	Name      string `json:"name"`
	ClaimName string `json:"claimName,omitempty"`
	VolumeID  string `json:"volumeId,omitempty"`
	MountPath string `json:"mountPath"`
	Status    string `json:"status"` // bound | pending | released
}

// RuntimeExtendedOperations provides checkpoint, bandwidth-aware deploy helpers, and volume lifecycle.
type RuntimeExtendedOperations interface {
	CheckpointWorkload(ctx context.Context, req CheckpointRequest) (*CheckpointResult, error)
	RestoreWorkload(ctx context.Context, req RestoreRequest) (*RestoreResult, error)
	ListCheckpoints(ctx context.Context, workloadID, namespace string) ([]CheckpointResult, error)
	ProvisionVolumes(ctx context.Context, spec *WorkloadSpec) (*VolumeLifecycleResult, error)
	CleanupVolumes(ctx context.Context, spec *WorkloadSpec) error
}

// RuntimePluginInfo describes a registered runtime backend plugin.
type RuntimePluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Builtin     bool   `json:"builtin"`
	Enabled     bool   `json:"enabled"`
}

// RuntimePluginListResponse lists available runtime plugins.
type RuntimePluginListResponse struct {
	Plugins []RuntimePluginInfo `json:"plugins"`
	Count   int                 `json:"count"`
}
