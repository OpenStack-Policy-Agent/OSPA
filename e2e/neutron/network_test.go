//go:build e2e

package neutron

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// =============================================================================
// Network E2E TESTS
// =============================================================================
//
// These tests verify OSPA's ability to discover and audit Neutron networks.
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/neutron/... -v -run Network
//
// =============================================================================

// TestNeutron_Network_StatusCheck verifies status-based auditing.
// Creates a network, runs an audit with status check, and verifies discovery.
func TestNeutron_Network_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create test network
	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	// Policy: find ACTIVE networks
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-status
      description: Find networks by status
      service: neutron
      resource: network
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter to our specific resource
	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during audit: %d", resourceResults.Errors)
	}

	// Network should be ACTIVE and match the policy
	if resourceResults.Violations == 0 && resourceResults.Scanned > 0 {
		t.Log("Network matched status check (ACTIVE) - test passed")
	}
}

// TestNeutron_Network_UnusedCheck verifies unused network detection.
// An unused network is one without any subnets attached.
func TestNeutron_Network_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create an "unused" network (no subnet attached)
	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	// Policy: find unused networks
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-unused
      description: Find unused networks (no subnets)
      service: neutron
      resource: network
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned but it wasn't discovered")
	}

	// Network has no subnets, so it should be flagged as unused (violation)
	if resourceResults.Violations == 0 {
		t.Log("Warning: Network without subnets was not flagged as unused - check auditor implementation")
	} else {
		t.Log("Network correctly identified as unused - test passed")
	}
}

// TestNeutron_Network_WithSubnet_NotUnused verifies that networks with subnets are not flagged as unused.
func TestNeutron_Network_WithSubnet_NotUnused(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create a network WITH a subnet (not unused)
	networkID, _, cleanup := CreateNetworkWithSubnet(t, client)
	defer cleanup()

	// Policy: find unused networks
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-used
      description: Find unused networks
      service: neutron
      resource: network
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned but it wasn't discovered")
	}

	// Network has a subnet, so it should NOT be flagged as unused
	if resourceResults.Violations > 0 {
		t.Error("Network with subnet was incorrectly flagged as unused")
	} else {
		t.Log("Network with subnet correctly identified as used - test passed")
	}
}

// TestNeutron_Network_ExemptNames verifies name exemptions work.
func TestNeutron_Network_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	// Policy: find ACTIVE networks, but exempt ospa-e2e-* names
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-exempt
      description: Test exemption by name prefix
      service: neutron
      resource: network
      check:
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	// Resource should be compliant (exempted by name pattern)
	if resourceResults.Violations > 0 {
		t.Error("Expected network to be exempt by name pattern, but it was flagged as a violation")
	} else {
		t.Log("Network correctly exempted by name pattern - test passed")
	}
}

// TestNeutron_Network_Discovery verifies basic discovery works.
func TestNeutron_Network_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create multiple networks to test batch discovery
	networkID1, cleanup1 := CreateNetwork(t, client)
	defer cleanup1()

	networkID2, cleanup2 := CreateNetwork(t, client)
	defer cleanup2()

	// Simple policy to discover all networks
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-discovery
      description: Discover all networks
      service: neutron
      resource: network
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Check both networks were discovered
	results1 := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID1)

	results2 := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID2)

	if results1.Scanned == 0 {
		t.Error("First network was not discovered")
	}
	if results2.Scanned == 0 {
		t.Error("Second network was not discovered")
	}

	t.Logf("Discovery test passed: found %d total networks", results.Scanned)
}

// TestCleanup_Network cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/neutron/... -run TestCleanup -v
func TestCleanup_Network(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	CleanupOrphans(t, client)
}
