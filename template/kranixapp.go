package template

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kranix-io/kranix-packages/types"
	"gopkg.in/yaml.v3"
)

const (
	formatKranixAppYAML = "kranixapp-yaml"
	apiVersion          = "kranix.io/v1alpha1"
)

// GenerateKranixApp builds a full KranixApp manifest from a request.
func GenerateKranixApp(req types.KranixAppTemplateRequest) (*types.KranixAppTemplateResponse, error) {
	if strings.TrimSpace(req.Description) == "" && req.Name == "" {
		return nil, fmt.Errorf("description or name is required")
	}

	parsed := parseDescription(req.Description)
	mergeRequest(&parsed, req)

	if parsed.Name == "" {
		parsed.Name = "app"
	}
	if parsed.Namespace == "" {
		parsed.Namespace = "default"
	}
	if parsed.Image == "" {
		parsed.Image = fmt.Sprintf("%s:latest", parsed.Name)
	}
	if parsed.Replicas < 1 {
		parsed.Replicas = 1
	}
	if parsed.Profile == "" {
		parsed.Profile = "basic"
	}

	doc := buildDocument(parsed, req)
	manifest, err := yaml.Marshal(doc)
	if err != nil {
		return nil, err
	}

	return &types.KranixAppTemplateResponse{
		Manifest: string(manifest),
		Format:   formatKranixAppYAML,
		Parsed: types.KranixAppTemplateParsed{
			Name:       parsed.Name,
			Image:      parsed.Image,
			Namespace:  parsed.Namespace,
			Replicas:   parsed.Replicas,
			Features:   parsed.Features,
			Profile:    parsed.Profile,
			Confidence: parsed.Confidence,
		},
		Confidence: parsed.Confidence,
	}, nil
}

type parsedTemplate struct {
	Name       string
	Image      string
	Namespace  string
	Replicas   int
	CPU        string
	Memory     string
	Environment string
	Team       string
	Profile    string
	Features   []string
	Env        map[string]string
	Ports      []int
	Confidence float64
}

func parseDescription(input string) parsedTemplate {
	p := parsedTemplate{
		Replicas: 1,
		Env:      make(map[string]string),
		Confidence: 0.5,
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return p
	}

	lower := strings.ToLower(input)

	deployRe := regexp.MustCompile(`(?i)(?:deploy|create|launch|run)\s+(?:the\s+)?(?:"([^"]+)"|'([^']+)'|([\w-]+))`)
	if m := deployRe.FindStringSubmatch(input); len(m) > 0 {
		for i := 1; i <= 3; i++ {
			if m[i] != "" {
				p.Name = m[i]
				break
			}
		}
	}

	imageRe := regexp.MustCompile(`[\w.-]+/[\w.-]+(?::[\w.-]+)?`)
	if img := imageRe.FindString(input); img != "" {
		p.Image = img
	}

	nsRe := regexp.MustCompile(`(?i)(?:to|in|namespace)\s+"?([\w-]+)"?`)
	if m := nsRe.FindStringSubmatch(input); len(m) > 1 {
		p.Namespace = m[1]
	}
	if p.Namespace == "" {
		switch {
		case strings.Contains(lower, "production") || regexp.MustCompile(`\bprod\b`).MatchString(lower):
			p.Namespace = "production"
		case strings.Contains(lower, "staging"):
			p.Namespace = "staging"
		case strings.Contains(lower, "development") || strings.Contains(lower, " dev"):
			p.Namespace = "development"
		}
	}

	replicaRe := regexp.MustCompile(`(?i)(\d+)\s+replicas?`)
	if m := replicaRe.FindStringSubmatch(input); len(m) > 1 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			p.Replicas = n
		}
	}

	cpuRe := regexp.MustCompile(`(?i)(?:cpu|processor)\s*[:=]?\s*"?([\d.]+m?)"?`)
	if m := cpuRe.FindStringSubmatch(input); len(m) > 1 {
		p.CPU = m[1]
	}
	memRe := regexp.MustCompile(`(?i)(?:memory|mem|ram)\s*[:=]?\s*"?([\d.]+(?:Mi|Gi|m)?)"?`)
	if m := memRe.FindStringSubmatch(input); len(m) > 1 {
		p.Memory = m[1]
	}

	envRe := regexp.MustCompile(`(?i)([\w-]+)\s*=\s*"?([\w./-]+)"?`)
	for _, m := range envRe.FindAllStringSubmatch(input, -1) {
		if len(m) > 2 && !isNoiseKey(m[1]) {
			p.Env[m[1]] = m[2]
		}
	}

	p.Features = detectFeatures(lower)
	if strings.Contains(lower, "cron") || strings.Contains(lower, "schedule") {
		p.Profile = "cron"
	} else if strings.Contains(lower, "spot") {
		p.Profile = "spot"
	} else if strings.Contains(lower, "production") || strings.Contains(lower, "critical") {
		p.Profile = "production"
	}

	p.Confidence = calcConfidence(p)
	return p
}

