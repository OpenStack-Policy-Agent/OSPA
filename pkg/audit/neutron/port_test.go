package neutron

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

func TestPortAuditor_ResourceType(t *testing.T) {
	a := &PortAuditor{}
	if got := a.ResourceType(); got != "port" {
		t.Errorf("ResourceType() = %q, want %q", got, "port")
	}
}

func TestPortAuditor_Check_StatusMatch(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "test-port", Status: "ACTIVE", TenantID: "t1"}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{Status: "ACTIVE"}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Compliant {
		t.Error("expected non-compliant when status matches")
	}
}

func TestPortAuditor_Check_StatusNoMatch(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "test-port", Status: "DOWN", TenantID: "t1"}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{Status: "ACTIVE"}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant when status does not match")
	}
}

func TestPortAuditor_Check_AgeGT_Compliant(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{
		ID:        "p1",
		Name:      "test-port",
		TenantID:  "t1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{AgeGT: "30d"}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant for freshly created port")
	}
}

func TestPortAuditor_Check_AgeGT_Violation(t *testing.T) {
	a := &PortAuditor{}
	old := time.Now().Add(-60 * 24 * time.Hour)
	port := ports.Port{
		ID:        "p1",
		Name:      "test-port",
		TenantID:  "t1",
		CreatedAt: old,
		UpdatedAt: old,
	}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{AgeGT: "30d"}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Compliant {
		t.Error("expected non-compliant for port older than 30d")
	}
}

func TestPortAuditor_Check_ExemptByName(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "keep-me", Status: "ACTIVE", TenantID: "t1"}
	rule := &policy.Rule{
		Name: "r1",
		Check: policy.CheckConditions{
			Status:      "ACTIVE",
			ExemptNames: []string{"keep-me"},
		},
	}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant when name is exempt")
	}
}

func TestPortAuditor_Check_ExemptByPattern(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "infra-port-01", Status: "ACTIVE", TenantID: "t1"}
	rule := &policy.Rule{
		Name: "r1",
		Check: policy.CheckConditions{
			Status:      "ACTIVE",
			ExemptNames: []string{"infra-*"},
		},
	}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant when name matches exempt pattern")
	}
}

func TestPortAuditor_Check_Unused_NoDevice(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "orphan-port", TenantID: "t1", DeviceID: ""}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{Unused: true}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Compliant {
		t.Error("expected non-compliant when port has no device attached")
	}
}

func TestPortAuditor_Check_Unused_WithDevice(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "attached-port", TenantID: "t1", DeviceID: "vm-123"}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{Unused: true}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant when port has a device attached")
	}
}

func TestPortAuditor_Check_NoSecurityGroup_Empty(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "unprotected", TenantID: "t1", SecurityGroups: []string{}}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{NoSecurityGroup: true}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Compliant {
		t.Error("expected non-compliant when port has no security groups")
	}
}

func TestPortAuditor_Check_NoSecurityGroup_HasGroups(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1", Name: "protected", TenantID: "t1", SecurityGroups: []string{"sg-1"}}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{NoSecurityGroup: true}}

	result, err := a.Check(context.Background(), port, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant when port has security groups")
	}
}

func TestPortAuditor_Check_InvalidType(t *testing.T) {
	a := &PortAuditor{}
	rule := &policy.Rule{Name: "r1", Check: policy.CheckConditions{Status: "ACTIVE"}}

	_, err := a.Check(context.Background(), "not-a-port", rule)
	if err == nil {
		t.Error("expected error for invalid resource type")
	}
}

func TestPortAuditor_Fix_Log(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1"}
	rule := &policy.Rule{Name: "r1", Action: "log"}

	err := a.Fix(context.Background(), nil, port, rule)
	if err != nil {
		t.Errorf("expected no error for log action, got: %v", err)
	}
}

func TestPortAuditor_Fix_Delete_RequiresClient(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1"}
	rule := &policy.Rule{Name: "r1", Action: "delete"}

	err := a.Fix(context.Background(), "not-a-client", port, rule)
	if err == nil {
		t.Error("expected error when client is wrong type")
	}
}

func TestPortAuditor_Fix_UnsupportedAction(t *testing.T) {
	a := &PortAuditor{}
	port := ports.Port{ID: "p1"}
	rule := &policy.Rule{Name: "r1", Action: "tag"}

	err := a.Fix(context.Background(), nil, port, rule)
	if err == nil {
		t.Error("expected error for tag action (not implemented)")
	}
}
