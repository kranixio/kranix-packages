package types

import "time"

// APIVersion represents the API version information.
type APIVersion struct {
	Version         string           `json:"version"`
	Major           int              `json:"major"`
	Minor           int              `json:"minor"`
	Patch           int              `json:"patch"`
	Prerelease      string           `json:"prerelease,omitempty"`
	BuildMetadata   string           `json:"buildMetadata,omitempty"`
	ReleasedAt      time.Time        `json:"releasedAt"`
	Deprecated      bool             `json:"deprecated"`
	DeprecationInfo *DeprecationInfo `json:"deprecationInfo,omitempty"`
	Supported       bool             `json:"supported"`
	SupportedUntil  time.Time        `json:"supportedUntil,omitempty"`
}

// DeprecationInfo contains deprecation details.
type DeprecationInfo struct {
	Message        string    `json:"message"`
	Since          string    `json:"since"`       // version when deprecated
	SunsetDate     time.Time `json:"sunsetDate"`  // when it will be removed
	Replacement    string    `json:"replacement"` // alternative version or endpoint
	MigrationGuide string    `json:"migrationGuide,omitempty"`
}

// ChangelogEntry represents a changelog entry.
type ChangelogEntry struct {
	ID          string         `json:"id"`
	Version     string         `json:"version"`
	ReleasedAt  time.Time      `json:"releasedAt"`
	Type        ChangeType     `json:"type"`     // added, changed, deprecated, removed, fixed, security
	Category    string         `json:"category"` // api, feature, bugfix, security, performance
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Affects     []string       `json:"affects,omitempty"` // endpoints, features affected
	Breaking    bool           `json:"breaking"`
	Migration   *MigrationInfo `json:"migration,omitempty"`
	Author      string         `json:"author,omitempty"`
}

// ChangeType represents the type of change.
type ChangeType string

const (
	ChangeTypeAdded      ChangeType = "added"
	ChangeTypeChanged    ChangeType = "changed"
	ChangeTypeDeprecated ChangeType = "deprecated"
	ChangeTypeRemoved    ChangeType = "removed"
	ChangeTypeFixed      ChangeType = "fixed"
	ChangeTypeSecurity   ChangeType = "security"
)

// MigrationInfo contains migration guidance.
type MigrationInfo struct {
	Required        bool     `json:"required"`
	GuideURL        string   `json:"guideUrl,omitempty"`
	Steps           []string `json:"steps,omitempty"`
	BreakingChanges []string `json:"breakingChanges,omitempty"`
}

// APIResponseMetadata contains response metadata including version info.
type APIResponseMetadata struct {
	APIVersion        string           `json:"apiVersion"`
	RequestID         string           `json:"requestId"`
	Timestamp         time.Time        `json:"timestamp"`
	DeprecationNotice *DeprecationInfo `json:"deprecationNotice,omitempty"`
}
