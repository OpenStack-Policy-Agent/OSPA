package policy

import (
	"fmt"
	"time"
)

// Policy represents the top-level policy configuration
type Policy struct {
	Version  string         `yaml:"version"`
	Defaults Defaults       `yaml:"defaults"`
	Policies []ServicePolicy `yaml:"policies"`
}

// Defaults contains default configuration values
type Defaults struct {
	Workers int    `yaml:"workers"`
	Days    int    `yaml:"days"`
	Output  string `yaml:"output"`
}

// ServicePolicy groups rules by OpenStack service
type ServicePolicy struct {
	Service string  `yaml:"service"`
	Rules   []Rule  `yaml:"rules"`
}

// Rule represents a single audit rule
type Rule struct {
	Name          string          `yaml:"name"`
	Description   string          `yaml:"description"`
	Service       string          `yaml:"service"`
	Resource      string          `yaml:"resource"`
	Check         CheckConditions `yaml:"check"`
	Action        string          `yaml:"action"`
	ActionTagName string          `yaml:"action_tag_name,omitempty"`
	TagName       string          `yaml:"tag_name,omitempty"`
}

// CheckConditions supports flexible condition matching
type CheckConditions struct {
	// Security group rule checks
	Direction      string `yaml:"direction,omitempty"`
	Ethertype      string `yaml:"ethertype,omitempty"`
	Protocol       string `yaml:"protocol,omitempty"`
	Port           int    `yaml:"port,omitempty"`
	RemoteIPPrefix string `yaml:"remote_ip_prefix,omitempty"`

	// Status-based checks
	Status string `yaml:"status,omitempty"`

	// Age-based checks (e.g., "30d", "7d", "90d")
	AgeGT string `yaml:"age_gt,omitempty"`

	// Usage checks
	Unused bool `yaml:"unused,omitempty"`

	// Exemptions
	ExemptNames    []string      `yaml:"exempt_names,omitempty"`
	ExemptMetadata *MetadataMatch `yaml:"exempt_metadata,omitempty"`

	// Image name checks
	ImageName []string `yaml:"image_name,omitempty"`
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
