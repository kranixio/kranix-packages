package types

// CostEstimateRequest estimates cost for a proposed deployment.
type CostEstimateRequest struct {
	Name      string       `json:"name,omitempty"`
	Namespace string       `json:"namespace,omitempty"`
	Spec      WorkloadSpec `json:"spec"`
	Duration  string       `json:"duration,omitempty"` // e.g. 30d, 7d
}

// CostBreakdownItem is one line in a cost breakdown.
type CostBreakdownItem struct {
	Resource string  `json:"resource"`
	Cost     float64 `json:"cost"`
	Usage    string  `json:"usage,omitempty"`
}

// RightsizingHint suggests resource adjustments based on utilization.
type RightsizingHint struct {
	RecommendedCPURequest string `json:"recommendedCpuRequest,omitempty"`
	RecommendedCPULimit   string `json:"recommendedCpuLimit,omitempty"`
	Reason                string `json:"reason,omitempty"`
}

// CostEstimateResponse is the result of a deployment cost estimate.
type CostEstimateResponse struct {
	WorkloadName          string              `json:"workloadName,omitempty"`
	Namespace               string              `json:"namespace,omitempty"`
	Duration                string              `json:"duration,omitempty"`
	TotalCost               float64             `json:"totalCost"`
	ComputeCost             float64             `json:"computeCost"`
	StorageCost             float64             `json:"storageCost"`
	NetworkCost             float64             `json:"networkCost"`
	UtilizationCPUPercent   float64             `json:"utilizationCpuPercent,omitempty"`
	Breakdown               []CostBreakdownItem `json:"breakdown,omitempty"`
	Rightsizing             *RightsizingHint    `json:"rightsizing,omitempty"`
	MonthlyCostPerReplica   float64             `json:"monthlyCostPerReplica,omitempty"`
	Message                 string              `json:"message,omitempty"`
}

// CostSummaryResponse aggregates cost across workloads in a namespace.
type CostSummaryResponse struct {
	Namespace        string                   `json:"namespace,omitempty"`
	Duration         string                   `json:"duration,omitempty"`
	TotalCost        float64                  `json:"totalCost"`
	WorkloadCount    int                      `json:"workloadCount"`
	AverageCost      float64                  `json:"averageCost"`
	TopCostWorkloads []map[string]interface{} `json:"topCostWorkloads,omitempty"`
	Message          string                   `json:"message,omitempty"`
}
