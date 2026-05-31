package types

// WorkloadProbes configures container health checks for a workload.
// Startup probes block liveness and readiness until the app is initialized.
type WorkloadProbes struct {
	Startup   *ProbeSpec `json:"startup,omitempty"`
	Liveness  *ProbeSpec `json:"liveness,omitempty"`
	Readiness *ProbeSpec `json:"readiness,omitempty"`
}

// ProbeSpec defines an HTTP, TCP, or exec health probe.
type ProbeSpec struct {
	Type                string   `json:"type,omitempty"` // http | tcp | exec
	Path                string   `json:"path,omitempty"`
	Port                int32    `json:"port,omitempty"`
	Command             []string `json:"command,omitempty"`
	Scheme              string   `json:"scheme,omitempty"` // HTTP | HTTPS
	InitialDelaySeconds int32    `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       int32    `json:"periodSeconds,omitempty"`
	TimeoutSeconds      int32    `json:"timeoutSeconds,omitempty"`
	FailureThreshold    int32    `json:"failureThreshold,omitempty"`
	SuccessThreshold    int32    `json:"successThreshold,omitempty"`
}
