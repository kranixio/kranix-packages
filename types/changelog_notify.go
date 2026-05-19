package types

import "time"

// ChangelogSubscription registers webhook or email alerts for API releases.
type ChangelogSubscription struct {
	ID           string    `json:"id"`
	Name         string    `json:"name,omitempty"`
	WebhookURL   string    `json:"webhookUrl,omitempty"`
	Email        string    `json:"email,omitempty"`
	BreakingOnly bool      `json:"breakingOnly,omitempty"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
}

// PublishChangelogReleaseRequest publishes a version and optionally notifies subscribers.
type PublishChangelogReleaseRequest struct {
	Version string           `json:"version"`
	Entries []ChangelogEntry `json:"entries"`
	Notify  bool             `json:"notify,omitempty"`
}

// ChangelogNotificationPayload is delivered to webhooks on breaking releases.
type ChangelogNotificationPayload struct {
	Version         string           `json:"version"`
	ReleasedAt      time.Time        `json:"releasedAt"`
	BreakingChanges []ChangelogEntry `json:"breakingChanges"`
	AllChanges      []ChangelogEntry `json:"allChanges,omitempty"`
	Message         string           `json:"message"`
}

// ChangelogNotifyResult summarizes notification delivery.
type ChangelogNotifyResult struct {
	Version          string   `json:"version"`
	Subscribers      int      `json:"subscribers"`
	WebhooksSent     int      `json:"webhooksSent"`
	EmailsSent       int      `json:"emailsSent"`
	Errors           []string `json:"errors,omitempty"`
}
