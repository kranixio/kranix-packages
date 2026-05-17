package types

import "time"

// APIRouteVersion represents an API version.
type APIRouteVersion struct {
	Version     string    `json:"version"` // e.g., v1, v2
	DisplayName string    `json:"displayName"`
	Status      string    `json:"status"` // stable, beta, alpha, deprecated
	Deprecated  bool      `json:"deprecated"`
	SunsetDate  time.Time `json:"sunsetDate,omitempty"`
	ReleaseDate time.Time `json:"releaseDate"`
	BasePath    string    `json:"basePath"` // e.g., /api/v1, /api/v2
}

// APIEndpoint represents an API endpoint with version information.
type APIEndpoint struct {
	Method         string      `json:"method"`   // GET, POST, PUT, DELETE
	Path           string      `json:"path"`     // /workloads
	Versions       []string    `json:"versions"` // v1, v2
	DefaultVersion string      `json:"defaultVersion"`
	Parameters     []Parameter `json:"parameters,omitempty"`
}

// Parameter represents an API parameter.
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // path, query, header, body
	Required    bool   `json:"required"`
	Type        string `json:"type"` // string, integer, boolean
	Description string `json:"description,omitempty"`
}

// APIVersionMapping represents version-specific endpoint mappings.
type APIVersionMapping struct {
	V1Endpoints        []APIEndpoint       `json:"v1Endpoints"`
	V2Endpoints        []APIEndpoint       `json:"v2Endpoints"`
	CompatibilityRules []CompatibilityRule `json:"compatibilityRules"`
}

// CompatibilityRule represents a compatibility rule between versions.
type CompatibilityRule struct {
	FromVersion string `json:"fromVersion"`
	ToVersion   string `json:"toVersion"`
	RuleType    string `json:"ruleType"` // field_rename, type_change, breaking_change
	Description string `json:"description"`
	FieldName   string `json:"fieldName,omitempty"`
	OldField    string `json:"oldField,omitempty"`
	NewField    string `json:"newField,omitempty"`
}

// APIVersioningConfig represents API versioning configuration.
type APIVersionConfig struct {
	DefaultVersion   string            `json:"defaultVersion"` // v1
	Versions         []APIRouteVersion `json:"versions"`
	EnableVersioning bool              `json:"enableVersioning"`
	HeaderName       string            `json:"headerName"` // X-API-Version
	QueryParam       string            `json:"queryParam"` // version
}

// VersionMigration represents a migration from one API version to another.
type VersionMigration struct {
	FromVersion   string `json:"fromVersion"`
	ToVersion     string `json:"toVersion"`
	TransformFunc string `json:"transformFunc"` // function name to apply transformation
}
