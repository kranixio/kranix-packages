package types

// Well-known label keys for workload tagging (filtering, billing, quotas).
const (
	LabelKeyTeam        = "kranix.io/team"
	LabelKeyEnvironment = "kranix.io/environment"
	LabelKeyCostCenter  = "kranix.io/cost-center"
)

// WorkloadTags groups structured tags for team, environment, and cost center.
type WorkloadTags struct {
	Team        string            `json:"team,omitempty"`
	Environment string            `json:"environment,omitempty"`
	CostCenter  string            `json:"costCenter,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}
