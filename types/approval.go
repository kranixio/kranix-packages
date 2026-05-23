package types

import "time"

// ApprovalStatus is the lifecycle state of a human approval gate.
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalDenied   ApprovalStatus = "denied"
	ApprovalExpired  ApprovalStatus = "expired"
)

// ApprovalCreateRequest requests human confirmation before a destructive MCP action.
type ApprovalCreateRequest struct {
	Tool       string                 `json:"tool"`
	Action     string                 `json:"action,omitempty"`
	Resource   string                 `json:"resource,omitempty"`
	Namespace  string                 `json:"namespace,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
	Inputs     map[string]interface{} `json:"inputs"`
	AgentID    string                 `json:"agentId"`
	TTLSeconds int                    `json:"ttlSeconds,omitempty"`
}

// ApprovalRecord is a pending or resolved approval gate.
type ApprovalRecord struct {
	ID         string                 `json:"id"`
	Status     ApprovalStatus         `json:"status"`
	Tool       string                 `json:"tool"`
	Action     string                 `json:"action,omitempty"`
	Resource   string                 `json:"resource,omitempty"`
	Namespace  string                 `json:"namespace,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
	Inputs     map[string]interface{} `json:"inputs,omitempty"`
	AgentID    string                 `json:"agentId"`
	Preview    string                 `json:"preview,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
	ExpiresAt  time.Time              `json:"expiresAt"`
	ResolvedAt *time.Time             `json:"resolvedAt,omitempty"`
	ResolvedBy string                 `json:"resolvedBy,omitempty"`
	Comment    string                 `json:"comment,omitempty"`
}

// ApprovalResolveRequest approves or denies a pending gate.
type ApprovalResolveRequest struct {
	Approved   bool   `json:"approved"`
	ResolvedBy string `json:"resolvedBy,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

// ApprovalListResponse lists pending approvals.
type ApprovalListResponse struct {
	Approvals []ApprovalRecord `json:"approvals"`
	Count     int              `json:"count"`
}
