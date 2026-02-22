package neutron

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

func TestSubnetAuditor_ResourceType(t *testing.T) {
	auditor := &SubnetAuditor{}
	if got := auditor.ResourceType(); got != "subnet" {
		t.Errorf("ResourceType() = %q, want %q", got, "subnet")
	}
}

func TestSubnetAuditor_Check_Compliant(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{
		ID:       "sub-123",
		Name:     "test-subnet",
		TenantID: "proj-456",
		CIDR:     "10.0.0.0/24",
		AllocationPools: []subnets.AllocationPool{
			{Start: "10.0.0.2", End: "10.0.0.254"},
		},
	}

	rule := &policy.Rule{
		Name:     "find-unused-subnets",
		Service:  "neutron",
		Resource: "subnet",
		Check:    policy.CheckConditions{Unused: true},
		Action:   "log",
	}

	result, err := auditor.Check(context.Background(), subnet, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for subnet with allocation pools")
	}
	if result.ResourceID != "sub-123" {
		t.Errorf("ResourceID = %q, want %q", result.ResourceID, "sub-123")
	}
	if result.ProjectID != "proj-456" {
		t.Errorf("ProjectID = %q, want %q", result.ProjectID, "proj-456")
	}
}

func TestSubnetAuditor_Check_Unused(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{
		ID:              "sub-123",
		Name:            "empty-subnet",
		AllocationPools: []subnets.AllocationPool{},
	}

	rule := &policy.Rule{
		Name:  "find-unused-subnets",
		Check: policy.CheckConditions{Unused: true},
	}

	result, err := auditor.Check(context.Background(), subnet, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for subnet with no allocation pools")
	}
}

func TestSubnetAuditor_Check_UnusedNotRequested(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{
		ID:              "sub-123",
		Name:            "empty-subnet",
		AllocationPools: []subnets.AllocationPool{},
	}

	rule := &policy.Rule{
		Name:  "some-rule",
		Check: policy.CheckConditions{},
	}

	result, err := auditor.Check(context.Background(), subnet, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant when unused check is not requested")
	}
}

func TestSubnetAuditor_Check_ExemptName(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{
		ID:              "sub-123",
		Name:            "default",
		AllocationPools: []subnets.AllocationPool{},
	}

	rule := &policy.Rule{
		Name:  "find-unused-subnets",
		Check: policy.CheckConditions{Unused: true, ExemptNames: []string{"default"}},
	}

	result, err := auditor.Check(context.Background(), subnet, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for exempt subnet")
	}
}

func TestSubnetAuditor_Check_ExemptNamePattern(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{
		ID:   "sub-456",
		Name: "ospa-e2e-subnet-12345",
	}

	rule := &policy.Rule{
		Name:  "find-unused-subnets",
		Check: policy.CheckConditions{Unused: true, ExemptNames: []string{"ospa-e2e-*"}},
	}

	result, err := auditor.Check(context.Background(), subnet, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for subnet matching exempt pattern ospa-e2e-*")
	}
	if result.Observation != "exempt by name" {
		t.Errorf("Check() expected observation 'exempt by name', got %q", result.Observation)
	}
}

func TestSubnetAuditor_Check_InvalidType(t *testing.T) {
	auditor := &SubnetAuditor{}

	_, err := auditor.Check(context.Background(), "not-a-subnet", &policy.Rule{})
	if err == nil {
		t.Error("Check() expected error for invalid resource type")
	}
}

func TestSubnetAuditor_Fix_Log(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{ID: "sub-123"}
	rule := &policy.Rule{Action: "log"}

	err := auditor.Fix(context.Background(), nil, subnet, rule)
	if err != nil {
		t.Errorf("Fix(log) error = %v, want nil", err)
	}
}

func TestSubnetAuditor_Fix_Delete_RequiresClient(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{ID: "sub-123"}
	rule := &policy.Rule{Action: "delete"}

	err := auditor.Fix(context.Background(), nil, subnet, rule)
	if err == nil {
		t.Error("Fix(delete) expected error without client")
	}
}

func TestSubnetAuditor_Fix_UnsupportedAction(t *testing.T) {
	auditor := &SubnetAuditor{}

	subnet := subnets.Subnet{ID: "sub-123"}
	rule := &policy.Rule{Action: "tag"}

	err := auditor.Fix(context.Background(), nil, subnet, rule)
	if err == nil {
		t.Error("Fix(tag) expected error for unimplemented action")
	}
}
