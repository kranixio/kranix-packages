//! Common types for the Kranix SDK

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Represents a managed workload in the Kranix system
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Workload {
    pub id: String,
    pub name: String,
    pub namespace: String,
    pub spec: WorkloadSpec,
    pub status: WorkloadStatus,
    #[serde(with = "chrono::serde::ts_seconds")]
    pub created_at: DateTime<Utc>,
    #[serde(with = "chrono::serde::ts_seconds")]
    pub updated_at: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub labels: Option<std::collections::HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tenant: Option<TenantInfo>,
}

/// Desired configuration of a workload
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkloadSpec {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub namespace: Option<String>,
    pub image: String,
    pub replicas: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub env: Option<std::collections::HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub command: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub resources: Option<ResourceSpec>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub ports: Vec<PortSpec>,
    pub backend: String, // docker | kubernetes | podman | compose | remote
    #[serde(skip_serializing_if = "Option::is_none")]
    pub compose_file: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub remote_host: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub rollout_strategy: Option<RolloutStrategy>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub auto_scaling: Option<AutoScalingConfig>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub scheduling: Option<SchedulingConfig>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub dependencies: Vec<Dependency>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub failure_prediction: Option<FailurePrediction>,
    #[serde(skip_serializing_if = "Option::is_none", rename = "crossNamespaceTraffic")]
    pub cross_namespace_traffic: Option<CrossNamespaceTrafficPolicy>,
}

/// Compute resource requirements
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResourceSpec {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cpu_request: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cpu_limit: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub memory_request: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub memory_limit: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub gpu: Option<GPUSpec>,
}

/// GPU resource requirements
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GPUSpec {
    pub vendor: String, // nvidia | amd
    pub count: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub r#type: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub sku: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub memory: Option<String>,
}

/// Port configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PortSpec {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    pub container_port: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub protocol: Option<String>, // tcp | udp
}

/// Current observed state of a workload
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkloadStatus {
    pub id: String,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub namespace: Option<String>,
    pub state: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub image: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub replicas: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ready: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub host: Option<String>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub pods: Vec<String>,
    pub phase: WorkloadPhase,
    pub ready_replicas: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub message: Option<String>,
    #[serde(with = "chrono::serde::ts_seconds")]
    pub last_updated: DateTime<Utc>,
}

/// Lifecycle phase of a workload
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "SCREAMING_SNAKE_CASE")]
pub enum WorkloadPhase {
    Pending,
    Deploying,
    Running,
    Degraded,
    Failed,
}

/// Rollout strategy for deployment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RolloutStrategy {
    #[serde(rename = "type")]
    pub strategy_type: String, // rolling, recreate, bluegreen, canary, abtest
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_unavailable: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_surge: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub canary_config: Option<CanaryConfig>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ab_test_config: Option<ABTestConfig>,
}

/// Auto-scaling configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AutoScalingConfig {
    pub enabled: bool,
    pub min_replicas: i32,
    pub max_replicas: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub target_cpu_utilization: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub target_memory_utilization: Option<i32>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub custom_metrics: Vec<CustomMetric>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub scale_down_cooldown_seconds: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub scale_up_cooldown_seconds: Option<i32>,
}

/// Custom metric for auto-scaling
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CustomMetric {
    pub name: String,
    pub metric_type: String, // pods, object
    pub metric_name: String,
    pub target: MetricTarget,
}

/// Target value for a metric
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetricTarget {
    pub target_type: String, // average, value
    #[serde(skip_serializing_if = "Option::is_none")]
    pub average_value: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub value: Option<String>,
}

/// Scheduling preferences
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SchedulingConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cost_aware: Option<bool>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub preferred_regions: Vec<String>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub preferred_zones: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub node_selectors: Option<std::collections::HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub affinity: Option<AffinityConfig>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub tolerations: Vec<Toleration>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_cost_per_hour: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub workload_priority: Option<String>,
    #[serde(skip_serializing_if = "std::ops::Not::not")]
    pub preemption_enabled: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub priority_class_name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub spot: Option<SpotWorkloadConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SpotWorkloadConfig {
    #[serde(default)]
    pub enabled: bool,
    #[serde(default)]
    pub reschedule_on_node_termination: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CrossNamespaceTrafficPolicy {
    #[serde(default)]
    pub enabled: bool,
    #[serde(default)]
    pub allowed_ingress_namespaces: Vec<String>,
    #[serde(default)]
    pub allowed_egress_namespaces: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub allow_same_namespace: Option<bool>,
    #[serde(default, rename = "blockClusterDNS")]
    pub block_cluster_dns: bool,
    #[serde(default)]
    pub allow_egress_internet: bool,
}

/// Affinity configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AffinityConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub node_affinity: Option<NodeAffinity>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub pod_affinity: Option<PodAffinity>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub pod_anti_affinity: Option<PodAffinity>,
}

