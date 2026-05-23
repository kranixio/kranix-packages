package types

import "time"

// Status represents a generic resource status.
type Status struct {
	Phase       string    `json:"phase"`
	Conditions  []Condition `json:"conditions,omitempty"`
	Message     string    `json:"message,omitempty"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// Condition represents the state of a resource at a certain point.
type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"` // True, False, Unknown
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

// AnalysisResult represents the result of a workload analysis.
type AnalysisResult struct {
	WorkloadID   string    `json:"workloadId"`
	Namespace    string    `json:"namespace,omitempty"`
	Status       string    `json:"status"`
	Issues       []Issue   `json:"issues,omitempty"`
	Suggestions  []string  `json:"suggestions,omitempty"`
	ProbableFix  string    `json:"probableFix,omitempty"`
	AnalyzedAt   time.Time `json:"analyzedAt"`
}

// Issue represents a detected issue with a workload.
type Issue struct {
	Severity string `json:"severity"` // error, warning, info
	Type     string `json:"type"`
	Message  string `json:"message"`
	Field    string `json:"field,omitempty"`
}
