package neutron

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
)

func TestSecurityGroupAuditor_ResourceType(t *testing.T) {
	auditor := &SecurityGroupAuditor{}
	if got := auditor.ResourceType(); got != "security_group" {
		t.Errorf("ResourceType() = %q, want %q", got, "security_group")
	}
}

func TestSecurityGroupAuditor_Check(t *testing.T) {
	auditor := &SecurityGroupAuditor{}

	// Create a proper gophercloud SecGroup for testing
	resource := groups.SecGroup{
		ID:          "test-sg-id",
		Name:        "test-security-group",
		Description: "Test security group",
		TenantID:    "test-tenant-id",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-1 * time.Hour),
	}

	rule := &policy.Rule{
		Name:     "test-rule",
		Service:  "neutron",
		Resource: "security_group",
		Check:    policy.CheckConditions{Status: "ACTIVE"},
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
	if result.ResourceID != resource.ID {
		t.Errorf("Result.ResourceID = %q, want %q", result.ResourceID, resource.ID)
	}
	if result.ResourceName != resource.Name {
		t.Errorf("Result.ResourceName = %q, want %q", result.ResourceName, resource.Name)
	}
	if result.ProjectID != resource.TenantID {
		t.Errorf("Result.ProjectID = %q, want %q", result.ProjectID, resource.TenantID)
	}
}

func TestSecurityGroupAuditor_Check_Exempt(t *testing.T) {
	auditor := &SecurityGroupAuditor{}

	resource := groups.SecGroup{
		ID:       "test-sg-id",
		Name:     "default",
		TenantID: "test-tenant-id",
	}

	rule := &policy.Rule{
		Name:     "test-exempt-rule",
		Service:  "neutron",
		Resource: "security_group",
		Check: policy.CheckConditions{
			Unused:      true,
			ExemptNames: []string{"default"},
		},
		Action: "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Errorf("Expected resource to be compliant (exempted), got non-compliant")
	}
	if result.Observation != "exempt by name" {
		t.Errorf("Expected observation 'exempt by name', got %q", result.Observation)
	}
}

func TestSecurityGroupAuditor_Check_AgeGT(t *testing.T) {
	auditor := &SecurityGroupAuditor{}

	// Create a security group that's 60 days old
	resource := groups.SecGroup{
		ID:        "old-sg-id",
		Name:      "old-security-group",
		TenantID:  "test-tenant-id",
		CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-60 * 24 * time.Hour),
	}

	rule := &policy.Rule{
		Name:     "test-age-rule",
		Service:  "neutron",
		Resource: "security_group",
		Check: policy.CheckConditions{
			AgeGT: "30d",
		},
		Action: "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Expected resource to be non-compliant (older than 30d)")
	}
}

func TestSecurityGroupAuditor_Fix(t *testing.T) {
	t.Skip("Fix() requires a mock gophercloud client")
}