func mergeRequest(parsed *parsedTemplate, req types.KranixAppTemplateRequest) {
	if req.Name != "" {
		parsed.Name = req.Name
	}
	if req.Image != "" {
		parsed.Image = req.Image
	}
	if req.Namespace != "" {
		parsed.Namespace = req.Namespace
	}
	if req.Replicas > 0 {
		parsed.Replicas = req.Replicas
	}
	if req.CPU != "" {
		parsed.CPU = req.CPU
	}
	if req.Memory != "" {
		parsed.Memory = req.Memory
	}
	if req.Environment != "" {
		parsed.Environment = req.Environment
	}
	if req.Team != "" {
		parsed.Team = req.Team
	}
	if req.Profile != "" {
		parsed.Profile = req.Profile
	}
	if len(req.Features) > 0 {
		parsed.Features = uniqueStrings(append(parsed.Features, req.Features...))
	}
	if len(req.Env) > 0 {
		for k, v := range req.Env {
			parsed.Env[k] = v
		}
	}
	if len(req.Ports) > 0 {
		parsed.Ports = req.Ports
	}
}

func buildDocument(p parsedTemplate, req types.KranixAppTemplateRequest) map[string]interface{} {
	spec := map[string]interface{}{
		"image":     p.Image,
		"replicas":  p.Replicas,
		"namespace": p.Namespace,
	}

	if p.CPU != "" || p.Memory != "" {
		res := map[string]interface{}{}
		if p.CPU != "" {
			res["cpu"] = p.CPU
		} else {
			res["cpu"] = defaultCPU(p.Profile)
		}
		if p.Memory != "" {
			res["memory"] = p.Memory
		} else {
			res["memory"] = defaultMemory(p.Profile)
		}
		spec["resources"] = res
	}

	if len(p.Env) > 0 {
		spec["env"] = p.Env
	}

	if len(p.Ports) > 0 {
		ports := make([]map[string]interface{}, 0, len(p.Ports))
		for i, port := range p.Ports {
			ports = append(ports, map[string]interface{}{
				"name":          fmt.Sprintf("port-%d", i+1),
				"containerPort": port,
				"protocol":      "TCP",
			})
		}
		spec["ports"] = ports
	}

	applyProfile(spec, p)
	applyFeatures(spec, p)

	if p.Team != "" || p.Environment != "" {
		tags := map[string]interface{}{}
		if p.Team != "" {
			tags["team"] = p.Team
		}
		if p.Environment != "" {
			tags["environment"] = p.Environment
		} else if p.Namespace != "" {
			tags["environment"] = p.Namespace
		}
		spec["tags"] = tags
	}

	return map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       "KranixApp",
		"metadata": map[string]interface{}{
			"name": p.Name,
		},
		"spec": spec,
	}
}

