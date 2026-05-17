package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// OIDCProvider represents an OIDC provider configuration.
type OIDCProvider struct {
	Name         string   `json:"name"` // google, github, okta
	DisplayName  string   `json:"displayName"`
	Issuer       string   `json:"issuer"`
	AuthURL      string   `json:"authUrl"`
	TokenURL     string   `json:"tokenUrl"`
	UserInfoURL  string   `json:"userInfoUrl"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	Scopes       []string `json:"scopes"`
	Enabled      bool     `json:"enabled"`
}

// OIDCConfig represents OIDC configuration.
type OIDCConfig struct {
	Providers     map[string]*OIDCProvider `json:"providers"`
	CallbackURL   string                   `json:"callbackUrl"`
	SessionSecret string                   `json:"sessionSecret"`
	SessionTTL    time.Duration            `json:"sessionTTL"`
}

// OIDCToken represents an OIDC token response.
type OIDCToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// OIDCUserInfo represents user information from OIDC provider.
type OIDCUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
	Provider      string `json:"provider"`
}

// Session represents an OIDC session.
type Session struct {
	ID           string        `json:"id"`
	UserInfo     *OIDCUserInfo `json:"userInfo"`
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken,omitempty"`
	ExpiresAt    time.Time     `json:"expiresAt"`
	CreatedAt    time.Time     `json:"createdAt"`
}

// OIDCManager manages OIDC authentication.
type OIDCManager struct {
	config   *OIDCConfig
	sessions map[string]*Session
}

// NewOIDCManager creates a new OIDC manager.
func NewOIDCManager(config *OIDCConfig) *OIDCManager {
	return &OIDCManager{
		config:   config,
		sessions: make(map[string]*Session),
	}
}

// GetAuthURL generates the authorization URL for a provider.
func (m *OIDCManager) GetAuthURL(providerName, state string) (string, error) {
	provider, ok := m.config.Providers[providerName]
	if !ok || !provider.Enabled {
		return "", fmt.Errorf("provider not found or disabled")
	}

	params := url.Values{}
	params.Set("client_id", provider.ClientID)
	params.Set("redirect_uri", m.config.CallbackURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(provider.Scopes, " "))
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", provider.AuthURL, params.Encode()), nil
}

// ExchangeCodeForToken exchanges an authorization code for tokens.
func (m *OIDCManager) ExchangeCodeForToken(providerName, code string) (*OIDCToken, error) {
	_, ok := m.config.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}

	// In production, make actual HTTP request to token endpoint
	// For now, return mock token
	return &OIDCToken{
		AccessToken:  generateRandomToken(),
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: generateRandomToken(),
		IDToken:      generateRandomToken(),
	}, nil
}

// GetUserInfo retrieves user information from the provider.
func (m *OIDCManager) GetUserInfo(providerName, accessToken string) (*OIDCUserInfo, error) {
	_, ok := m.config.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}

	// In production, make actual HTTP request to userinfo endpoint
	// For now, return mock user info
	return &OIDCUserInfo{
		Sub:           generateRandomToken(),
		Name:          "Test User",
		Email:         "test@example.com",
		EmailVerified: true,
		Picture:       "https://example.com/avatar.png",
		Provider:      providerName,
	}, nil
}

// CreateSession creates a new session.
func (m *OIDCManager) CreateSession(userInfo *OIDCUserInfo, token *OIDCToken) (*Session, error) {
	sessionID := generateRandomToken()
	session := &Session{
		ID:           sessionID,
		UserInfo:     userInfo,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(m.config.SessionTTL),
		CreatedAt:    time.Now(),
	}
	m.sessions[sessionID] = session
	return session, nil
}

// GetSession retrieves a session by ID.
func (m *OIDCManager) GetSession(sessionID string) (*Session, error) {
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		delete(m.sessions, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// DeleteSession deletes a session.
func (m *OIDCManager) DeleteSession(sessionID string) {
	delete(m.sessions, sessionID)
}

// GetConfig returns the OIDC configuration.
func (m *OIDCManager) GetConfig() *OIDCConfig {
	return m.config
}

// generateRandomToken generates a random token.
func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ValidateIDToken validates an ID token (JWT).
func ValidateIDToken(idToken string, issuer string) error {
	// In production, validate JWT signature and claims
	// For now, just check if it's not empty
	if idToken == "" {
		return fmt.Errorf("id token is empty")
	}
	return nil
}

// GetGoogleProvider returns a Google OIDC provider configuration.
func GetGoogleProvider(clientID, clientSecret, callbackURL string) *OIDCProvider {
	return &OIDCProvider{
		Name:         "google",
		DisplayName:  "Google",
		Issuer:       "https://accounts.google.com",
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		Enabled:      true,
	}
}

// GetGitHubProvider returns a GitHub OAuth provider configuration.
func GetGitHubProvider(clientID, clientSecret, callbackURL string) *OIDCProvider {
	return &OIDCProvider{
		Name:         "github",
		DisplayName:  "GitHub",
		Issuer:       "https://github.com",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"read:user", "user:email"},
		Enabled:      true,
	}
}

// GetOktaProvider returns an Okta OIDC provider configuration.
func GetOktaProvider(domain, clientID, clientSecret, callbackURL string) *OIDCProvider {
	return &OIDCProvider{
		Name:         "okta",
		DisplayName:  "Okta",
		Issuer:       fmt.Sprintf("https://%s", domain),
		AuthURL:      fmt.Sprintf("https://%s/oauth2/v1/authorize", domain),
		TokenURL:     fmt.Sprintf("https://%s/oauth2/v1/token", domain),
		UserInfoURL:  fmt.Sprintf("https://%s/oauth2/v1/userinfo", domain),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		Enabled:      true,
	}
}
