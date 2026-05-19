package auth

import (
	"fmt"
	"net"
	"strings"
)

// ClientIP extracts the client IP from an HTTP-style remote address (host:port or IP).
func ClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return host
}

// IPAllowed reports whether clientIP matches any entry in allowed (exact IP or CIDR).
// An empty allowed list means all IPs are permitted.
func IPAllowed(clientIP string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return false
	}
	for _, entry := range allowed {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			_, network, err := net.ParseCIDR(entry)
			if err == nil && network.Contains(ip) {
				return true
			}
			continue
		}
		if allowedIP := net.ParseIP(entry); allowedIP != nil && allowedIP.Equal(ip) {
			return true
		}
	}
	return false
}

// ValidateAllowedIPs returns an error when any entry is not a valid IP or CIDR.
func ValidateAllowedIPs(allowed []string) error {
	for _, entry := range allowed {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return fmt.Errorf("invalid CIDR %q", entry)
			}
			continue
		}
		if net.ParseIP(entry) == nil {
			return fmt.Errorf("invalid IP %q", entry)
		}
	}
	return nil
}
