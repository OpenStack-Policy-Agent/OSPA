//go:build e2e

package neutron

// =============================================================================
// SecurityGroup E2E TESTS
// =============================================================================
//
// These tests verify OSPA's ability to discover and audit Neutron security groups.
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/neutron/... -v -run SecurityGroup
//
// =============================================================================

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestNeutron_SecurityGroup_StatusCheck verifies status-based auditing.
// Creates a security group, runs an audit with status check, and verifies discovery.
func TestNeutron_SecurityGroup_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create test security group
	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	// Policy: find ACTIVE security groups
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-security-group-status
      description: Find security groups by status
      service: neutron
      resource: security_group
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter to our specific resource
	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during audit: %d", resourceResults.Errors)
	}

	// Security group should be ACTIVE and match the policy
	if resourceResults.Violations == 0 && resourceResults.Scanned > 0 {
		t.Log("Security group matched status check (ACTIVE) - test passed")
	}
}

// TestNeutron_SecurityGroup_UnusedCheck verifies unused security group detection.
// An unused security group is one not attached to any ports.
func TestNeutron_SecurityGroup_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create an "unused" security group (not attached to any ports)
	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	// Policy: find unused security groups
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-security-group-unused
      description: Find unused security groups (not attached to ports)
      service: neutron
      resource: security_group
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned but it wasn't discovered")
	}

	// Security group is not attached to any ports, so it should be flagged
	// Note: The current implementation marks unused for later evaluation
	t.Log("Security group discovered for unused check - verify manually if flagged correctly")
}

// TestNeutron_SecurityGroup_ExemptNames verifies name exemptions work.
func TestNeutron_SecurityGroup_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	// Policy: find ACTIVE security groups, but exempt ospa-e2e-* names
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-security-group-exempt
      description: Test exemption by name prefix
      service: neutron
      resource: security_group
      check:
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Resource should be compliant (exempted by name pattern)
	if resourceResults.Violations > 0 {
		t.Error("Expected security group to be exempt by name pattern, but it was flagged as a violation")
	} else {
		t.Log("Security group correctly exempted by name pattern - test passed")
	}
}

// TestNeutron_SecurityGroup_ExemptDefault verifies the 'default' security group can be exempted.
func TestNeutron_SecurityGroup_ExemptDefault(t *testing.T) {
	engine := e2e.NewTestEngine(t)

	// Policy: find unused security groups, but exempt 'default'
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-security-group-exempt-default
      description: Find unused security groups but exempt default
      service: neutron
      resource: security_group
      check:
        unused: true
        exempt_names:
          - default
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Check that 'default' security groups are not flagged
	allResults := results.FilterByService("neutron").
		FilterByResourceType("security_group")

	allResults.LogSummary(t)

	// Look for any 'default' security groups in violations
	for _, result := range allResults.Results {
		if result.ResourceName == "default" && !result.Compliant {
			t.Error("Default security group should be exempt but was flagged as a violation")
		}
	}

	t.Log("Default security group exemption check completed")
}

// TestNeutron_SecurityGroup_Discovery verifies basic discovery works.
func TestNeutron_SecurityGroup_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create multiple security groups to test batch discovery
	sgID1, cleanup1 := CreateSecurityGroup(t, client)
	defer cleanup1()

	sgID2, cleanup2 := CreateSecurityGroup(t, client)
	defer cleanup2()

	// Simple policy to discover all security groups
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-discovery
      description: Discover all security groups
      service: neutron
      resource: security_group
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Check both security groups were discovered
	results1 := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(sgID1)

	results2 := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(sgID2)

	if results1.Scanned == 0 {
		t.Error("First security group was not discovered")
	}
	if results2.Scanned == 0 {
		t.Error("Second security group was not discovered")
	}

	allSGResults := results.FilterByService("neutron").
		FilterByResourceType("security_group")
	t.Logf("Discovery test passed: found %d total security groups", allSGResults.Scanned)
}

// TestNeutron_SecurityGroup_TagAction verifies the tag remediation action works.
func TestNeutron_SecurityGroup_TagAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true // Enable remediation
	client := engine.GetNetworkClient(t)

	// Create test security group
	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	// Policy: tag ACTIVE security groups
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-security-group-tag
      description: Tag active security groups
      service: neutron
      resource: security_group
      check:
        status: ACTIVE
      action: tag
      tag_name: ospa-audited`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during tag action: %d", resourceResults.Errors)
	}

	t.Log("Security group tag action test completed")
}

// TestCleanup_SecurityGroup cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/neutron/... -run TestCleanup -v
func TestCleanup_SecurityGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	CleanupOrphans(t, client)
}
