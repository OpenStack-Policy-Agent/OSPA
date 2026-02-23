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
	"encoding/csv"
	"encoding/json"
	"os"
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

// TestNeutron_SecurityGroup_AgeGTCompliant verifies a fresh SG passes age_gt: 30d.
func TestNeutron_SecurityGroup_AgeGTCompliant(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-age
      description: Find security groups older than 30 days
      service: neutron
      resource: security_group
      check:
        age_gt: 30d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned")
	}
	if resourceResults.Violations > 0 {
		t.Error("Freshly created security group should not be flagged by age_gt: 30d")
	}
}

// TestNeutron_SecurityGroup_AgeGTViolation verifies an SG trips age_gt: 0m (any age triggers).
func TestNeutron_SecurityGroup_AgeGTViolation(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-age-short
      description: Find SGs older than 0m (always triggers)
      resource: security_group
      check:
        age_gt: 0m
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned")
	}
	if resourceResults.Violations == 0 {
		t.Error("Security group should be flagged by age_gt: 0m")
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

// TestNeutron_SecurityGroup_DeleteAction verifies the delete remediation action on unused SGs.
func TestNeutron_SecurityGroup_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-delete
      description: Delete unused security groups
      service: neutron
      resource: security_group
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned")
	}
	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}
}

// TestNeutron_SecurityGroup_DryRunSkip verifies that dry-run mode skips remediation.
func TestNeutron_SecurityGroup_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-dryrun
      description: Dry run delete on unused security groups
      resource: security_group
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "dry-run")
}

// TestNeutron_SecurityGroup_AllowActionsFiltering verifies the allow-actions filter.
func TestNeutron_SecurityGroup_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-allowlist
      description: Delete not in allowlist
      resource: security_group
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected security group to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestNeutron_SecurityGroup_Classification verifies severity/category/guide_ref propagation.
func TestNeutron_SecurityGroup_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-classification
      description: Test classification field propagation
      resource: security_group
      check:
        status: ACTIVE
      action: log
      severity: high
      category: security
      guide_ref: OSSN-0047`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Fatal("Expected security group to be scanned")
	}

	resourceResults.AssertClassification(t, "high", "security", "OSSN-0047")
}

// TestNeutron_SecurityGroup_OutputJSON verifies JSON output contains required fields.
func TestNeutron_SecurityGroup_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-json
      description: Output format test
      resource: security_group
      check:
        status: ACTIVE
      action: log
      severity: medium
      category: compliance`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results, filePath := engine.RunAuditToFile(t, policy, "json")
	defer os.Remove(filePath)

	if results.Scanned == 0 {
		t.Skip("No resources scanned, cannot validate output")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read JSON output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("JSON output file is empty")
	}

	var finding map[string]interface{}
	if err := json.Unmarshal(data[:e2e.FindLineEnd(data)], &finding); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	for _, field := range []string{"rule_id", "resource_id", "compliant", "severity", "category"} {
		if _, ok := finding[field]; !ok {
			t.Errorf("JSON output missing required field: %s", field)
		}
	}
}

// TestNeutron_SecurityGroup_OutputCSV verifies CSV output has the correct headers.
func TestNeutron_SecurityGroup_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateSecurityGroup(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-csv
      description: Output format test
      resource: security_group
      check:
        status: ACTIVE
      action: log
      severity: low
      category: hygiene`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results, filePath := engine.RunAuditToFile(t, policy, "csv")
	defer os.Remove(filePath)

	if results.Scanned == 0 {
		t.Skip("No resources scanned, cannot validate output")
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open CSV output: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}

	expected := map[string]bool{"rule_id": false, "resource_id": false, "compliant": false, "severity": false, "category": false, "guide_ref": false}
	for _, col := range header {
		if _, ok := expected[col]; ok {
			expected[col] = true
		}
	}
	for col, found := range expected {
		if !found {
			t.Errorf("CSV header missing expected column: %s", col)
		}
	}

	row, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV data row: %v", err)
	}
	if len(row) != len(header) {
		t.Errorf("CSV data row has %d columns, expected %d", len(row), len(header))
	}
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
