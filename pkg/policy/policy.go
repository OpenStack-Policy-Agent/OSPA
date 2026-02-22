package policy

import (
	"fmt"
	"time"
)

// Policy represents the top-level policy configuration
type Policy struct {
	Version    string                   `yaml:"version"`
	Defaults   Defaults                 `yaml:"defaults"`
	Policies   []ServicePolicy          `yaml:"policies"`
	Composites []CompositeServicePolicy `yaml:"composites,omitempty"`
}

// Defaults contains default configuration values
type Defaults struct {
	Workers int    `yaml:"workers"`
	Days    int    `yaml:"days"`
	Output  string `yaml:"output"`
}

// ServicePolicy groups rules by OpenStack service
type ServicePolicy struct {
	Service string `yaml:"service"`
	Rules   []Rule `yaml:"rules"`
}

// CompositeServicePolicy groups composite rules by OpenStack service.
type CompositeServicePolicy struct {
	Service string          `yaml:"service"`
	Rules   []CompositeRule `yaml:"rules"`
}

// Severity constants for rule classification
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Category constants for rule classification
const (
	CategorySecurity   = "security"
	CategoryCompliance = "compliance"
	CategoryCost       = "cost"
	CategoryHygiene    = "hygiene"
)

// Rule represents a single audit rule
type Rule struct {
	Name          string          `yaml:"name"`
	Description   string          `yaml:"description"`
	Service       string          `yaml:"service,omitempty"`
	Resource      string          `yaml:"resource"`
	Check         CheckConditions `yaml:"check"`
	Action        string          `yaml:"action"`
	Severity      string          `yaml:"severity,omitempty"`
	Category      string          `yaml:"category,omitempty"`
	GuideRef      string          `yaml:"guide_ref,omitempty"`
	ActionTagName string          `yaml:"action_tag_name,omitempty"`
	TagName       string          `yaml:"tag_name,omitempty"`
}

// CompositeRule represents a rule that evaluates multiple resource types together.
type CompositeRule struct {
	Name          string                 `yaml:"name"`
	Description   string                 `yaml:"description"`
	Service       string                 `yaml:"service"`
	Resources     []string               `yaml:"resources"`
	Check         map[string]interface{} `yaml:"check"`
	Action        string                 `yaml:"action"`
	Severity      string                 `yaml:"severity,omitempty"`
	Category      string                 `yaml:"category,omitempty"`
	GuideRef      string                 `yaml:"guide_ref,omitempty"`
	ActionTagName string                 `yaml:"action_tag_name,omitempty"`
	TagName       string                 `yaml:"tag_name,omitempty"`
}

// CheckConditions supports flexible condition matching.
//
// Fields are grouped by scope: universal checks apply to all resources,
// while service-specific checks are gated by the validation layer to
// only the resources where they are meaningful.
type CheckConditions struct {
	// --- Universal checks (all resources) ---

	Status         string         `yaml:"status,omitempty"`
	AgeGT          string         `yaml:"age_gt,omitempty"`
	Unused         bool           `yaml:"unused,omitempty"`
	ExemptNames    []string       `yaml:"exempt_names,omitempty"`
	ExemptMetadata *MetadataMatch `yaml:"exempt_metadata,omitempty"`

	// --- Neutron checks ---

	Direction      string `yaml:"direction,omitempty"`
	Ethertype      string `yaml:"ethertype,omitempty"`
	Protocol       string `yaml:"protocol,omitempty"`
	Port           int    `yaml:"port,omitempty"`
	RemoteIPPrefix string `yaml:"remote_ip_prefix,omitempty"`
	PortRangeWide  bool   `yaml:"port_range_wide,omitempty"`
	Unassociated   bool   `yaml:"unassociated,omitempty"`
	SharedNetwork  bool   `yaml:"shared_network,omitempty"`
	NoSecurityGroup bool  `yaml:"no_security_group,omitempty"`

	// --- Nova checks ---

	ImageName []string `yaml:"image_name,omitempty"`
	NoKeypair bool     `yaml:"no_keypair,omitempty"`

	// --- Cinder checks ---

	Encrypted *bool `yaml:"encrypted,omitempty"`
	Attached  *bool `yaml:"attached,omitempty"`
	HasBackup *bool `yaml:"has_backup,omitempty"`

	// --- Glance checks ---

	Visibility string `yaml:"visibility,omitempty"`

	// --- Keystone checks ---

	PasswordExpired bool   `yaml:"password_expired,omitempty"`
	MFAEnabled      *bool  `yaml:"mfa_enabled,omitempty"`
	InactiveDays    int    `yaml:"inactive_days,omitempty"`
	HasAdminRole    bool   `yaml:"has_admin_role,omitempty"`
	TokenProvider   string `yaml:"token_provider,omitempty"`
}

