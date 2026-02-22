package neutron

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
)

func TestRouterAuditor_ResourceType(t *testing.T) {
	auditor := &RouterAuditor{}
	if got := auditor.ResourceType(); got != "router" {
		t.Errorf("ResourceType() = %q, want %q", got, "router")
	}
}

func TestRouterAuditor_Check_StatusMatch(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{
		ID:       "rtr-123",
		Name:     "test-router",
		TenantID: "proj-456",
		Status:   "DOWN",
	}

	rule := &policy.Rule{
		Name:     "find-down-routers",
		Service:  "neutron",
		Resource: "router",
		Check:    policy.CheckConditions{Status: "DOWN"},
		Action:   "log",
	}

	result, err := auditor.Check(context.Background(), router, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for DOWN router")
	}
	if result.ResourceID != "rtr-123" {
		t.Errorf("ResourceID = %q, want %q", result.ResourceID, "rtr-123")
	}
}

func TestRouterAuditor_Check_StatusNoMatch(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{
		ID:     "rtr-123",
		Name:   "test-router",
		Status: "ACTIVE",
	}

	rule := &policy.Rule{
		Name:  "find-down-routers",
		Check: policy.CheckConditions{Status: "DOWN"},
	}

	result, err := auditor.Check(context.Background(), router, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for ACTIVE router when checking for DOWN")
	}
}

func TestRouterAuditor_Check_AgeGT_NoOp(t *testing.T) {
	auditor := &RouterAuditor{}

	// gophercloud v1.14.1 routers.Router has no timestamp fields,
	// so age_gt is accepted but always compliant (no-op).
	router := routers.Router{
		ID:   "rtr-123",
		Name: "old-router",
	}

	rule := &policy.Rule{
		Name:  "find-old-routers",
		Check: policy.CheckConditions{AgeGT: "30d"},
	}

	result, err := auditor.Check(context.Background(), router, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant (age_gt is a no-op without timestamps)")
	}
}

func TestRouterAuditor_Check_ExemptName(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{
		ID:     "rtr-123",
		Name:   "default",
		Status: "DOWN",
	}

	rule := &policy.Rule{
		Name:  "find-down-routers",
		Check: policy.CheckConditions{Status: "DOWN", ExemptNames: []string{"default"}},
	}

	result, err := auditor.Check(context.Background(), router, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for exempt router")
	}
}

func TestRouterAuditor_Check_ExemptNamePattern(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{
		ID:     "rtr-456",
		Name:   "ospa-e2e-router-12345",
		Status: "ACTIVE",
	}

	rule := &policy.Rule{
		Name:  "find-active-routers",
		Check: policy.CheckConditions{Status: "ACTIVE", ExemptNames: []string{"ospa-e2e-*"}},
	}

	result, err := auditor.Check(context.Background(), router, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for router matching exempt pattern ospa-e2e-*")
	}
	if result.Observation != "exempt by name" {
		t.Errorf("Check() expected observation 'exempt by name', got %q", result.Observation)
	}
}

func TestRouterAuditor_Check_Unused(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{
		ID:          "rtr-123",
		Name:        "isolated-router",
		GatewayInfo: routers.GatewayInfo{},
	}

	rule := &policy.Rule{
		Name:  "find-unused-routers",
		Check: policy.CheckConditions{Unused: true},
	}

	result, err := auditor.Check(context.Background(), router, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Compliant {
		t.Error("Check() expected non-compliant for router with no external gateway")
	}
}

func TestRouterAuditor_Check_UsedWithGateway(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{
		ID:   "rtr-123",
		Name: "connected-router",
		GatewayInfo: routers.GatewayInfo{
			NetworkID: "ext-net-456",
		},
	}

	rule := &policy.Rule{
		Name:  "find-unused-routers",
		Check: policy.CheckConditions{Unused: true},
	}

	result, err := auditor.Check(context.Background(), router, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Compliant {
		t.Error("Check() expected compliant for router with external gateway")
	}
}

func TestRouterAuditor_Check_InvalidType(t *testing.T) {
	auditor := &RouterAuditor{}

	_, err := auditor.Check(context.Background(), "not-a-router", &policy.Rule{})
	if err == nil {
		t.Error("Check() expected error for invalid resource type")
	}
}

func TestRouterAuditor_Fix_Log(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{ID: "rtr-123"}
	rule := &policy.Rule{Action: "log"}

	err := auditor.Fix(context.Background(), nil, router, rule)
	if err != nil {
		t.Errorf("Fix(log) error = %v, want nil", err)
	}
}

func TestRouterAuditor_Fix_Delete_RequiresClient(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{ID: "rtr-123"}
	rule := &policy.Rule{Action: "delete"}

	err := auditor.Fix(context.Background(), nil, router, rule)
	if err == nil {
		t.Error("Fix(delete) expected error without client")
	}
}

func TestRouterAuditor_Fix_UnsupportedAction(t *testing.T) {
	auditor := &RouterAuditor{}

	router := routers.Router{ID: "rtr-123"}
	rule := &policy.Rule{Action: "reboot"}

	err := auditor.Fix(context.Background(), nil, router, rule)
	if err == nil {
		t.Error("Fix(reboot) expected error for unsupported action")
	}
}
