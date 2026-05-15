package types

import "time"

// Pod represents a runtime pod/container instance.
type Pod struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	WorkloadID  string            `json:"workloadId"`
	Node        string            `json:"node,omitempty"`
	Phase       PodPhase          `json:"phase"`
	IP          string            `json:"ip,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// PodPhase represents the lifecycle phase of a pod.
type PodPhase string

const (
	PodPhasePending   PodPhase = "Pending"
	PodPhaseRunning   PodPhase = "Running"
	PodPhaseSucceeded PodPhase = "Succeeded"
	PodPhaseFailed    PodPhase = "Failed"
	PodPhaseUnknown   PodPhase = "Unknown"
)

// PodSpec defines the desired state of a pod.
type PodSpec struct {
	Containers []ContainerSpec `json:"containers"`
}

// ContainerSpec defines a container in a pod.
type ContainerSpec struct {
	Name  string            `json:"name"`
	Image string            `json:"image"`
	Env   map[string]string `json:"env,omitempty"`
}

// PodStatus represents the current observed state of a pod.
type PodStatus struct {
	Phase      PodPhase          `json:"phase"`
	Message    string            `json:"message,omitempty"`
	Conditions []PodCondition    `json:"conditions,omitempty"`
}

// PodCondition describes the state of a pod at a certain point.
type PodCondition struct {
	Type               PodConditionType `json:"type"`
	Status             ConditionStatus  `json:"status"`
	LastTransitionTime time.Time        `json:"lastTransitionTime"`
	Reason             string           `json:"reason,omitempty"`
	Message            string           `json:"message,omitempty"`
}

// PodConditionType defines the type of pod condition.
type PodConditionType string

const (
	PodReady            PodConditionType = "Ready"
	PodInitialized      PodConditionType = "Initialized"
	PodContainersReady  PodConditionType = "ContainersReady"
	PodScheduled        PodConditionType = "PodScheduled"
)

// ConditionStatus defines the status of a condition.
type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// LogOptions defines options for streaming pod logs.
type LogOptions struct {
	Follow bool  `json:"follow"`
	Tail   int64 `json:"tail,omitempty"`
	Since  int64 `json:"since,omitempty"`
}
