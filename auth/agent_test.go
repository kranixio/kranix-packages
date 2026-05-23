package auth

import "testing"

func TestValidateAgentClaim(t *testing.T) {
	agent := AgentIdentity{AgentID: "agent-a", Scope: AgentScopeWrite}

	if err := ValidateAgentClaim(agent, "agent-a", false); err != nil {
		t.Fatalf("expected same agent to pass: %v", err)
	}

	if err := ValidateAgentClaim(agent, "agent-b", false); err == nil {
		t.Fatal("expected impersonation to fail")
	}

	admin := AgentIdentity{AgentID: "admin", Scope: AgentScopeAdmin}
	if err := ValidateAgentClaim(admin, "agent-b", true); err != nil {
		t.Fatalf("expected admin cross-agent read: %v", err)
	}
	if err := ValidateAgentClaim(admin, "agent-b", false); err == nil {
		t.Fatal("expected admin mutating impersonation to fail without allowCrossAgentRead")
	}
}

func TestResolveAgentFromCredential(t *testing.T) {
	creds := []AgentCredential{
		{AgentID: "bot-1", APIKey: "krane_test123", Scope: AgentScopeWrite},
	}
	id, ok := ResolveAgentFromCredential("krane_test123", creds)
	if !ok || id.AgentID != "bot-1" {
		t.Fatalf("expected credential resolution, got %+v ok=%v", id, ok)
	}
}
