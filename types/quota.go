package types

// HardResourceQuota defines hard aggregate limits per namespace or team (label kranix.io/team), aligned with kranix-core admission.
type HardResourceQuota struct {
	Namespace string `json:"namespace,omitempty"`
	TeamID    string `json:"teamId,omitempty"`
	// Resource requests (Kubernetes-style quantities).
	MaxCPURequests    string `json:"maxCpuRequests,omitempty"`
	MaxMemoryRequests string `json:"maxMemoryRequests,omitempty"`
	MaxWorkloads      int32  `json:"maxWorkloads,omitempty"`
	MaxReplicasTotal  int32  `json:"maxReplicasTotal,omitempty"`
}
