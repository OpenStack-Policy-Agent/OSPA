package neutron

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
)

func TestFloatingIpAuditor_ResourceType(t *testing.T) {
	auditor := &FloatingIpAuditor{}
	if got := auditor.ResourceType(); got != "floating_ip" {
		t.Errorf("ResourceType() = %q, want %q", got, "floating_ip")
	}
}

func TestFloatingIpAuditor_Check_StatusMatch(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{
		ID:          "fip-123",
		Description: "test-fip",
		TenantID:    "proj-456",
		Status:      "DOWN",
		FloatingIP:  "192.168.1.100",
	}

	rule := &policy.Rule{
		Name:     "find-down-fips",
		Service:  "neutron",
		Resource: "floating_ip",
		Check:    policy.CheckConditions{Status: "DOWN"},
		Action:   "log",
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for DOWN floating IP")
	}
	if result.ResourceID != "fip-123" {
		t.Errorf("ResourceID = %q, want %q", result.ResourceID, "fip-123")
	}
}

func TestFloatingIpAuditor_Check_StatusNoMatch(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{
		ID:     "fip-123",
		Status: "ACTIVE",
	}

	rule := &policy.Rule{
		Name:  "find-down-fips",
		Check: policy.CheckConditions{Status: "DOWN"},
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for ACTIVE floating IP when checking for DOWN")
	}
}

func TestFloatingIpAuditor_Check_AgeGT(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	oldTime := time.Now().Add(-60 * 24 * time.Hour)
	fip := floatingips.FloatingIP{
		ID:          "fip-123",
		Description: "old-fip",
		UpdatedAt:   oldTime,
	}

	rule := &policy.Rule{
		Name:  "find-old-fips",
		Check: policy.CheckConditions{AgeGT: "30d"},
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for 60-day-old floating IP with 30d threshold")
	}
}

func TestFloatingIpAuditor_Check_ExemptName(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{
		ID:          "fip-123",
		Description: "keep-this",
		Status:      "DOWN",
	}

	rule := &policy.Rule{
		Name:  "find-down-fips",
		Check: policy.CheckConditions{Status: "DOWN", ExemptNames: []string{"keep-this"}},
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for exempt floating IP")
	}
}

func TestFloatingIpAuditor_Check_ExemptNamePattern(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{
		ID:          "fip-456",
		Description: "ospa-e2e-fip-12345",
		Status:      "ACTIVE",
	}

	rule := &policy.Rule{
		Name:  "find-active-fips",
		Check: policy.CheckConditions{Status: "ACTIVE", ExemptNames: []string{"ospa-e2e-*"}},
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for floating IP matching exempt pattern ospa-e2e-*")
	}
	if result.Observation != "exempt by name" {
		t.Errorf("Check() expected observation 'exempt by name', got %q", result.Observation)
	}
}

func TestFloatingIpAuditor_Check_Unused(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{
		ID:     "fip-123",
		PortID: "",
	}

	rule := &policy.Rule{
		Name:  "find-unused-fips",
		Check: policy.CheckConditions{Unused: true},
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for floating IP with no port")
	}
}

func TestFloatingIpAuditor_Check_UsedWithPort(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{
		ID:     "fip-123",
		PortID: "port-789",
	}

	rule := &policy.Rule{
		Name:  "find-unused-fips",
		Check: policy.CheckConditions{Unused: true},
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for floating IP attached to a port")
	}
}

func TestFloatingIpAuditor_Check_Unassociated(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{
		ID:     "fip-123",
		PortID: "",
	}

	rule := &policy.Rule{
		Name:  "find-unassociated-fips",
		Check: policy.CheckConditions{Unassociated: true},
	}

	result, err := auditor.Check(context.Background(), fip, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for unassociated floating IP")
	}
}

func TestFloatingIpAuditor_Check_InvalidType(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	_, err := auditor.Check(context.Background(), "not-a-fip", &policy.Rule{})
	if err == nil {
		t.Error("Check() expected error for invalid resource type")
	}
}

func TestFloatingIpAuditor_Fix_Log(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{ID: "fip-123"}
	rule := &policy.Rule{Action: "log"}

	err := auditor.Fix(context.Background(), nil, fip, rule)
	if err != nil {
		t.Errorf("Fix(log) error = %v, want nil", err)
	}
}

func TestFloatingIpAuditor_Fix_Delete_RequiresClient(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{ID: "fip-123"}
	rule := &policy.Rule{Action: "delete"}

	err := auditor.Fix(context.Background(), nil, fip, rule)
	if err == nil {
		t.Error("Fix(delete) expected error without client")
	}
}

func TestFloatingIpAuditor_Fix_UnsupportedAction(t *testing.T) {
	auditor := &FloatingIpAuditor{}

	fip := floatingips.FloatingIP{ID: "fip-123"}
	rule := &policy.Rule{Action: "reboot"}

	err := auditor.Fix(context.Background(), nil, fip, rule)
	if err == nil {
		t.Error("Fix(reboot) expected error for unsupported action")
	}
}
