//go:build e2e

package neutron

// =============================================================================
// Router E2E TESTS
// =============================================================================
//
// These tests verify OSPA's ability to discover and audit Neutron routers.
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/neutron/... -v -run Router
//
// =============================================================================

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestNeutron_Router_StatusCheck verifies status-based auditing.
func TestNeutron_Router_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	routerID, cleanup := CreateRouter(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-router-status
      description: Find routers with ACTIVE status
      service: neutron
      resource: router
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("router").
		FilterByResourceID(routerID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected router to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during audit: %d", resourceResults.Errors)
	}

	// A freshly created router is ACTIVE, so it should be flagged as non-compliant
	if resourceResults.Violations == 0 {
		t.Error("Expected ACTIVE router to be flagged by status: ACTIVE check")
	} else {
		t.Log("Router correctly flagged by status check - test passed")
	}
}

// TestNeutron_Router_UnusedCheck verifies unused detection (no external gateway).
func TestNeutron_Router_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// CreateRouter creates a router without an external gateway
	routerID, cleanup := CreateRouter(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-router-unused
      description: Find routers without external gateway
      service: neutron
      resource: router
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("router").
		FilterByResourceID(routerID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected router to be scanned but it wasn't discovered")
	}

	// Router has no external gateway, so it should be flagged as unused
	if resourceResults.Violations == 0 {
		t.Error("Expected router without external gateway to be flagged as unused")
	} else {
		t.Log("Router correctly flagged as unused (no external gateway) - test passed")
	}
}

// TestNeutron_Router_ExemptNames verifies name exemptions work.
func TestNeutron_Router_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	routerID, cleanup := CreateRouter(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-router-exempt
      description: Test exemption by name prefix
      service: neutron
      resource: router
      check:
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("router").
		FilterByResourceID(routerID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected router to be scanned but it wasn't discovered")
	}

	if resourceResults.Violations > 0 {
		t.Error("Expected router to be exempt by name pattern, but it was flagged")
	} else {
		t.Log("Router correctly exempted by name pattern - test passed")
	}
}

// TestNeutron_Router_MultipleDiscovery verifies batch discovery of routers.
func TestNeutron_Router_MultipleDiscovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	routerID1, cleanup1 := CreateRouter(t, client)
	defer cleanup1()

	routerID2, cleanup2 := CreateRouter(t, client)
	defer cleanup2()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-router-batch
      description: Discover multiple routers
      service: neutron
      resource: router
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	results1 := results.FilterByService("neutron").
		FilterByResourceType("router").
		FilterByResourceID(routerID1)

	results2 := results.FilterByService("neutron").
		FilterByResourceType("router").
		FilterByResourceID(routerID2)

	if results1.Scanned == 0 {
		t.Error("First router was not discovered")
	}
	if results2.Scanned == 0 {
		t.Error("Second router was not discovered")
	}

	t.Log("Batch discovery test passed: both routers found")
}

// TestCleanup_Router cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/neutron/... -run TestCleanup -v
func TestCleanup_Router(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	CleanupOrphans(t, client)
}
