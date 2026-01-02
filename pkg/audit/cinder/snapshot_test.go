package cinder

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

func TestSnapshotAuditor_ResourceType(t *testing.T) {
	auditor := &SnapshotAuditor{}
	if got := auditor.ResourceType(); got != "snapshot" {
		t.Errorf("ResourceType() = %q, want %q", got, "snapshot")
	}
}

func TestSnapshotAuditor_Check(t *testing.T) {
	auditor := &SnapshotAuditor{}

	// TODO(OSPA): Replace this placeholder resource with the real SDK type used by the discoverer.
	resource := map[string]interface{}{"id": "test-id", "name": "test-resource"}

	rule := &policy.Rule{
		Name:     "test-rule",
		Service:  "cinder",
		Resource: "snapshot",
		Check: policy.CheckConditions{
			Status: "active",
		},
		Action: "log",
	}

	ctx := context.Background()
	result, err := auditor.Check(ctx, resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result == nil {
		t.Fatal("Check() returned nil result")
	}

	if result.RuleID != rule.Name {
		t.Errorf("Result.RuleID = %q, want %q", result.RuleID, rule.Name)
	}

	// TODO(OSPA): Add assertions for ResourceID/ProjectID/Compliant/Observation once real extraction is implemented.
}

func TestSnapshotAuditor_Check_AgeGT(t *testing.T) {
	t.Skip("placeholder auditor does not implement age-based checks yet")
}

func TestSnapshotAuditor_Fix(t *testing.T) {
	// TODO: Implement integration test with mock client
	// This requires setting up a mock gophercloud client
	t.Skip("Fix() test requires mock client setup")
}