// UsedChecks returns the YAML field names of all non-zero check conditions.
// The orchestrator compares this against Auditor.ImplementedChecks() to
// detect policy rules that reference checks no auditor handles.
func (c *CheckConditions) UsedChecks() []string {
	if c == nil {
		return nil
	}
	var used []string
	if c.Status != "" {
		used = append(used, "status")
	}
	if c.AgeGT != "" {
		used = append(used, "age_gt")
	}
	if c.Unused {
		used = append(used, "unused")
	}
	if len(c.ExemptNames) > 0 {
		used = append(used, "exempt_names")
	}
	if c.ExemptMetadata != nil {
		used = append(used, "exempt_metadata")
	}
	if c.Direction != "" {
		used = append(used, "direction")
	}
	if c.Ethertype != "" {
		used = append(used, "ethertype")
	}
	if c.Protocol != "" {
		used = append(used, "protocol")
	}
	if c.Port != 0 {
		used = append(used, "port")
	}
	if c.RemoteIPPrefix != "" {
		used = append(used, "remote_ip_prefix")
	}
	if c.PortRangeWide {
		used = append(used, "port_range_wide")
	}
	if c.Unassociated {
		used = append(used, "unassociated")
	}
	if c.SharedNetwork {
		used = append(used, "shared_network")
	}
	if c.NoSecurityGroup {
		used = append(used, "no_security_group")
	}
	if len(c.ImageName) > 0 {
		used = append(used, "image_name")
	}
	if c.NoKeypair {
		used = append(used, "no_keypair")
	}
	if c.Encrypted != nil {
		used = append(used, "encrypted")
	}
	if c.Attached != nil {
		used = append(used, "attached")
	}
	if c.HasBackup != nil {
		used = append(used, "has_backup")
	}
	if c.Visibility != "" {
		used = append(used, "visibility")
	}
	if c.PasswordExpired {
		used = append(used, "password_expired")
	}
	if c.MFAEnabled != nil {
		used = append(used, "mfa_enabled")
	}
	if c.InactiveDays != 0 {
		used = append(used, "inactive_days")
	}
	if c.HasAdminRole {
		used = append(used, "has_admin_role")
	}
	if c.TokenProvider != "" {
		used = append(used, "token_provider")
	}
	return used
}

// MetadataMatch represents metadata key-value matching for exemptions
type MetadataMatch struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// ParseAgeGT parses an age_gt string (e.g., "30d", "7d") into a time.Duration
func (c *CheckConditions) ParseAgeGT() (time.Duration, error) {
	if c.AgeGT == "" {
		return 0, nil
	}

	var duration time.Duration
	var unit string
	var value int

	_, err := fmt.Sscanf(c.AgeGT, "%d%s", &value, &unit)
	if err != nil {
		return 0, fmt.Errorf("invalid age_gt format %q: %w", c.AgeGT, err)
	}

	switch unit {
	case "d", "day", "days":
		duration = time.Duration(value) * 24 * time.Hour
	case "h", "hour", "hours":
		duration = time.Duration(value) * time.Hour
	case "m", "min", "minute", "minutes":
		duration = time.Duration(value) * time.Minute
	default:
		return 0, fmt.Errorf("unsupported age_gt unit %q (supported: d, h, m)", unit)
	}

	return duration, nil
}

// EffectiveWorkers returns the configured workers or a safe default
func (p *Policy) EffectiveWorkers(fallback int) int {
	if p.Defaults.Workers > 0 {
		return p.Defaults.Workers
	}
	if fallback > 0 {
		return fallback
	}
	return 16
}

// GetAllRules returns all rules from all service policies, flattened
func (p *Policy) GetAllRules() []Rule {
	var allRules []Rule
	for _, sp := range p.Policies {
		for _, rule := range sp.Rules {
			// Ensure service is set from parent if not in rule
			if rule.Service == "" {
				rule.Service = sp.Service
			}
			allRules = append(allRules, rule)
		}
	}
	return allRules
}

// GetAllCompositeRules returns all composite rules from all composite service policies.
func (p *Policy) GetAllCompositeRules() []CompositeRule {
	var allRules []CompositeRule
	for _, sp := range p.Composites {
		for _, rule := range sp.Rules {
			if rule.Service == "" {
				rule.Service = sp.Service
			}
			allRules = append(allRules, rule)
		}
	}
	return allRules
}
