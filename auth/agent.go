package auth

import (
	"fmt"
	"strings"
)

// AgentScope mirrors MCP agent permission scopes.
type AgentScope string

const (
	AgentScopeReadOnly AgentScope = "readonly"
	AgentScopeWrite    AgentScope = "write"
	AgentScopeAdmin    AgentScope = "admin"
)

// AgentIdentity is the authenticated identity of an MCP agent caller.
type AgentIdentity struct {
	AgentID string     `json:"agentId"`
	Scope   AgentScope `json:"scope"`
}

// AgentCredential binds an API key to a single agent identity.
type AgentCredential struct {
	AgentID string     `json:"agentId"`
	APIKey  string     `json:"apiKey,omitempty"`
	Scope   AgentScope `json:"scope"`
}

// ImpersonationError indicates an agent attempted to act as another agent.
type ImpersonationError struct {
	AuthenticatedAgent string
	ClaimedAgent     string
}

func (e *ImpersonationError) Error() string {
	return fmt.Sprintf(
		"agent impersonation denied: authenticated as %q but attempted to act as %q",
		e.AuthenticatedAgent,
		e.ClaimedAgent,
	)
}

// ValidateAgentClaim ensures a caller cannot impersonate another agent.
// Admin agents may query other agents' audit data but cannot perform
// mutating operations on their behalf.
func ValidateAgentClaim(authenticated AgentIdentity, claimedAgentID string, allowCrossAgentRead bool) error {
	claimedAgentID = strings.TrimSpace(claimedAgentID)
	if claimedAgentID == "" {
		return nil
	}
	if authenticated.AgentID == "" {
		return fmt.Errorf("authenticated agent identity is required")
	}
	if authenticated.AgentID == claimedAgentID {
		return nil
	}
	if allowCrossAgentRead && authenticated.Scope == AgentScopeAdmin {
		return nil
	}
	return &ImpersonationError{
		AuthenticatedAgent: authenticated.AgentID,
		ClaimedAgent:     claimedAgentID,
	}
}

// ResolveAgentFromCredential finds the agent bound to an API key.
func ResolveAgentFromCredential(apiKey string, credentials []AgentCredential) (AgentIdentity, bool) {
	for _, cred := range credentials {
		if cred.APIKey != "" && cred.APIKey == apiKey {
			return AgentIdentity{
				AgentID: cred.AgentID,
				Scope:   cred.Scope,
			}, true
		}
	}
	return AgentIdentity{}, false
}
