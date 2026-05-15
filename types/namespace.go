package types

import "time"

// Namespace represents a logical isolation boundary for workloads.
type Namespace struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName,omitempty"`
	Description string            `json:"description,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Labels      map[string]string `json:"labels,omitempty"`
	Status      NamespaceStatus   `json:"status"`
}

// NamespaceStatus represents the status of a namespace.
type NamespaceStatus struct {
	Phase   NamespacePhase `json:"phase"`
	Message string         `json:"message,omitempty"`
}

// NamespacePhase represents the lifecycle phase of a namespace.
type NamespacePhase string

const (
	NamespacePhaseActive  NamespacePhase = "Active"
	NamespacePhaseTerminating NamespacePhase = "Terminating"
	NamespacePhaseTerminated NamespacePhase = "Terminated"
)
