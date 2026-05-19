package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// TokenType represents the type of authentication token.
type TokenType string

const (
	TokenTypeAPIKey TokenType = "api_key"
	TokenTypeJWT    TokenType = "jwt"
)

// Token represents an authentication token.
type Token struct {
	Type      TokenType `json:"type"`
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	IssuedAt  time.Time `json:"issuedAt"`
}

// Claims represents JWT claims.
type Claims struct {
	UserID    string   `json:"userId"`
	Username  string   `json:"username"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
}

// ValidateAPIKey validates an API key format.
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("api key cannot be empty")
	}

	if !strings.HasPrefix(apiKey, "krane_") {
		return fmt.Errorf("api key must start with 'krane_'")
	}

	if len(apiKey) < 16 {
		return fmt.Errorf("api key must be at least 16 characters")
	}

	return nil
}

// GenerateAPIKey generates a new API key.
func GenerateAPIKey() string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 16)
	// In production, use crypto/rand
	for i := range randomBytes {
		randomBytes[i] = byte(timestamp % 256)
	}

	encoded := base64.URLEncoding.EncodeToString(randomBytes)
	return fmt.Sprintf("krane_%s", encoded)
}

// HashSecret hashes a secret using HMAC-SHA256.
func HashSecret(secret, salt string) string {
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(secret))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ValidateToken checks if a token is valid based on its type and expiration.
func ValidateToken(token *Token) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	if token.Value == "" {
		return fmt.Errorf("token value cannot be empty")
	}

	if !token.ExpiresAt.IsZero() && time.Now().After(token.ExpiresAt) {
		return fmt.Errorf("token has expired")
	}

	switch token.Type {
	case TokenTypeAPIKey:
		return ValidateAPIKey(token.Value)
	case TokenTypeJWT:
		// JWT validation would be implemented here
		return nil
	default:
		return fmt.Errorf("unknown token type: %s", token.Type)
	}
}

// HasRole checks if a claims object has a specific role.
func HasRole(claims *Claims, role string) bool {
	for _, r := range claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if a claims object has any of the specified roles.
func HasAnyRole(claims *Claims, roles ...string) bool {
	for _, role := range roles {
		if HasRole(claims, role) {
			return true
		}
	}
	return false
}

// Scope represents an API key scope for fine-grained access control.
type Scope string

const (
	ScopeRead  Scope = "read"
	ScopeWrite Scope = "write"
	ScopeAdmin Scope = "admin"
)

// ResourceType represents the type of resource a scope applies to.
type ResourceType string

const (
	ResourceWorkload  ResourceType = "workload"
	ResourceNamespace ResourceType = "namespace"
	ResourcePod       ResourceType = "pod"
	ResourceEvent     ResourceType = "event"
	ResourceWebhook   ResourceType = "webhook"
	ResourceTenant    ResourceType = "tenant"
	ResourceQuota     ResourceType = "quota"
)

// Permission represents a fine-grained permission.
type Permission struct {
	ResourceType ResourceType `json:"resourceType"`
	ResourceID   string       `json:"resourceId,omitempty"` // Empty means all resources of this type
	Scope        Scope        `json:"scope"`
}

// APIKey represents an API key with fine-grained scopes.
type APIKey struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Key         string       `json:"key"`
	Permissions []Permission `json:"permissions"`
	ExpiresAt   time.Time    `json:"expiresAt,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	CreatedBy   string       `json:"createdBy,omitempty"`
	TenantID    string       `json:"tenantId,omitempty"`
	Revoked     bool         `json:"revoked"`
	// AllowedIPs restricts which client IPs may use this key (empty = any).
	AllowedIPs []string `json:"allowedIps,omitempty"`
}

// HasPermission checks if an API key has a specific permission.
func HasPermission(apiKey *APIKey, resourceType ResourceType, scope Scope, resourceID string) bool {
	if apiKey == nil || apiKey.Revoked {
		return false
	}

	for _, perm := range apiKey.Permissions {
		if perm.ResourceType != resourceType {
			continue
		}

		// If resourceID is specified, check if it matches or is wildcard
		if perm.ResourceID != "" && perm.ResourceID != resourceID {
			continue
		}

		// Check scope hierarchy: admin > write > read
		switch perm.Scope {
		case ScopeAdmin:
			return true
		case ScopeWrite:
			if scope == ScopeWrite || scope == ScopeRead {
				return true
			}
		case ScopeRead:
			if scope == ScopeRead {
				return true
			}
		}
	}

	return false
}

// GenerateScopedAPIKey generates a new API key with specific permissions.
func GenerateScopedAPIKey(name string, permissions []Permission) *APIKey {
	return &APIKey{
		ID:          generateID(),
		Name:        name,
		Key:         GenerateAPIKey(),
		Permissions: permissions,
		CreatedAt:   time.Now(),
		Revoked:     false,
	}
}

// generateID generates a unique ID.
func generateID() string {
	return fmt.Sprintf("key_%d", time.Now().UnixNano())
}
