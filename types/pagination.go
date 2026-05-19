package types

// PageInfo describes cursor pagination for list endpoints.
type PageInfo struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
	TotalCount int    `json:"totalCount"`
}

// PaginatedWorkloadListResponse is a cursor-paginated workload list.
type PaginatedWorkloadListResponse struct {
	Workloads []Workload          `json:"workloads"`
	PageInfo  PageInfo            `json:"pageInfo"`
	Query     WorkloadSearchQuery `json:"query,omitempty"`
}
