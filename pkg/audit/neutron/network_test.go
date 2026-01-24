package neutron

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
)

func TestNetworkAuditor_ResourceType(t *testing.T) {
	auditor := &NetworkAuditor{}
	if got := auditor.ResourceType(); got != "network" {
		t.Errorf("ResourceType() = %q, want %q", got, "network")
	}
}

func TestNetworkAuditor_Check_StatusMatch(t *testing.T) {
	auditor := &NetworkAuditor{}

	network := networks.Network{
		ID:       "net-123",
		Name:     "test-network",
		TenantID: "proj-456",
		Status:   "DOWN",
	}

	rule := &policy.Rule{
		Name:     "find-down-networks",
		Service:  "neutron",
		Resource: "network",
		Check:    policy.CheckConditions{Status: "DOWN"},
		Action:   "log",
	}

	result, err := auditor.Check(context.Background(), network, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for DOWN network")
	}
	if result.ResourceID != "net-123" {
		t.Errorf("ResourceID = %q, want %q", result.ResourceID, "net-123")
	}
}

func TestNetworkAuditor_Check_StatusNoMatch(t *testing.T) {
	auditor := &NetworkAuditor{}

	network := networks.Network{
		ID:     "net-123",
		Name:   "test-network",
		Status: "ACTIVE",
	}

	rule := &policy.Rule{
		Name:  "find-down-networks",
		Check: policy.CheckConditions{Status: "DOWN"},
	}

	result, err := auditor.Check(context.Background(), network, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for ACTIVE network when checking for DOWN")
	}
}

func TestNetworkAuditor_Check_AgeGT(t *testing.T) {
	auditor := &NetworkAuditor{}

	oldTime := time.Now().Add(-60 * 24 * time.Hour) // 60 days ago
	network := networks.Network{
		ID:        "net-123",
		Name:      "old-network",
		UpdatedAt: oldTime,
	}

	rule := &policy.Rule{
		Name:  "find-old-networks",
		Check: policy.CheckConditions{AgeGT: "30d"},
	}

	result, err := auditor.Check(context.Background(), network, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for 60-day-old network with 30d threshold")
	}
}

func TestNetworkAuditor_Check_ExemptName(t *testing.T) {
	auditor := &NetworkAuditor{}

	network := networks.Network{
		ID:     "net-123",
		Name:   "default",
		Status: "DOWN",
	}

	rule := &policy.Rule{
		Name:  "find-down-networks",
		Check: policy.CheckConditions{Status: "DOWN", ExemptNames: []string{"default"}},
	}

	result, err := auditor.Check(context.Background(), network, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for exempt network")
	}
}

func TestNetworkAuditor_Check_Unused(t *testing.T) {
	auditor := &NetworkAuditor{}

	// Network with no subnets
	network := networks.Network{
		ID:      "net-123",
		Name:    "empty-network",
		Subnets: []string{},
	}

	rule := &policy.Rule{
		Name:  "find-unused-networks",
		Check: policy.CheckConditions{Unused: true},
	}

	result, err := auditor.Check(context.Background(), network, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for network with no subnets")
	}
}

func TestNetworkAuditor_Check_InvalidType(t *testing.T) {
	auditor := &NetworkAuditor{}

	_, err := auditor.Check(context.Background(), "not-a-network", &policy.Rule{})
	if err == nil {
		t.Error("Check() expected error for invalid resource type")
	}
}

func TestNetworkAuditor_Fix_Log(t *testing.T) {
	auditor := &NetworkAuditor{}

	network := networks.Network{ID: "net-123"}
	rule := &policy.Rule{Action: "log"}

	// Log action should always succeed without a client
	err := auditor.Fix(context.Background(), nil, network, rule)
	if err != nil {
		t.Errorf("Fix(log) error = %v, want nil", err)
	}
}

func TestNetworkAuditor_Fix_Delete_RequiresClient(t *testing.T) {
	auditor := &NetworkAuditor{}

	network := networks.Network{ID: "net-123"}
	rule := &policy.Rule{Action: "delete"}

	// Delete action should fail without proper client
	err := auditor.Fix(context.Background(), nil, network, rule)
	if err == nil {
		t.Error("Fix(delete) expected error without client")
	}
}
