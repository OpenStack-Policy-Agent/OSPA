//go:build e2e

package neutron

// =============================================================================
// Subnet E2E TESTS
// =============================================================================
//
// These tests verify OSPA's ability to discover and audit Neutron subnets.
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/neutron/... -v -run Subnet
//
// =============================================================================

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestNeutron_Subnet_Discovery verifies that subnets are discovered and audited.
func TestNeutron_Subnet_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-discovery
      description: Discover all subnets
      service: neutron
      resource: subnet
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected subnet to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during audit: %d", resourceResults.Errors)
	}

	// CreateSubnet creates a subnet with allocation pools, so it should be compliant
	if resourceResults.Violations > 0 {
		t.Error("Subnet with allocation pools was incorrectly flagged as unused")
	} else {
		t.Log("Subnet correctly identified as used (has allocation pools) - test passed")
	}
}

// TestNeutron_Subnet_ExemptNames verifies name exemptions work.
func TestNeutron_Subnet_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-exempt
      description: Test exemption by name prefix
      service: neutron
      resource: subnet
      check:
        unused: true
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected subnet to be scanned but it wasn't discovered")
	}

	if resourceResults.Violations > 0 {
		t.Error("Expected subnet to be exempt by name pattern, but it was flagged as a violation")
	} else {
		t.Log("Subnet correctly exempted by name pattern - test passed")
	}
}

// TestNeutron_Subnet_MultipleDiscovery verifies batch discovery of subnets.
func TestNeutron_Subnet_MultipleDiscovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	subnetID1, cleanup1 := CreateSubnet(t, client)
	defer cleanup1()

	subnetID2, cleanup2 := CreateSubnet(t, client)
	defer cleanup2()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-batch
      description: Discover multiple subnets
      service: neutron
      resource: subnet
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	results1 := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID1)

	results2 := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID2)

	if results1.Scanned == 0 {
		t.Error("First subnet was not discovered")
	}
	if results2.Scanned == 0 {
		t.Error("Second subnet was not discovered")
	}

	t.Logf("Batch discovery test passed: both subnets found")
}

// TestCleanup_Subnet cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/neutron/... -run TestCleanup -v
func TestCleanup_Subnet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	CleanupOrphans(t, client)
}
