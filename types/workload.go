package types

import "time"

// Workload represents a managed workload in the Kranix system.
type Workload struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Spec      WorkloadSpec      `json:"spec"`
	Status    WorkloadStatus    `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Labels    map[string]string `json:"labels,omitempty"`
	Tags      *WorkloadTags     `json:"tags,omitempty"`
	Tenant    *TenantInfo       `json:"tenant,omitempty"`
	// RollbackVersions holds the last N spec snapshots (newest first).
	RollbackVersions []WorkloadRevision `json:"rollbackVersions,omitempty"`
}

// WorkloadSpec defines the desired configuration of a workload.
type WorkloadSpec struct {
	Name              string             `json:"name"`
	Namespace         string             `json:"namespace,omitempty"`
	Image             string             `json:"image"`
	Replicas          int                `json:"replicas"`
	Env               map[string]string  `json:"env,omitempty"`
	Command           string             `json:"command,omitempty"`
	Resources         ResourceSpec       `json:"resources,omitempty"`
	Ports             []PortSpec         `json:"ports,omitempty"`
	Backend           string             `json:"backend"` // docker | kubernetes | podman | compose | remote
	ComposeFile       string             `json:"composeFile,omitempty"`
	RemoteHost        string             `json:"remoteHost,omitempty"` // For remote SSH backend
	RolloutStrategy   RolloutStrategy    `json:"rolloutStrategy,omitempty"`
	AutoScaling       *AutoScalingConfig `json:"autoScaling,omitempty"`
	Scheduling        *SchedulingConfig  `json:"scheduling,omitempty"`
	Dependencies      []Dependency       `json:"dependencies,omitempty"`
	FailurePrediction *FailurePrediction `json:"failurePrediction,omitempty"`
	// CrossNamespaceTraffic restricts namespace-to-namespace traffic when enforced by Kubernetes NetworkPolicy drivers.
	CrossNamespaceTraffic *CrossNamespaceTrafficPolicy `json:"crossNamespaceTraffic,omitempty"`
	// CronSchedule enables cron-style scheduling in core; Kubernetes runtime maps this to a CronJob when set.
	CronSchedule *CronScheduleSpec `json:"cronSchedule,omitempty"`
	// Tags classify the workload (team, environment, cost center); mirrored to standard kranix.io/* labels on Kubernetes.
	Tags *WorkloadTags `json:"tags,omitempty"`
	// CircuitBreaker stops routing while unhealthy or when the circuit is open.
	CircuitBreaker *CircuitBreakerSpec `json:"circuitBreaker,omitempty"`
	// WarmStandby keeps a cold replica workload ready for instant failover.
	WarmStandby *WarmStandbySpec `json:"warmStandby,omitempty"`
	// SecretRotation triggers rolling restarts when referenced secrets change.
	SecretRotation *SecretRotationSpec `json:"secretRotation,omitempty"`
	// Labels are optional workload labels propagated to Kubernetes (e.g. warm standby role).
	Labels map[string]string `json:"labels,omitempty"`
	// Volumes defines persistent volumes to auto-create, attach, and optionally clean up.
	Volumes []VolumeSpec `json:"volumes,omitempty"`
	// NetworkBandwidth limits egress/ingress per workload when supported by the backend.
	NetworkBandwidth *NetworkBandwidthSpec `json:"networkBandwidth,omitempty"`
}

// ResourceSpec defines compute resource requirements.
type ResourceSpec struct {
	CPURequest    string   `json:"cpuRequest,omitempty"`
	CPULimit      string   `json:"cpuLimit,omitempty"`
	MemoryRequest string   `json:"memoryRequest,omitempty"`
	MemoryLimit   string   `json:"memoryLimit,omitempty"`
	GPU           *GPUSpec `json:"gpu,omitempty"`
}

// GPUSpec defines GPU resource requirements.
type GPUSpec struct {
	Vendor string `json:"vendor"`           // nvidia | amd
	Count  int32  `json:"count"`            // Number of GPUs
	Type   string `json:"type,omitempty"`   // GPU type (e.g., "A100", "V100", "MI250")
	SKU    string `json:"sku,omitempty"`    // GPU SKU for specific models
	Memory string `json:"memory,omitempty"` // GPU memory requirement (e.g., "16Gi")
}

// PortSpec defines a port configuration.
type PortSpec struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"` // tcp | udp
}

