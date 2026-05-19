package types

import "time"

// Circuit breaker states.
const (
	CircuitStateClosed   = "closed"
	CircuitStateOpen     = "open"
	CircuitStateHalfOpen = "half-open"
)

// CircuitBreakerSpec configures automatic traffic cut-off for unhealthy workloads.
type CircuitBreakerSpec struct {
	Enabled             bool  `json:"enabled,omitempty"`
	FailureThreshold    int32 `json:"failureThreshold,omitempty"`
	SuccessThreshold    int32 `json:"successThreshold,omitempty"`
	OpenDurationSeconds int32 `json:"openDurationSeconds,omitempty"`
	HalfOpenMaxRequests int32 `json:"halfOpenMaxRequests,omitempty"`
	TripOnDegraded      *bool `json:"tripOnDegraded,omitempty"`
}

// CircuitBreakerStatus is observed circuit state persisted on the workload.
type CircuitBreakerStatus struct {
	State            string     `json:"state,omitempty"`
	ConsecutiveFails int32      `json:"consecutiveFails,omitempty"`
	ConsecutiveOK    int32      `json:"consecutiveOK,omitempty"`
	HalfOpenAttempts int32      `json:"halfOpenAttempts,omitempty"`
	LastTransition   time.Time  `json:"lastTransition,omitempty"`
	OpenUntil        *time.Time `json:"openUntil,omitempty"`
	Message          string     `json:"message,omitempty"`
}

// WarmStandbySpec keeps a cold replica workload ready for failover.
type WarmStandbySpec struct {
	Enabled           bool   `json:"enabled,omitempty"`
	Replicas          int32  `json:"replicas,omitempty"`
	AutoPromote       bool   `json:"autoPromote,omitempty"`
	StandbyWorkloadID string `json:"standbyWorkloadId,omitempty"`
}

// WarmStandbyPhase describes standby lifecycle.
type WarmStandbyPhase string

const (
	WarmStandbyPhaseCold     WarmStandbyPhase = "Cold"
	WarmStandbyPhaseWarming  WarmStandbyPhase = "Warming"
	WarmStandbyPhasePromoted WarmStandbyPhase = "Promoted"
)

// WarmStandbyStatus records linked standby workload state.
type WarmStandbyStatus struct {
	StandbyWorkloadID string           `json:"standbyWorkloadId,omitempty"`
	Phase             WarmStandbyPhase `json:"phase,omitempty"`
	ReadyReplicas     int32            `json:"readyReplicas,omitempty"`
	LastFailover      *time.Time       `json:"lastFailover,omitempty"`
	Message           string           `json:"message,omitempty"`
}
