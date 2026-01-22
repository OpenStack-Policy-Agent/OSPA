package cinder

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

func TestVolumeAuditor_ResourceType(t *testing.T) {
	auditor := &VolumeAuditor{}
	if got := auditor.ResourceType(); got != "volume" {
		t.Errorf("ResourceType() = %q, want %q", got, "volume")
	}
}

func TestVolumeAuditor_Check(t *testing.T) {
	auditor := &VolumeAuditor{}

	// TODO: Replace with the real gophercloud type once the auditor is implemented.
	resource := map[string]interface{}{"id": "test-id", "name": "test-resource"}

	rule := &policy.Rule{
		Name:     "test-rule",
		Service:  "cinder",
		Resource: "volume",
		Check:    policy.CheckConditions{Status: "active"},
		Action:   "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result == nil {
		t.Fatal("Check() returned nil result")
	}
	if result.RuleID != rule.Name {
		t.Errorf("Result.RuleID = %q, want %q", result.RuleID, rule.Name)
	}
}

func TestVolumeAuditor_Fix(t *testing.T) {
	t.Skip("Fix() requires a mock gophercloud client")
}
