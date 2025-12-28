package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRule_EffectiveRemediation_PrefersRemediationOverRecommendation(t *testing.T) {
	r := Rule{
		Remediation:    "delete",
		Recommendation: "noop",
	}
	if got := r.EffectiveRemediation(); got != "delete" {
		t.Fatalf("expected delete, got %q", got)
	}
}

func TestRule_EffectiveRemediation_FallsBackToRecommendation(t *testing.T) {
	r := Rule{
		Recommendation: "delete",
	}
	if got := r.EffectiveRemediation(); got != "delete" {
		t.Fatalf("expected delete, got %q", got)
	}
}

func TestPolicy_Validate_RequiresVersionAndRules(t *testing.T) {
	p := &Policy{}
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error for missing version")
	}
	p.Version = "v1"
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error for missing rules")
	}
}

func TestPolicy_Validate_RejectsDuplicateRuleIDs(t *testing.T) {
	r1 := Rule{ID: "dup", Resource: "compute.server", Mode: "audit"}
	r1.Filters.Status = "SHUTOFF"
	r1.Conditions.UpdatedOlderThanDays = 30

	r2 := Rule{ID: "dup", Resource: "compute.server", Mode: "audit"}
	r2.Filters.Status = "SHUTOFF"
	r2.Conditions.UpdatedOlderThanDays = 30

	p := &Policy{Version: "v1", Rules: []Rule{r1, r2}}
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error for duplicate rule ids")
	}
}

func TestPolicy_Validate_EnforceRequiresRemediation(t *testing.T) {
	r := Rule{ID: "r1", Resource: "compute.server", Mode: "enforce"}
	r.Filters.Status = "SHUTOFF"
	r.Conditions.UpdatedOlderThanDays = 30
	p := &Policy{Version: "v1", Rules: []Rule{r}}
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error when mode=enforce but remediation is empty")
	}

	p.Rules[0].Remediation = "delete"
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid policy when remediation is present, got: %v", err)
	}
}

func TestPolicy_EffectiveWorkers(t *testing.T) {
	p := &Policy{Version: "v1"}
	if got := p.EffectiveWorkers(99); got != 99 {
		t.Fatalf("expected fallback workers, got %d", got)
	}
	p.Defaults.Workers = 50
	if got := p.EffectiveWorkers(99); got != 50 {
		t.Fatalf("expected defaults workers, got %d", got)
	}
}

func TestLoad_ParsesYAMLAndValidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(path, []byte(`
version: v1
defaults:
  workers: 10
rules:
  - id: r1
    resource: compute.server
    mode: audit
    filters:
      status: SHUTOFF
    conditions:
      updatedOlderThanDays: 30
    remediation: delete
`), 0o644); err != nil {
		t.Fatalf("write temp policy: %v", err)
	}

	p, err := Load(path)
	if err != nil {
		t.Fatalf("expected load to succeed, got: %v", err)
	}
	if p.Version != "v1" {
		t.Fatalf("expected version v1, got %q", p.Version)
	}
	if len(p.Rules) != 1 || p.Rules[0].ID != "r1" {
		t.Fatalf("unexpected rules parsed: %+v", p.Rules)
	}
}


