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

// ResourceQuotaUsage reports limits and current aggregate usage for a namespace.
type ResourceQuotaUsage struct {
	Namespace string            `json:"namespace"`
	Limits    HardResourceQuota `json:"limits"`
	Used      QuotaUsageTotals  `json:"used"`
}

// QuotaUsageTotals is observed consumption against a namespace quota.
type QuotaUsageTotals struct {
	WorkloadCount  int    `json:"workloadCount"`
	ReplicaCount   int32  `json:"replicaCount"`
	CPURequests    string `json:"cpuRequests,omitempty"`
	MemoryRequests string `json:"memoryRequests,omitempty"`
}

// ResourceQuotaListResponse lists namespace quotas.
type ResourceQuotaListResponse struct {
	Quotas []HardResourceQuota `json:"quotas"`
	Count  int                 `json:"count"`
}
