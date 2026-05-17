package types

import "time"

// Webhook represents a webhook configuration.
type Webhook struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Secret      string            `json:"secret,omitempty"`
	Events      []WebhookEvent    `json:"events"`
	Enabled     bool              `json:"enabled"`
	Headers     map[string]string `json:"headers,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	CreatedBy   string            `json:"createdBy,omitempty"`
	TenantID    string            `json:"tenantId,omitempty"`
}

// WebhookEvent represents the type of event that triggers a webhook.
type WebhookEvent string

const (
	WebhookEventDeployStart      WebhookEvent = "deploy.start"
	WebhookEventDeploySuccess    WebhookEvent = "deploy.success"
	WebhookEventDeployFailure    WebhookEvent = "deploy.failure"
	WebhookEventWorkloadCreated  WebhookEvent = "workload.created"
	WebhookEventWorkloadUpdated  WebhookEvent = "workload.updated"
	WebhookEventWorkloadDeleted  WebhookEvent = "workload.deleted"
	WebhookEventScaleUp          WebhookEvent = "scale.up"
	WebhookEventScaleDown        WebhookEvent = "scale.down"
	WebhookEventHealthCheckFail  WebhookEvent = "health.check.fail"
	WebhookEventDriftDetected    WebhookEvent = "drift.detected"
	WebhookEventRollbackStart    WebhookEvent = "rollback.start"
	WebhookEventRollbackComplete WebhookEvent = "rollback.complete"
)

// WebhookDelivery represents a webhook delivery attempt.
type WebhookDelivery struct {
	ID          string          `json:"id"`
	WebhookID   string          `json:"webhookId"`
	EventType   WebhookEvent    `json:"eventType"`
	Payload     WebhookPayload  `json:"payload"`
	StatusCode  int             `json:"statusCode"`
	Response    string          `json:"response,omitempty"`
	Success     bool            `json:"success"`
	Attempt     int             `json:"attempt"`
	MaxAttempts int             `json:"maxAttempts"`
	DeliveredAt time.Time       `json:"deliveredAt"`
	NextRetryAt time.Time       `json:"nextRetryAt,omitempty"`
}

// WebhookPayload represents the payload sent to a webhook.
type WebhookPayload struct {
	Event      WebhookEvent    `json:"event"`
	Timestamp  time.Time       `json:"timestamp"`
	WorkloadID string          `json:"workloadId,omitempty"`
	Namespace  string          `json:"namespace,omitempty"`
	Data       map[string]any  `json:"data,omitempty"`
}

// WebhookProvider represents a webhook integration provider.
type WebhookProvider string

const (
	WebhookProviderCustom   WebhookProvider = "custom"
	WebhookProviderSlack    WebhookProvider = "slack"
	WebhookProviderPagerDuty WebhookProvider = "pagerduty"
	WebhookProviderDiscord  WebhookProvider = "discord"
	WebhookProviderTeams    WebhookProvider = "teams"
	WebhookProviderCI       WebhookProvider = "ci"
)

// WebhookConfig represents provider-specific webhook configuration.
type WebhookConfig struct {
	Provider WebhookProvider `json:"provider"`
	Slack    *SlackConfig    `json:"slack,omitempty"`
	PagerDuty *PagerDutyConfig `json:"pagerduty,omitempty"`
	Discord  *DiscordConfig  `json:"discord,omitempty"`
	Teams    *TeamsConfig    `json:"teams,omitempty"`
	CI       *CIConfig       `json:"ci,omitempty"`
}

// SlackConfig represents Slack webhook configuration.
type SlackConfig struct {
	Channel  string `json:"channel,omitempty"`
	Username string `json:"username,omitempty"`
	IconEmoji string `json:"iconEmoji,omitempty"`
}

// PagerDutyConfig represents PagerDuty webhook configuration.
type PagerDutyConfig struct {
	RoutingKey string `json:"routingKey,omitempty"`
	Severity   string `json:"severity,omitempty"` // critical, error, warning, info
}

// DiscordConfig represents Discord webhook configuration.
type DiscordConfig struct {
	Username string `json:"username,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

// TeamsConfig represents Microsoft Teams webhook configuration.
type TeamsConfig struct {
	Color string `json:"color,omitempty"`
}

// CIConfig represents CI/CD webhook configuration.
type CIConfig struct {
	PipelineID   string `json:"pipelineId,omitempty"`
	Stage        string `json:"stage,omitempty"`
	TriggerBuild bool   `json:"triggerBuild,omitempty"`
}
