package types

// WorkloadSearchQuery filters workloads in list/search APIs.
type WorkloadSearchQuery struct {
	// AllNamespaces when true lists workloads across every namespace (ignores Namespace filter).
	AllNamespaces bool   `json:"allNamespaces,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	Phase       string `json:"phase,omitempty"`       // Pending | Running | Degraded | Failed
	Status      string `json:"status,omitempty"`      // alias for phase
	Image       string `json:"image,omitempty"`       // substring match
	Team        string `json:"team,omitempty"`
	Environment string `json:"environment,omitempty"`
	CostCenter  string `json:"costCenter,omitempty"`
	LabelKey    string `json:"labelKey,omitempty"`
	LabelValue  string `json:"labelValue,omitempty"`
}

// WorkloadListResponse is returned by filtered workload list endpoints (non-paginated legacy).
type WorkloadListResponse struct {
	Workloads []Workload          `json:"workloads"`
	Count     int                 `json:"count"`
	Query     WorkloadSearchQuery `json:"query,omitempty"`
	PageInfo  *PageInfo           `json:"pageInfo,omitempty"`
}
