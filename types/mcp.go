package types

import "time"

// ToolChainStep is a single step in a composed tool chain.
type ToolChainStep struct {
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Tool      string                 `json:"tool"`
	Inputs    map[string]interface{} `json:"inputs,omitempty"`
	OnFailure string                 `json:"on_failure,omitempty"` // stop | continue
}

// ToolChainRequest composes multiple MCP tools into one logical operation.
type ToolChainRequest struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Steps       []ToolChainStep `json:"steps"`
	Context     map[string]interface{} `json:"context,omitempty"`
	OnFailure   string          `json:"on_failure,omitempty"` // default for steps
}

// ToolChainStepResult is the outcome of one chained step.
type ToolChainStepResult struct {
	StepID   string        `json:"step_id,omitempty"`
	StepName string        `json:"step_name,omitempty"`
	Tool     string        `json:"tool"`
	Status   string        `json:"status"` // success | failed | skipped
	Output   string        `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// ToolChainResult is the aggregate result of a tool chain execution.
type ToolChainResult struct {
	Name      string                `json:"name,omitempty"`
	Status    string                `json:"status"` // completed | failed | partial
	Results   []ToolChainStepResult `json:"results"`
	StartedAt time.Time             `json:"started_at"`
	CompletedAt time.Time           `json:"completed_at,omitempty"`
	Error     string                `json:"error,omitempty"`
}

// ClusterHealth summarizes cluster-wide health signals.
type ClusterHealth struct {
	Status      string    `json:"status"` // healthy | degraded | critical
	NodesReady  int       `json:"nodes_ready"`
	NodesTotal  int       `json:"nodes_total"`
	PodsRunning int       `json:"pods_running"`
	PodsTotal   int       `json:"pods_total"`
	DegradedWorkloads int `json:"degraded_workloads,omitempty"`
	LastChecked time.Time `json:"last_checked"`
}

// ActionSuggestion recommends a next MCP tool action based on cluster state.
type ActionSuggestion struct {
	Tool        string                 `json:"tool"`
	Reason      string                 `json:"reason"`
	Priority    string                 `json:"priority"` // high | medium | low
	Inputs      map[string]interface{} `json:"inputs,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
}

// SuggestionsResponse bundles context-aware next-action recommendations.
type SuggestionsResponse struct {
	ClusterStatus string             `json:"cluster_status"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Suggestions   []ActionSuggestion `json:"suggestions"`
	GeneratedAt   time.Time          `json:"generated_at"`
}

// ClusterEventSubscriptionRequest configures a real-time event subscription poll.
type ClusterEventSubscriptionRequest struct {
	Namespaces     []string `json:"namespaces,omitempty"`
	EventTypes     []string `json:"eventTypes,omitempty"`
	TimeoutSeconds int      `json:"timeoutSeconds,omitempty"`
	MaxEvents      int      `json:"maxEvents,omitempty"`
	ClientID       string   `json:"clientId,omitempty"`
}

// ClusterEventItem is one SSE frame delivered to an agent.
type ClusterEventItem struct {
	ID         string      `json:"id"`
	Event      string      `json:"event"`
	Data       interface{} `json:"data,omitempty"`
	ReceivedAt time.Time   `json:"receivedAt"`
}

// ClusterEventBatch is the result of subscribe_cluster_events.
type ClusterEventBatch struct {
	Events   []ClusterEventItem `json:"events"`
	Count    int                `json:"count"`
	TimedOut bool               `json:"timedOut"`
}
