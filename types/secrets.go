package types

import "time"

// SecretRef identifies a secret consumed by a workload.
type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	EnvKey    string `json:"envKey,omitempty"`
}

// SecretRotationSpec enables rolling restarts when linked secrets change.
type SecretRotationSpec struct {
	Enabled    bool        `json:"enabled,omitempty"`
	SecretRefs []SecretRef `json:"secretRefs,omitempty"`
}

// SecretRotationStatus records secret rotation restart state.
type SecretRotationStatus struct {
	PendingRestart bool       `json:"pendingRestart,omitempty"`
	LastRotation   *time.Time `json:"lastRotation,omitempty"`
	SecretName     string     `json:"secretName,omitempty"`
	SecretVersion  string     `json:"secretVersion,omitempty"`
	RestartCount   int32      `json:"restartCount,omitempty"`
	Message        string     `json:"message,omitempty"`
}