// WorkloadStatus represents the current observed state of a workload.
type WorkloadStatus struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Namespace     string        `json:"namespace,omitempty"`
	State         string        `json:"state"`
	Image         string        `json:"image,omitempty"`
	Replicas      int           `json:"replicas,omitempty"`
	Ready         int           `json:"ready,omitempty"`
	Host          string        `json:"host,omitempty"`
	Pods          []string      `json:"pods,omitempty"`
	Phase         WorkloadPhase `json:"phase"`
	ReadyReplicas int           `json:"readyReplicas"`
	Message       string        `json:"message,omitempty"`
	LastUpdated   time.Time     `json:"lastUpdated"`
	Cron          *CronScheduleStatus `json:"cron,omitempty"`
	Rollback        *RollbackHistoryStatus `json:"rollback,omitempty"`
	CircuitBreaker  *CircuitBreakerStatus  `json:"circuitBreaker,omitempty"`
	WarmStandby     *WarmStandbyStatus     `json:"warmStandby,omitempty"`
	SecretRotation  *SecretRotationStatus  `json:"secretRotation,omitempty"`
}

// CronScheduleSpec defines optional periodic execution (standard 5-field cron, e.g. "0 * * * *").
type CronScheduleSpec struct {
	Schedule          string `json:"schedule"`
	Suspended         bool   `json:"suspended,omitempty"`
	TimeZone          string `json:"timeZone,omitempty"`
	ConcurrencyPolicy string `json:"concurrencyPolicy,omitempty"` // Allow | Forbid | Replace
}

// CronScheduleStatus records observed cron trigger metadata (e.g. from core).
type CronScheduleStatus struct {
	LastScheduleTime *time.Time `json:"lastScheduleTime,omitempty"`
}

// WorkloadPhase represents the lifecycle phase of a workload.
type WorkloadPhase string

const (
	WorkloadPhasePending   WorkloadPhase = "Pending"
	WorkloadPhaseDeploying WorkloadPhase = "Deploying"
	WorkloadPhaseRunning   WorkloadPhase = "Running"
	WorkloadPhaseDegraded  WorkloadPhase = "Degraded"
	WorkloadPhaseFailed    WorkloadPhase = "Failed"
)

// RolloutStrategy defines how the workload should be deployed.
type RolloutStrategy struct {
	Type           string        `json:"type"` // rolling, recreate, bluegreen, canary, abtest
	MaxUnavailable int32         `json:"maxUnavailable,omitempty"`
	MaxSurge       int32         `json:"maxSurge,omitempty"`
	CanaryConfig   *CanaryConfig `json:"canaryConfig,omitempty"`
	ABTestConfig   *ABTestConfig `json:"abTestConfig,omitempty"`
}

// AutoScalingConfig defines auto-scaling behavior.
type AutoScalingConfig struct {
	Enabled                  bool           `json:"enabled"`
	MinReplicas              int32          `json:"minReplicas"`
	MaxReplicas              int32          `json:"maxReplicas"`
	TargetCPUUtilization     int32          `json:"targetCPUUtilization,omitempty"`    // percentage
	TargetMemoryUtilization  int32          `json:"targetMemoryUtilization,omitempty"` // percentage
	CustomMetrics            []CustomMetric `json:"customMetrics,omitempty"`
	ScaleDownCooldownSeconds int32          `json:"scaleDownCooldownSeconds,omitempty"`
	ScaleUpCooldownSeconds   int32          `json:"scaleUpCooldownSeconds,omitempty"`
}

// CustomMetric defines a custom metric for auto-scaling.
type CustomMetric struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"` // pods, object
	MetricName string       `json:"metricName"`
	Target     MetricTarget `json:"target"`
}

