package types

import "time"

// RateLimitConfig represents rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requestsPerSecond"`
	BurstSize         int           `json:"burstSize"`
	WindowDuration    time.Duration `json:"windowDuration"`
	Enabled           bool          `json:"enabled"`
}

// NamespaceQuota represents resource quotas for a namespace.
type NamespaceQuota struct {
	Namespace        string `json:"namespace"`
	MaxWorkloads     int64  `json:"maxWorkloads"`
	MaxCPU           int64  `json:"maxCPU"`     // in millicores
	MaxMemory        int64  `json:"maxMemory"`  // in MB
	MaxStorage       int64  `json:"maxStorage"` // in GB
	CurrentWorkloads int64  `json:"currentWorkloads"`
	CurrentCPU       int64  `json:"currentCPU"`
	CurrentMemory    int64  `json:"currentMemory"`
	CurrentStorage   int64  `json:"currentStorage"`
}

// NamespaceQuotaUsage represents current namespace quota usage.
type NamespaceQuotaUsage struct {
	Namespace     string  `json:"namespace"`
	WorkloadCount int64   `json:"workloadCount"`
	CPUUsage      float64 `json:"cpuUsage"`     // percentage
	MemoryUsage   float64 `json:"memoryUsage"`  // percentage
	StorageUsage  float64 `json:"storageUsage"` // percentage
	WorkloadLimit int64   `json:"workloadLimit"`
	CPULimit      int64   `json:"cpuLimit"`     // millicores
	MemoryLimit   int64   `json:"memoryLimit"`  // MB
	StorageLimit  int64   `json:"storageLimit"` // GB
}

// RateLimitInfo represents rate limit information for a client.
type RateLimitInfo struct {
	ClientID          string        `json:"clientId"`
	RequestsAllowed   int           `json:"requestsAllowed"`
	RequestsRemaining int           `json:"requestsRemaining"`
	ResetTime         time.Time     `json:"resetTime"`
	RetryAfter        time.Duration `json:"retryAfter"`
	LimitExceeded     bool          `json:"limitExceeded"`
}

// QuotaRequest represents a quota request.
type QuotaRequest struct {
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"` // workload, cpu, memory, storage
	Amount    int64  `json:"amount"`
}

// QuotaResponse represents a quota response.
type QuotaResponse struct {
	Approved bool                 `json:"approved"`
	Message  string               `json:"message"`
	Quota    *NamespaceQuotaUsage `json:"quota,omitempty"`
}
