package types

// WorkloadSearchQuery filters workloads in list/search APIs.
type WorkloadSearchQuery struct {
	Namespace   string `json:"namespace,omitempty"`
	Phase       string `json:"phase,omitempty"`       // Pending | Running | Degraded | Failed
	Status      string `json:"status,omitempty"`      // alias for phase
	Image       string `json:"image,omitempty"`       // substring match
	Team        string `json:"team,omitempty"`
	Environment string `json:"environment,omitempty"`
	CostCenter  string `json:"costCenter,omitempty"`
	LabelKey    string `json:"labelKey,omitempty"`
	LabelValue  string `json:"labelValue,omitempty"`
}

// WorkloadListResponse is returned by filtered workload list endpoints.
type WorkloadListResponse struct {
	Workloads []Workload     `json:"workloads"`
	Count     int            `json:"count"`
	Query     WorkloadSearchQuery `json:"query,omitempty"`
}