// MetricTarget defines the target value for a metric.
type MetricTarget struct {
	Type         string `json:"type"` // average, value
	AverageValue string `json:"averageValue,omitempty"`
	Value        string `json:"value,omitempty"`
}

// SchedulingConfig defines scheduling preferences.
type SchedulingConfig struct {
	CostAware         bool                `json:"costAware,omitempty"`
	PreferredRegions  []string            `json:"preferredRegions,omitempty"`
	PreferredZones    []string            `json:"preferredZones,omitempty"`
	NodeSelectors     map[string]string   `json:"nodeSelectors,omitempty"`
	Affinity          *AffinityConfig     `json:"affinity,omitempty"`
	Tolerations       []Toleration        `json:"tolerations,omitempty"`
	MaxCostPerHour    string              `json:"maxCostPerHour,omitempty"`
	WorkloadPriority  string              `json:"workloadPriority,omitempty"`
	PreemptionEnabled bool                `json:"preemptionEnabled,omitempty"`
	PriorityClassName string              `json:"priorityClassName,omitempty"`
	Spot              *SpotWorkloadConfig `json:"spot,omitempty"`
	// Architecture routes workloads to amd64 or arm64 nodes (kubernetes.io/arch).
	Architecture string `json:"architecture,omitempty"`
	// AvoidDrainingNodes excludes nodes marked for maintenance from placement.
	AvoidDrainingNodes bool `json:"avoidDrainingNodes,omitempty"`
}

// WorkloadPriority enumerates coarse scheduling tiers.
type WorkloadPriority string

const (
	WorkloadPriorityCritical WorkloadPriority = "critical"
	WorkloadPriorityHigh     WorkloadPriority = "high"
	WorkloadPriorityNormal   WorkloadPriority = "normal"
	WorkloadPriorityLow      WorkloadPriority = "low"
)

// SpotWorkloadConfig configures spot/preemptible placement.
type SpotWorkloadConfig struct {
	Enabled                     bool `json:"enabled,omitempty"`
	RescheduleOnNodeTermination bool `json:"rescheduleOnNodeTermination,omitempty"`
}

// CrossNamespaceTrafficPolicy controls which namespaces may exchange traffic with this workload via NetworkPolicy.
type CrossNamespaceTrafficPolicy struct {
	Enabled                  bool     `json:"enabled,omitempty"`
	AllowedIngressNamespaces []string `json:"allowedIngressNamespaces,omitempty"`
	AllowedEgressNamespaces  []string `json:"allowedEgressNamespaces,omitempty"`
	AllowSameNamespace       *bool    `json:"allowSameNamespace,omitempty"`
	BlockClusterDNS          bool     `json:"blockClusterDNS,omitempty"`
	AllowEgressInternet      bool     `json:"allowEgressInternet,omitempty"`
}

// AffinityConfig defines pod affinity/anti-affinity rules.
type AffinityConfig struct {
	NodeAffinity    *NodeAffinity `json:"nodeAffinity,omitempty"`
	PodAffinity     *PodAffinity  `json:"podAffinity,omitempty"`
	PodAntiAffinity *PodAffinity  `json:"podAntiAffinity,omitempty"`
}

// NodeAffinity defines node affinity rules.
type NodeAffinity struct {
	RequiredDuringScheduling  []NodeSelectorTerm        `json:"requiredDuringScheduling,omitempty"`
	PreferredDuringScheduling []PreferredSchedulingTerm `json:"preferredDuringScheduling,omitempty"`
}

// NodeSelectorTerm defines a node selector term.
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions,omitempty"`
	MatchFields      []NodeSelectorRequirement `json:"matchFields,omitempty"`
}

// NodeSelectorRequirement defines a node selector requirement.
type NodeSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"` // In, NotIn, Exists, DoesNotExist, Gt, Lt
	Values   []string `json:"values,omitempty"`
}

// PreferredSchedulingTerm defines a preferred scheduling term.
type PreferredSchedulingTerm struct {
	Weight     int32            `json:"weight"`
	Preference NodeSelectorTerm `json:"preference"`
}