/// Node affinity rules
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeAffinity {
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub required_during_scheduling: Vec<NodeSelectorTerm>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub preferred_during_scheduling: Vec<PreferredSchedulingTerm>,
}

/// Node selector term
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeSelectorTerm {
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub match_expressions: Vec<NodeSelectorRequirement>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub match_fields: Vec<NodeSelectorRequirement>,
}

/// Node selector requirement
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeSelectorRequirement {
    pub key: String,
    pub operator: String, // In, NotIn, Exists, DoesNotExist, Gt, Lt
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub values: Vec<String>,
}

/// Preferred scheduling term
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PreferredSchedulingTerm {
    pub weight: i32,
    pub preference: NodeSelectorTerm,
}

/// Pod affinity rules
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PodAffinity {
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub required_during_scheduling: Vec<PodAffinityTerm>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub preferred_during_scheduling: Vec<WeightedPodAffinityTerm>,
}

/// Pod affinity term
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PodAffinityTerm {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub label_selector: Option<std::collections::HashMap<String, String>>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub namespaces: Vec<String>,
    pub topology_key: String,
}

/// Weighted pod affinity term
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WeightedPodAffinityTerm {
    pub weight: i32,
    pub pod_affinity_term: PodAffinityTerm,
}

/// Toleration for taints
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Toleration {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub operator: Option<String>, // Exists, Equal
    #[serde(skip_serializing_if = "Option::is_none")]
    pub value: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub effect: Option<String>, // NoSchedule, PreferNoSchedule, NoExecute
    #[serde(skip_serializing_if = "Option::is_none")]
    pub toleration_seconds: Option<i64>,
}

/// Canary deployment configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CanaryConfig {
    pub replicas: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub percentage: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub analysis_duration: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub success_threshold: Option<i32>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub metrics: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub auto_promote: Option<bool>,
}

/// A/B testing configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ABTestConfig {
    pub variant_a: String,
    pub variant_b: String,
    pub traffic_split: i32, // percentage for variant B
    #[serde(skip_serializing_if = "Option::is_none")]
    pub analysis_duration: Option<String>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub metrics: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub auto_select_winner: Option<bool>,
}

/// Dependency relationship between workloads
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Dependency {
    pub workload_id: String,
    pub dependency_type: String, // depends_on, waits_for, requires
    #[serde(skip_serializing_if = "Option::is_none")]
    pub condition: Option<String>, // running, healthy, ready
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timeout: Option<String>, // duration string like "5m"
}

/// Failure prediction configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FailurePrediction {
    pub enabled: bool,
    pub model_type: String, // simple, ml, custom
    #[serde(skip_serializing_if = "Option::is_none")]
    pub prediction_window: Option<String>,
    pub threshold: f64, // probability threshold (0-1)
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub features: Vec<String>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub mitigation_actions: Vec<String>,
}

/// Multi-tenancy information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TenantInfo {
    pub id: String,
    pub name: String,
    pub namespace: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub labels: Option<std::collections::HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub quota: Option<TenantQuota>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub isolation: Option<TenantIsolation>,
}

/// Resource quotas for a tenant
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TenantQuota {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_cpu: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_memory: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_workloads: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_replicas: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_storage: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_custom_metrics: Option<i32>,
}

/// Isolation mechanisms for a tenant
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TenantIsolation {
    pub network_policy: bool,
    pub resource_quota: bool,
    pub limit_range: bool,
    pub pod_security_policy: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub storage_class: Option<String>,
}

/// Namespace in the Kranix system
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Namespace {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub labels: Option<std::collections::HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub quota: Option<NamespaceQuota>,
    #[serde(with = "chrono::serde::ts_seconds")]
    pub created_at: DateTime<Utc>,
    #[serde(with = "chrono::serde::ts_seconds")]
    pub updated_at: DateTime<Utc>,
}

/// Resource quotas for a namespace
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NamespaceQuota {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_workloads: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_cpu: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_memory: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_storage: Option<String>,
}

/// Pod in the system
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Pod {
    pub id: String,
    pub name: String,
    pub namespace: String,
    pub workload_id: String,
    pub phase: String, // Pending, Running, Succeeded, Failed, Unknown
    #[serde(skip_serializing_if = "Option::is_none")]
    pub node: Option<String>,
    pub created_at: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub started_at: Option<DateTime<Utc>>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub container_statuses: Vec<ContainerStatus>,
}

/// Status of a container
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ContainerStatus {
    pub name: String,
    pub image: String,
    pub ready: bool,
    pub restart_count: i32,
    pub state: String, // running, waiting, terminated
}

/// SSE event for workload changes
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkloadEvent {
    pub event_type: String, // workload.created, workload.updated, workload.deleted
    pub workload: Workload,
    pub timestamp: DateTime<Utc>,
}
