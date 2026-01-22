package neutron

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

func TestNetworkAuditor_ResourceType(t *testing.T) {
	auditor := &NetworkAuditor{}
	if got := auditor.ResourceType(); got != "network" {
		t.Errorf("ResourceType() = %q, want %q", got, "network")
	}
}

func TestNetworkAuditor_Check(t *testing.T) {
	auditor := &NetworkAuditor{}

	// TODO: Replace with the real gophercloud type once the auditor is implemented.
	resource := map[string]interface{}{"id": "test-id", "name": "test-resource"}

	rule := &policy.Rule{
		Name:     "test-rule",
		Service:  "neutron",
		Resource: "network",
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

func TestNetworkAuditor_Fix(t *testing.T) {
	t.Skip("Fix() requires a mock gophercloud client")
}