// PodAffinity defines pod affinity rules.
type PodAffinity struct {
	RequiredDuringScheduling  []PodAffinityTerm         `json:"requiredDuringScheduling,omitempty"`
	PreferredDuringScheduling []WeightedPodAffinityTerm `json:"preferredDuringScheduling,omitempty"`
}

// PodAffinityTerm defines a pod affinity term.
type PodAffinityTerm struct {
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
	Namespaces    []string          `json:"namespaces,omitempty"`
	TopologyKey   string            `json:"topologyKey"`
}

// WeightedPodAffinityTerm defines a weighted pod affinity term.
type WeightedPodAffinityTerm struct {
	Weight          int32           `json:"weight"`
	PodAffinityTerm PodAffinityTerm `json:"podAffinityTerm"`
}

// Toleration defines a toleration for taints.
type Toleration struct {
	Key               string `json:"key,omitempty"`
	Operator          string `json:"operator,omitempty"` // Exists, Equal
	Value             string `json:"value,omitempty"`
	Effect            string `json:"effect,omitempty"` // NoSchedule, PreferNoSchedule, NoExecute
	TolerationSeconds *int64 `json:"tolerationSeconds,omitempty"`
}

// CanaryConfig defines canary deployment configuration.
type CanaryConfig struct {
	Replicas         int32    `json:"replicas"`
	Percentage       int32    `json:"percentage,omitempty"`
	AnalysisDuration string   `json:"analysisDuration,omitempty"`
	SuccessThreshold int32    `json:"successThreshold,omitempty"`
	Metrics          []string `json:"metrics,omitempty"`
	AutoPromote      bool     `json:"autoPromote,omitempty"`
}

// ABTestConfig defines A/B testing configuration.
type ABTestConfig struct {
	VariantA         string   `json:"variantA"`
	VariantB         string   `json:"variantB"`
	TrafficSplit     int32    `json:"trafficSplit"` // percentage for variant B
	AnalysisDuration string   `json:"analysisDuration,omitempty"`
	Metrics          []string `json:"metrics,omitempty"`
	AutoSelectWinner bool     `json:"autoSelectWinner,omitempty"`
}

// Dependency defines a dependency relationship between workloads.
type Dependency struct {
	WorkloadID string `json:"workloadId"`
	Type       string `json:"type"`                // depends_on, waits_for, requires
	Condition  string `json:"condition,omitempty"` // running, healthy, ready
	Timeout    string `json:"timeout,omitempty"`   // duration string like "5m"
}

// FailurePrediction defines ML-based failure prediction configuration.
type FailurePrediction struct {
	Enabled           bool     `json:"enabled"`
	ModelType         string   `json:"modelType"`                   // simple, ml, custom
	PredictionWindow  string   `json:"predictionWindow,omitempty"`  // time window for prediction
	Threshold         float64  `json:"threshold"`                   // probability threshold (0-1)
	Features          []string `json:"features,omitempty"`          // cpu_usage, memory_usage, request_rate, error_rate
	MitigationActions []string `json:"mitigationActions,omitempty"` // scale_up, restart, migrate
}

// TenantInfo defines multi-tenancy information for a workload.
type TenantInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
	Quota     *TenantQuota      `json:"quota,omitempty"`
	Isolation *TenantIsolation  `json:"isolation,omitempty"`
}

// TenantQuota defines resource quotas for a tenant.
type TenantQuota struct {
	MaxCPU           string `json:"maxCPU,omitempty"`
	MaxMemory        string `json:"maxMemory,omitempty"`
	MaxWorkloads     int32  `json:"maxWorkloads,omitempty"`
	MaxReplicas      int32  `json:"maxReplicas,omitempty"`
	MaxStorage       string `json:"maxStorage,omitempty"`
	MaxCustomMetrics int32  `json:"maxCustomMetrics,omitempty"`
}

