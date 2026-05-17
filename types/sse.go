package types

import "time"

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	ID        string      `json:"id"`
	Event     string      `json:"event"`     // e.g., workload.created, workload.updated, workload.deleted
	Data      interface{} `json:"data"`
	Retry     int         `json:"retry,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// WorkloadStateChange represents a workload state change event.
type WorkloadStateChange struct {
	WorkloadID   string    `json:"workloadId"`
	Namespace    string    `json:"namespace"`
	OldState     string    `json:"oldState"`
	NewState     string    `json:"newState"`
	Reason       string    `json:"reason,omitempty"`
	ChangedAt    time.Time `json:"changedAt"`
	ChangedBy    string    `json:"changedBy,omitempty"`
}

// SSESubscription represents an SSE subscription.
type SSESubscription struct {
	ID         string            `json:"id"`
	ClientID   string            `json:"clientId"`
	Namespaces []string          `json:"namespaces"`
	Events     []string          `json:"events"`     // e.g., workload.created, workload.updated
	Filter     map[string]string `json:"filter"`     // additional filters
	CreatedAt  time.Time         `json:"createdAt"`
	LastActive time.Time         `json:"lastActive"`
}

// SSEClient represents an SSE client connection.
type SSEClient struct {
	ID         string            `json:"id"`
	ClientID   string            `json:"clientId"`
	UserAgent  string            `json:"userAgent"`
	RemoteAddr string            `json:"remoteAddr"`
	ConnectedAt time.Time        `json:"connectedAt"`
	Subscriptions []*SSESubscription `json:"subscriptions"`
}

// BroadcastMessage represents a message to be broadcast to SSE clients.
type BroadcastMessage struct {
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	Filter    map[string]string `json:"filter,omitempty"`
}
