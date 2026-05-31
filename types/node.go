package types

import (
	"context"
	"time"
)

// Architecture identifies CPU platform for workload placement.
type Architecture string

const (
	ArchAMD64 Architecture = "amd64"
	ArchARM64 Architecture = "arm64"
)

// BackendHealthReport scores a runtime backend 0-100 from latency and error rate.
type BackendHealthReport struct {
	Backend   string    `json:"backend"`
	Score     int       `json:"score"`
	LatencyMs float64   `json:"latencyMs"`
	ErrorRate float64   `json:"errorRate"`
	Healthy   bool      `json:"healthy"`
	CheckedAt time.Time `json:"checkedAt"`
}

// NodeHealthReport scores a cluster node 0-100 based on readiness and pressure signals.
type NodeHealthReport struct {
	Name          string    `json:"name"`
	Backend       string    `json:"backend,omitempty"`
	Score         int       `json:"score"`
	Architecture  string    `json:"architecture,omitempty"`
	Ready         bool      `json:"ready"`
	Draining      bool      `json:"draining"`
	Unschedulable bool      `json:"unschedulable"`
	LatencyMs     float64   `json:"latencyMs,omitempty"`
	ErrorRate     float64   `json:"errorRate,omitempty"`
	Conditions    []string  `json:"conditions,omitempty"`
	CheckedAt     time.Time `json:"checkedAt"`
}

// NodeHealthListResponse lists node and backend health scores.
type NodeHealthListResponse struct {
	Nodes    []NodeHealthReport    `json:"nodes"`
	Backends []BackendHealthReport `json:"backends"`
	Count    int                   `json:"count"`
}

// NodeDrainRequest safely evicts workloads before node maintenance.
type NodeDrainRequest struct {
	NodeName           string `json:"nodeName"`
	GracePeriodSeconds int    `json:"gracePeriodSeconds,omitempty"`
	IgnoreDaemonSets   bool   `json:"ignoreDaemonSets,omitempty"`
	Reason             string `json:"reason,omitempty"`
}

// NodeDrainPhase tracks drain lifecycle.
type NodeDrainPhase string

const (
	DrainPhaseCordoning NodeDrainPhase = "cordoning"
	DrainPhaseEvicting  NodeDrainPhase = "evicting"
	DrainPhaseDrained   NodeDrainPhase = "drained"
	DrainPhaseFailed    NodeDrainPhase = "failed"
)

// NodeDrainResult is the outcome of a node drain operation.
type NodeDrainResult struct {
	NodeName      string         `json:"nodeName"`
	Phase         NodeDrainPhase `json:"phase"`
	PodsEvicted   int            `json:"podsEvicted"`
	PodsRemaining int            `json:"podsRemaining"`
	Message       string         `json:"message,omitempty"`
}

// NodeOperations extends runtime drivers with node lifecycle APIs.
type NodeOperations interface {
	ListBackendHealth(ctx context.Context) ([]BackendHealthReport, error)
	ListNodeHealth(ctx context.Context) ([]NodeHealthReport, error)
	DrainNode(ctx context.Context, req NodeDrainRequest) (*NodeDrainResult, error)
}
