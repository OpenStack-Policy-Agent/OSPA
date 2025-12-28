package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Policy is the top-level YAML schema.
type Policy struct {
	Version  string        `yaml:"version"`
	Defaults Defaults       `yaml:"defaults"`
	Rules    []Rule        `yaml:"rules"`
}

type Defaults struct {
	Workers int    `yaml:"workers"`
	Days    int    `yaml:"days"`
	Output  string `yaml:"output"`
}

// Rule is an MVP rule schema. It is intentionally narrow and will evolve.
type Rule struct {
	ID             string `yaml:"id"`
	Description    string `yaml:"description"`
	Resource       string `yaml:"resource"`
	Mode           string `yaml:"mode"`
	Recommendation string `yaml:"recommendation"` // legacy field (kept for backwards-compat)
	Remediation    string `yaml:"remediation"`    // preferred field

	Filters struct {
		Status string `yaml:"status"`
	} `yaml:"filters"`

	Conditions struct {
		UpdatedOlderThanDays int `yaml:"updatedOlderThanDays"`
	} `yaml:"conditions"`
}

func Load(path string) (*Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("parse policy yaml: %w", err)
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *Policy) Validate() error {
	if p.Version == "" {
		return fmt.Errorf("policy.version is required")
	}
	if len(p.Rules) == 0 {
		return fmt.Errorf("policy.rules must contain at least one rule")
	}

	seen := map[string]struct{}{}
	for i, r := range p.Rules {
		if r.ID == "" {
			return fmt.Errorf("policy.rules[%d].id is required", i)
		}
		if _, ok := seen[r.ID]; ok {
			return fmt.Errorf("duplicate rule id %q", r.ID)
		}
		seen[r.ID] = struct{}{}

		if r.Resource != "compute.server" {
			return fmt.Errorf("rule %q: unsupported resource %q (supported: compute.server)", r.ID, r.Resource)
		}
		if r.Mode == "" {
			return fmt.Errorf("rule %q: mode is required (audit/enforce)", r.ID)
		}
		if r.Mode != "audit" && r.Mode != "enforce" {
			return fmt.Errorf("rule %q: unsupported mode %q (audit/enforce)", r.ID, r.Mode)
		}
		if r.Filters.Status == "" {
			return fmt.Errorf("rule %q: filters.status is required", r.ID)
		}
		if r.Conditions.UpdatedOlderThanDays < 0 {
			return fmt.Errorf("rule %q: conditions.updatedOlderThanDays must be >= 0", r.ID)
		}
		if r.Mode == "enforce" && r.EffectiveRemediation() == "" {
			return fmt.Errorf("rule %q: remediation is required when mode=enforce", r.ID)
		}
	}

	return nil
}

// EffectiveWorkers returns the configured workers or a safe default.
func (p *Policy) EffectiveWorkers(fallback int) int {
	if p.Defaults.Workers > 0 {
		return p.Defaults.Workers
	}
	if fallback > 0 {
		return fallback
	}
	return 16
}

// EffectiveRemediation returns the remediation action for the rule.
// Backwards-compatible: if remediation isn't set, it falls back to recommendation.
func (r *Rule) EffectiveRemediation() string {
	if r.Remediation != "" {
		return r.Remediation
	}
	return r.Recommendation
}