func applyProfile(spec map[string]interface{}, p parsedTemplate) {
	switch p.Profile {
	case "production":
		spec["autoHeal"] = true
		spec["scheduling"] = map[string]interface{}{
			"workloadPriority":    "critical",
			"preemptionEnabled":   true,
		}
		if res, ok := spec["resources"].(map[string]interface{}); ok {
			if res["cpu"] == nil || res["cpu"] == "" {
				res["cpu"] = "500m"
			}
			if res["memory"] == nil || res["memory"] == "" {
				res["memory"] = "1Gi"
			}
		}
	case "spot":
		spec["scheduling"] = map[string]interface{}{
			"workloadPriority": "normal",
			"spot": map[string]interface{}{
				"enabled":                       true,
				"rescheduleOnNodeTermination": true,
			},
			"tolerations": []map[string]interface{}{
				{
					"key":      "eks.amazonaws.com/capacityType",
					"operator": "Equal",
					"value":    "SPOT",
					"effect":   "NoSchedule",
				},
			},
		}
	case "cron":
		spec["cronSchedule"] = map[string]interface{}{
			"schedule": "0 */6 * * *",
			"timeZone": "UTC",
			"concurrencyPolicy": "Forbid",
		}
	}
}

func applyFeatures(spec map[string]interface{}, p parsedTemplate) {
	for _, f := range p.Features {
		switch strings.ToLower(f) {
		case "spot", "spot-instances":
			applyProfile(spec, parsedTemplate{Profile: "spot"})
		case "critical", "high-priority":
			if sched, ok := spec["scheduling"].(map[string]interface{}); ok {
				sched["workloadPriority"] = "critical"
				sched["preemptionEnabled"] = true
			} else {
				spec["scheduling"] = map[string]interface{}{
					"workloadPriority":  "critical",
					"preemptionEnabled": true,
				}
			}
		case "cross-namespace-traffic", "cross_ns":
			spec["crossNamespaceTraffic"] = map[string]interface{}{
				"enabled":                  true,
				"allowedIngressNamespaces": []string{p.Namespace},
				"allowEgressInternet":      true,
			}
		case "circuit-breaker", "circuit_breaker":
			spec["circuitBreaker"] = map[string]interface{}{
				"enabled":          true,
				"failureThreshold": 5,
				"successThreshold": 2,
				"openDuration":     "30s",
			}
		case "warm-standby", "warm_standby":
			spec["warmStandby"] = map[string]interface{}{
				"enabled":     true,
				"replicas":    1,
				"autoPromote": true,
			}
		case "auto-heal", "auto_heal":
			spec["autoHeal"] = true
		}
	}
}

func detectFeatures(lower string) []string {
	var features []string
	checks := map[string]string{
		"spot":                     "spot",
		"critical":                 "critical",
		"cross-namespace":          "cross-namespace-traffic",
		"circuit breaker":          "circuit-breaker",
		"circuit-breaker":          "circuit-breaker",
		"warm standby":             "warm-standby",
		"warm-standby":             "warm-standby",
		"auto-heal":                "auto-heal",
		"auto heal":                "auto-heal",
	}
	for needle, feature := range checks {
		if strings.Contains(lower, needle) {
			features = append(features, feature)
		}
	}
	return uniqueStrings(features)
}

func defaultCPU(profile string) string {
	if profile == "production" {
		return "500m"
	}
	return "250m"
}

func defaultMemory(profile string) string {
	if profile == "production" {
		return "1Gi"
	}
	return "512Mi"
}

func calcConfidence(p parsedTemplate) float64 {
	score := 0.4
	if p.Name != "" && p.Name != "app" {
		score += 0.2
	}
	if p.Namespace != "" {
		score += 0.1
	}
	if p.Image != "" {
		score += 0.15
	}
	if p.Replicas > 1 {
		score += 0.05
	}
	if len(p.Features) > 0 {
		score += 0.1
	}
	if score > 1 {
		return 1
	}
	return score
}

func isNoiseKey(key string) bool {
	switch strings.ToLower(key) {
	case "deploy", "create", "launch", "replicas", "namespace", "cpu", "memory", "version":
		return true
	}
	return false
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
