package types

// KranixAppTemplateRequest generates a full KranixApp manifest from a description.
type KranixAppTemplateRequest struct {
	Description string            `json:"description"`
	Name        string            `json:"name,omitempty"`
	Image       string            `json:"image,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Replicas    int               `json:"replicas,omitempty"`
	CPU         string            `json:"cpu,omitempty"`
	Memory      string            `json:"memory,omitempty"`
	Environment string            `json:"environment,omitempty"`
	Team        string            `json:"team,omitempty"`
	Profile     string            `json:"profile,omitempty"` // basic | production | spot | cron
	Features    []string          `json:"features,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Ports       []int             `json:"ports,omitempty"`
}

// KranixAppTemplateParsed summarizes fields inferred from the description.
type KranixAppTemplateParsed struct {
	Name       string   `json:"name"`
	Image      string   `json:"image"`
	Namespace  string   `json:"namespace"`
	Replicas   int      `json:"replicas"`
	Features   []string `json:"features,omitempty"`
	Profile    string   `json:"profile,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
}

// KranixAppTemplateResponse is the generated KranixApp YAML and metadata.
type KranixAppTemplateResponse struct {
	Manifest   string                  `json:"manifest"`
	Format     string                  `json:"format"`
	Parsed     KranixAppTemplateParsed `json:"parsed,omitempty"`
	Confidence float64                 `json:"confidence,omitempty"`
}
