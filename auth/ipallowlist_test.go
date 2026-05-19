package auth

import "testing"

func TestIPAllowed(t *testing.T) {
	if !IPAllowed("203.0.113.10", nil) {
		t.Fatal("empty allowlist should permit any IP")
	}
	if !IPAllowed("203.0.113.10", []string{"203.0.113.10"}) {
		t.Fatal("exact match should be allowed")
	}
	if IPAllowed("203.0.113.11", []string{"203.0.113.10"}) {
		t.Fatal("non-matching IP should be denied")
	}
	if !IPAllowed("10.0.0.5", []string{"10.0.0.0/8"}) {
		t.Fatal("CIDR match should be allowed")
	}
}

func TestValidateAllowedIPs(t *testing.T) {
	if err := ValidateAllowedIPs([]string{"not-an-ip"}); err == nil {
		t.Fatal("expected validation error")
	}
	if err := ValidateAllowedIPs([]string{"10.0.0.0/8"}); err != nil {
		t.Fatalf("valid CIDR: %v", err)
	}
}
