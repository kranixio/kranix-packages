package types

import "time"

// AnalyticsMetrics represents usage analytics metrics.
type AnalyticsMetrics struct {
	MetricType    string                 `json:"metricType"`
	ResourceID    string                 `json:"resourceId"`
	ResourceType  string                 `json:"resourceType"` // workload, namespace, tenant
	Timestamp     time.Time              `json:"timestamp"`
	Value         float64                `json:"value"`
	Labels        map[string]string      `json:"labels,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DeployMetrics represents deployment-related metrics.
type DeployMetrics struct {
	WorkloadID    string        `json:"workloadId"`
	Namespace     string        `json:"namespace"`
	DeployCount   int64         `json:"deployCount"`
	SuccessCount  int64         `json:"successCount"`
	FailureCount  int64         `json:"failureCount"`
	SuccessRate   float64       `json:"successRate"`
	AverageLatency time.Duration `json:"averageLatency"`
	LastDeployAt  time.Time     `json:"lastDeployAt"`
	WindowStart   time.Time     `json:"windowStart"`
	WindowEnd     time.Time     `json:"windowEnd"`
}

// ErrorMetrics represents error-related metrics.
type ErrorMetrics struct {
	WorkloadID      string    `json:"workloadId"`
	Namespace       string    `json:"namespace"`
	ErrorCount      int64     `json:"errorCount"`
	ErrorRate       float64   `json:"errorRate"`
	ErrorTypes      map[string]int64 `json:"errorTypes,omitempty"`
	LastErrorAt     time.Time `json:"lastErrorAt"`
	LastErrorType   string    `json:"lastErrorType"`
	WindowStart     time.Time `json:"windowStart"`
	WindowEnd       time.Time `json:"windowEnd"`
}

// LatencyMetrics represents latency-related metrics.
type LatencyMetrics struct {
	WorkloadID      string        `json:"workloadId"`
	Namespace       string        `json:"namespace"`
	P50Latency      time.Duration `json:"p50Latency"`
	P95Latency      time.Duration `json:"p95Latency"`
	P99Latency      time.Duration `json:"p99Latency"`
	AverageLatency  time.Duration `json:"averageLatency"`
	MaxLatency      time.Duration `json:"maxLatency"`
	WindowStart     time.Time     `json:"windowStart"`
	WindowEnd       time.Time     `json:"windowEnd"`
}

// UsageSummary represents a summary of usage metrics.
type UsageSummary struct {
	TotalWorkloads   int64         `json:"totalWorkloads"`
	TotalDeploys     int64         `json:"totalDeploys"`
	TotalErrors      int64         `json:"totalErrors"`
	AverageLatency   time.Duration `json:"averageLatency"`
	WindowStart      time.Time     `json:"windowStart"`
	WindowEnd        time.Time     `json:"windowEnd"`
	ByNamespace      map[string]*NamespaceUsage `json:"byNamespace,omitempty"`
	ByTenant         map[string]*TenantUsage    `json:"byTenant,omitempty"`
}

// NamespaceUsage represents usage metrics for a namespace.
type NamespaceUsage struct {
	Namespace      string        `json:"namespace"`
	WorkloadCount  int64         `json:"workloadCount"`
	DeployCount    int64         `json:"deployCount"`
	ErrorCount     int64         `json:"errorCount"`
	AverageLatency time.Duration `json:"averageLatency"`
}

// TenantUsage represents usage metrics for a tenant.
type TenantUsage struct {
	TenantID       string        `json:"tenantId"`
	TenantName     string        `json:"tenantName"`
	WorkloadCount  int64         `json:"workloadCount"`
	DeployCount    int64         `json:"deployCount"`
	ErrorCount     int64         `json:"errorCount"`
	AverageLatency time.Duration `json:"averageLatency"`
	QuotaUsage     *QuotaUsage   `json:"quotaUsage,omitempty"`
}

// QuotaUsage represents quota utilization.
type QuotaUsage struct {
	CPUUsage       float64 `json:"cpuUsage"`       // percentage
	MemoryUsage    float64 `json:"memoryUsage"`    // percentage
	StorageUsage   float64 `json:"storageUsage"`   // percentage
	WorkloadUsage  float64 `json:"workloadUsage"`  // percentage
}

// AnalyticsQuery represents a query for analytics data.
type AnalyticsQuery struct {
	ResourceType  string    `json:"resourceType"`  // workload, namespace, tenant
	ResourceID    string    `json:"resourceId,omitempty"`
	MetricType    string    `json:"metricType"`    // deploy, error, latency
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Granularity   string    `json:"granularity"`   // minute, hour, day
	GroupBy       []string  `json:"groupBy,omitempty"` // namespace, tenant
	Filters       map[string]string `json:"filters,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
}
