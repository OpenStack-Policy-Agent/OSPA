//go:build e2e

package neutron

// =============================================================================
// Subnet E2E TESTS
// =============================================================================
//
// BEFORE WRITING TESTS:
// 1. Implement CreateSubnet() in resource_creator.go
// 2. The creator should handle all dependencies (network, subnet, etc.)
// 3. The creator returns a cleanup function - always defer it!
//
// TEST COVERAGE CHECKLIST:
// - [ ] Status check (status: ACTIVE, DOWN, ERROR, etc.)
// - [ ] Age check (age_gt: 30d)
// - [ ] Unused check (unused: true) - if applicable
// - [ ] Exempt names (exempt_names: [...])
// - [ ] Multiple resources (batch discovery)
// - [ ] Error handling (invalid resource, missing permissions)
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/neutron/... -v -run Subnet
//
// =============================================================================

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestNeutron_Subnet_StatusCheck verifies status-based auditing.
func TestNeutron_Subnet_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create test resource using the helper from resource_creator.go
	resourceID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	// Run audit with status check
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-status
      description: Find subnet by status
      service: neutron
      resource: subnet
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter to our specific resource
	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors: %d", resourceResults.Errors)
	}
}

// TestNeutron_Subnet_UnusedCheck verifies unused detection.
func TestNeutron_Subnet_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create an "unused" resource (no attachments/dependencies)
	resourceID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-unused
      description: Find unused subnet
      service: neutron
      resource: subnet
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// TODO: Add assertions based on whether the resource should be flagged
}

// TestNeutron_Subnet_ExemptNames verifies name exemptions work.
func TestNeutron_Subnet_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	// The resource name starts with "ospa-e2e-" - exempt it
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
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(resourceID)

	// Resource should be compliant (exempted by name)
	if resourceResults.Violations > 0 {
		t.Error("Expected resource to be exempt by name")
	}
}

// TestCleanup_Subnet cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/neutron/... -run TestCleanup
func TestCleanup_Subnet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	CleanupOrphans(t, client)
}
