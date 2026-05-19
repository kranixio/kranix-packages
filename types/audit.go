package types

import "time"

// AuditEntry records an API or control-plane action on a resource.
type AuditEntry struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Actor        string                 `json:"actor,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId"`
	Outcome      string                 `json:"outcome"` // success | error
	Error        string                 `json:"error,omitempty"`
	RequestID    string                 `json:"requestId,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// AuditQuery filters audit log entries.
type AuditQuery struct {
	ResourceType string    `json:"resourceType,omitempty"`
	ResourceID   string    `json:"resourceId,omitempty"`
	Action       string    `json:"action,omitempty"`
	Actor        string    `json:"actor,omitempty"`
	Since        time.Time `json:"since,omitempty"`
	Until        time.Time `json:"until,omitempty"`
	Limit        int       `json:"limit,omitempty"`
}

// AuditListResponse returns audit entries for a resource.
type AuditListResponse struct {
	ResourceType string       `json:"resourceType,omitempty"`
	ResourceID   string       `json:"resourceId,omitempty"`
	Entries      []AuditEntry `json:"entries"`
	CoreEvents   interface{}  `json:"coreEvents,omitempty"`
}