// TenantIsolation defines isolation mechanisms for a tenant.
type TenantIsolation struct {
	NetworkPolicy     bool   `json:"networkPolicy"`
	ResourceQuota     bool   `json:"resourceQuota"`
	LimitRange        bool   `json:"limitRange"`
	PodSecurityPolicy bool   `json:"podSecurityPolicy"`
	StorageClass      string `json:"storageClass,omitempty"`
}

// EphemeralEnvironmentSpec defines ephemeral environment lifecycle configuration.
type EphemeralEnvironmentSpec struct {
	Enabled         bool   `json:"enabled"`
	TriggerType     string `json:"triggerType"`   // pull_request | branch_push | manual
	TriggerSource   string `json:"triggerSource"` // github | gitlab | bitbucket
	PRNumber        int32  `json:"prNumber,omitempty"`
	BranchName      string `json:"branchName,omitempty"`
	CommitSHA       string `json:"commitSHA,omitempty"`
	TTL             string `json:"ttl"` // Time to live (e.g., "2h", "24h")
	AutoTeardown    bool   `json:"autoTeardown"`
	TeardownOnMerge bool   `json:"teardownOnMerge"`
	TeardownOnClose bool   `json:"teardownOnClose"`
	MaxEnvironments int32  `json:"maxEnvironments"` // Max concurrent ephemeral envs
	NamespacePrefix string `json:"namespacePrefix"` // e.g., "pr-" or "ephem-"
}

