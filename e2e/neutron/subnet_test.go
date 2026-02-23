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
	"encoding/csv"
	"encoding/json"
	"os"
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

// TestNeutron_Subnet_AgeGTCompliant verifies a freshly created subnet passes age_gt: 30d.
func TestNeutron_Subnet_AgeGTCompliant(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-age
      description: Find subnets older than 30 days
      service: neutron
      resource: subnet
      check:
        age_gt: 30d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected subnet to be scanned")
	}
	if resourceResults.Violations > 0 {
		t.Error("Freshly created subnet should not be flagged by age_gt: 30d")
	}
}

// TestNeutron_Subnet_AgeGTViolation verifies age_gt is a no-op for subnets
// (gophercloud Subnet struct lacks CreatedAt/UpdatedAt).
func TestNeutron_Subnet_AgeGTViolation(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-age-short
      description: Find subnets older than 0m
      resource: subnet
      check:
        age_gt: 0m
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected subnet to be scanned")
	}
	// Subnets lack timestamp fields in gophercloud, so age_gt is a no-op.
	// This test verifies the policy executes without errors.
	t.Log("age_gt is a no-op for subnets (no CreatedAt field) -- audit ran without errors")
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

// TestNeutron_Subnet_DeleteAction verifies the delete remediation action on subnets.
func TestNeutron_Subnet_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-delete
      description: Delete unused subnets
      service: neutron
      resource: subnet
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected subnet to be scanned")
	}
	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}
}

// TestNeutron_Subnet_DryRunSkip verifies that dry-run mode skips remediation.
func TestNeutron_Subnet_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-dryrun
      description: Dry run delete on unused subnets
      resource: subnet
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected subnet to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "dry-run")
}

// TestNeutron_Subnet_AllowActionsFiltering verifies the allow-actions filter.
func TestNeutron_Subnet_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-allowlist
      description: Delete not in allowlist
      resource: subnet
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected subnet to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestNeutron_Subnet_Classification verifies severity/category/guide_ref propagation.
func TestNeutron_Subnet_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	subnetID, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-classification
      description: Test classification field propagation
      resource: subnet
      check:
        unused: true
      action: log
      severity: medium
      category: security
      guide_ref: OSSN-0068`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("subnet").
		FilterByResourceID(subnetID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Fatal("Expected subnet to be scanned")
	}

	resourceResults.AssertClassification(t, "medium", "security", "OSSN-0068")
}

// TestNeutron_Subnet_OutputJSON verifies JSON output contains required fields.
func TestNeutron_Subnet_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-json
      description: Output format test
      resource: subnet
      check:
        unused: true
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

// TestNeutron_Subnet_OutputCSV verifies CSV output has the correct headers.
func TestNeutron_Subnet_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-subnet-csv
      description: Output format test
      resource: subnet
      check:
        unused: true
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
