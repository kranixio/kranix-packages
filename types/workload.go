package types

import "time"

// Workload represents a managed workload in the Kranix system.
type Workload struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Namespace string         `json:"namespace"`
	Spec      WorkloadSpec   `json:"spec"`
	Status    WorkloadStatus `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// WorkloadSpec defines the desired configuration of a workload.
type WorkloadSpec struct {
	Image     string            `json:"image"`
	Replicas  int               `json:"replicas"`
	Env       map[string]string `json:"env,omitempty"`
	Resources ResourceSpec      `json:"resources,omitempty"`
	Ports     []PortSpec        `json:"ports,omitempty"`
	Backend   string            `json:"backend"` // docker | kubernetes
}

// ResourceSpec defines compute resource requirements.
type ResourceSpec struct {
	CPURequest    string `json:"cpuRequest,omitempty"`
	CPULimit      string `json:"cpuLimit,omitempty"`
	MemoryRequest string `json:"memoryRequest,omitempty"`
	MemoryLimit   string `json:"memoryLimit,omitempty"`
}

// PortSpec defines a port configuration.
type PortSpec struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"` // tcp | udp
}

// WorkloadStatus represents the current observed state of a workload.
type WorkloadStatus struct {
	Phase         WorkloadPhase `json:"phase"`
	ReadyReplicas int           `json:"readyReplicas"`
	Message       string        `json:"message,omitempty"`
	LastUpdated   time.Time     `json:"lastUpdated"`
}

// WorkloadPhase represents the lifecycle phase of a workload.
type WorkloadPhase string

const (
	WorkloadPhasePending   WorkloadPhase = "Pending"
	WorkloadPhaseDeploying WorkloadPhase = "Deploying"
	WorkloadPhaseRunning   WorkloadPhase = "Running"
	WorkloadPhaseDegraded  WorkloadPhase = "Degraded"
	WorkloadPhaseFailed    WorkloadPhase = "Failed"
)