// EphemeralEnvironmentStatus represents the status of an ephemeral environment.
type EphemeralEnvironmentStatus struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Namespace         string     `json:"namespace"`
	Phase             string     `json:"phase"` // Creating | Ready | Terminating | Terminated
	TriggerType       string     `json:"triggerType"`
	PRNumber          int32      `json:"prNumber,omitempty"`
	BranchName        string     `json:"branchName,omitempty"`
	CommitSHA         string     `json:"commitSHA,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	ExpiresAt         time.Time  `json:"expiresAt"`
	TerminatedAt      *time.Time `json:"terminatedAt,omitempty"`
	TerminationReason string     `json:"terminationReason,omitempty"`
}

// EdgeNodeSpec defines edge node agent configuration.
type EdgeNodeSpec struct {
	NodeID            string            `json:"nodeID"`
	NodeName          string            `json:"nodeName"`
	IPAddress         string            `json:"ipAddress"`
	Port              int32             `json:"port"`
	Labels            map[string]string `json:"labels,omitempty"`
	Capabilities      []string          `json:"capabilities,omitempty"` // gpu | storage | compute
	Architecture      string            `json:"architecture"`           // amd64 | arm64
	OS                string            `json:"os"`                     // linux | windows
	Resources         *ResourceSpec     `json:"resources,omitempty"`
	HeartbeatInterval string            `json:"heartbeatInterval"` // e.g., "30s"
	AuthToken         string            `json:"authToken,omitempty"`
}

// EdgeNodeStatus represents the status of an edge node.
type EdgeNodeStatus struct {
	NodeID           string    `json:"nodeID"`
	Phase            string    `json:"phase"` // Connecting | Connected | Disconnected | Error
	LastHeartbeat    time.Time `json:"lastHeartbeat"`
	Version          string    `json:"version,omitempty"`
	AvailableGPU     int32     `json:"availableGPU,omitempty"`
	TotalGPU         int32     `json:"totalGPU,omitempty"`
	AvailableMemory  string    `json:"availableMemory,omitempty"`
	TotalMemory      string    `json:"totalMemory,omitempty"`
	AvailableCPU     string    `json:"availableCPU,omitempty"`
	TotalCPU         string    `json:"totalCPU,omitempty"`
	RunningWorkloads int32     `json:"runningWorkloads"`
	Message          string    `json:"message,omitempty"`
}

// ImageCacheConfig defines image caching configuration.
type ImageCacheConfig struct {
	Enabled         bool     `json:"enabled"`
	CacheSizeGB     int32    `json:"cacheSizeGB"`     // Maximum cache size in GB
	MaxCachedImages int32    `json:"maxCachedImages"` // Maximum number of images to cache
	TTL             string   `json:"ttl"`             // Time-to-live for cached images (e.g., "168h")
	PrepullImages   []string `json:"prepullImages"`   // Images to prepull on node startup
	RegistryMirrors []string `json:"registryMirrors"` // Registry mirror URLs
}

// ImageCacheStatus represents the status of image cache.
type ImageCacheStatus struct {
	TotalSizeGB   float64   `json:"totalSizeGB"`
	CachedImages  int32     `json:"cachedImages"`
	HitRate       float64   `json:"hitRate"` // Cache hit rate percentage
	LastCleanup   time.Time `json:"lastCleanup"`
	CacheLocation string    `json:"cacheLocation"`
}

// ResourceMetrics represents resource usage metrics for a workload.
type ResourceMetrics struct {
	WorkloadID     string         `json:"workloadId"`
	WorkloadName   string         `json:"workloadName"`
	Namespace      string         `json:"namespace"`
	Timestamp      time.Time      `json:"timestamp"`
	CPUUsage       CPUMetrics     `json:"cpuUsage"`
	MemoryUsage    MemoryMetrics  `json:"memoryUsage"`
	GPUUsage       []GPUMetrics   `json:"gpuUsage,omitempty"`
	NetworkMetrics NetworkMetrics `json:"networkMetrics,omitempty"`
	StorageMetrics StorageMetrics `json:"storageMetrics,omitempty"`
}

// CPUMetrics represents CPU usage metrics.
type CPUMetrics struct {
	UsageCores   float64 `json:"usageCores"`   // Current CPU usage in cores
	UsagePercent float64 `json:"usagePercent"` // CPU usage as percentage of limit
	RequestCores string  `json:"requestCores"` // CPU request
	LimitCores   string  `json:"limitCores"`   // CPU limit
}

// MemoryMetrics represents memory usage metrics.
type MemoryMetrics struct {
	UsageBytes   int64   `json:"usageBytes"`   // Current memory usage in bytes
	UsagePercent float64 `json:"usagePercent"` // Memory usage as percentage of limit
	RequestBytes int64   `json:"requestBytes"` // Memory request in bytes
	LimitBytes   int64   `json:"limitBytes"`   // Memory limit in bytes
	CacheBytes   int64   `json:"cacheBytes"`   // Memory cache usage
}

// GPUMetrics represents GPU usage metrics.
type GPUMetrics struct {
	DeviceID      int32   `json:"deviceId"`
	DeviceName    string  `json:"deviceName"`
	Utilization   float64 `json:"utilization"`   // GPU utilization percentage
	MemoryUsedMB  int64   `json:"memoryUsedMB"`  // GPU memory used in MB
	MemoryTotalMB int64   `json:"memoryTotalMB"` // Total GPU memory in MB
	TemperatureC  float64 `json:"temperatureC"`  // GPU temperature in Celsius
	PowerUsageW   float64 `json:"powerUsageW"`   // Power usage in watts
}

// NetworkMetrics represents network metrics.
type NetworkMetrics struct {
	ReceiveBytesPerSecond  int64 `json:"receiveBytesPerSecond"`
	TransmitBytesPerSecond int64 `json:"transmitBytesPerSecond"`
	ReceivePacketsPerSec   int64 `json:"receivePacketsPerSec"`
	TransmitPacketsPerSec  int64 `json:"transmitPacketsPerSec"`
	ErrorsPerSec           int64 `json:"errorsPerSec"`
}

// StorageMetrics represents storage metrics.
type StorageMetrics struct {
	ReadBytesPerSecond  int64 `json:"readBytesPerSecond"`
	WriteBytesPerSecond int64 `json:"writeBytesPerSecond"`
	ReadOpsPerSecond    int64 `json:"readOpsPerSecond"`
	WriteOpsPerSecond   int64 `json:"writeOpsPerSecond"`
	DiskUsageBytes      int64 `json:"diskUsageBytes"`
	DiskTotalBytes      int64 `json:"diskTotalBytes"`
}
