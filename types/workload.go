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
	Tenant    *TenantInfo       `json:"tenant,omitempty"`
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
	Backend           string             `json:"backend"` // docker | kubernetes
	ComposeFile       string             `json:"composeFile,omitempty"`
	RolloutStrategy   RolloutStrategy    `json:"rolloutStrategy,omitempty"`
	AutoScaling       *AutoScalingConfig `json:"autoScaling,omitempty"`
	Scheduling        *SchedulingConfig  `json:"scheduling,omitempty"`
	Dependencies      []Dependency       `json:"dependencies,omitempty"`
	FailurePrediction *FailurePrediction `json:"failurePrediction,omitempty"`
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
	CostAware        bool              `json:"costAware,omitempty"`
	PreferredRegions []string          `json:"preferredRegions,omitempty"`
	PreferredZones   []string          `json:"preferredZones,omitempty"`
	NodeSelectors    map[string]string `json:"nodeSelectors,omitempty"`
	Affinity         *AffinityConfig   `json:"affinity,omitempty"`
	Tolerations      []Toleration      `json:"tolerations,omitempty"`
	MaxCostPerHour   string            `json:"maxCostPerHour,omitempty"`
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
