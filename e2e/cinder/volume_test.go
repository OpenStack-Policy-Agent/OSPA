//go:build e2e

package cinder

// =============================================================================
// Volume E2E TESTS
// =============================================================================
//
// BEFORE WRITING TESTS:
// 1. Implement CreateVolume() in resource_creator.go
// 2. The creator should handle all dependencies (network, subnet, etc.)
// 3. The creator returns a cleanup function - always defer it!
//
// TEST COVERAGE CHECKLIST:
// - [x] Status check (status: ACTIVE, DOWN, ERROR, etc.)
// - [x] Age check (age_gt: 30d)
// - [x] Unused check (unused: true) - if applicable
// - [x] Exempt names (exempt_names: [...])
// - [x] Discovery (multiple resources)
// - [x] Classification propagation (severity/category/guide_ref)
// - [x] Output JSON
// - [x] Output CSV
// - [x] Delete action
// - [x] Tag action
// - [x] Dry-run remediation skip
// - [x] Allow-actions filtering
// - [ ] Domain check: encrypted (Volume is not encrypted)
// - [ ] Domain check: attached (Volume is not attached to any instance)
// - [ ] Domain check: has_backup (Volume has no backup)
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/cinder/... -v -run Volume
//
// =============================================================================

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestCinder_Volume_StatusCheck verifies status-based auditing.
func TestCinder_Volume_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-status
      description: Find volume by status
      resource: volume
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors: %d", resourceResults.Errors)
	}

	// TODO: Adjust expected status for this resource type (e.g. "available", "in-use", "BUILD")
	// TODO: Add a non-compliant status test (create resource, check for a status it does NOT have)
}

// TestCinder_Volume_AgeGTCheck verifies age-based auditing.
func TestCinder_Volume_AgeGTCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-age
      description: Find volume older than 30 days
      resource: volume
      check:
        age_gt: 30d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	// Freshly created resource should be compliant (younger than 30 days)
	if resourceResults.Violations > 0 {
		t.Error("Freshly created resource should not be flagged by age_gt: 30d")
	}

	// TODO: Add an AgeGTViolation test: create resource, audit with age_gt: 0m, expect violation
}

// TestCinder_Volume_UnusedCheck verifies unused detection.
func TestCinder_Volume_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-unused
      description: Find unused volume
      resource: volume
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// TODO: Assert based on resource-specific unused semantics
}

// TestCinder_Volume_ExemptNames verifies name exemptions work.
func TestCinder_Volume_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-exempt
      description: Test exemption by name prefix
      resource: volume
      check:
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	if resourceResults.Violations > 0 {
		t.Error("Expected resource to be exempt by name")
	}
}

// TestCinder_Volume_Discovery verifies batch discovery.
func TestCinder_Volume_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	id1, cleanup1 := CreateVolume(t, client)
	defer cleanup1()
	id2, cleanup2 := CreateVolume(t, client)
	defer cleanup2()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-discovery
      description: Discover volume resources
      resource: volume
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	r1 := results.FilterByService("cinder").FilterByResourceType("volume").FilterByResourceID(id1)
	r2 := results.FilterByService("cinder").FilterByResourceType("volume").FilterByResourceID(id2)

	if r1.Scanned == 0 {
		t.Error("First resource was not discovered")
	}
	if r2.Scanned == 0 {
		t.Error("Second resource was not discovered")
	}
}

// TestCinder_Volume_Classification verifies severity/category/guide_ref propagation.
func TestCinder_Volume_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-classify
      description: Test classification fields
      resource: volume
      check:
        status: ACTIVE
      action: log
      severity: high
      category: security
      guide_ref: TEST-001`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	resourceResults.AssertClassification(t, "high", "security", "TEST-001")

	// TODO: Replace TEST-001 with a real OpenStack Security Guide reference for this resource
}

// TestCinder_Volume_OutputJSON verifies JSON output contains required fields.
func TestCinder_Volume_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	_, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-json
      description: Output format test
      resource: volume
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

// TestCinder_Volume_OutputCSV verifies CSV output has the correct headers.
func TestCinder_Volume_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	_, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-csv
      description: Output format test
      resource: volume
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

// TestCinder_Volume_DeleteAction verifies delete remediation.
func TestCinder_Volume_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-delete
      description: Delete volume resources
      resource: volume
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}

	// TODO: Verify the resource was actually deleted (e.g. GET returns 404)
}

// TestCinder_Volume_TagAction verifies tag remediation.
func TestCinder_Volume_TagAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-tag
      description: Tag volume resources
      resource: volume
      check:
        status: ACTIVE
      action: tag
      tag_name: ospa-e2e-tagged`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during tag: %d", resourceResults.Errors)
	}

	// TODO: Verify the tag was actually applied (e.g. GET resource, check tags list)
}

// TestCinder_Volume_DryRunSkip verifies dry-run skips remediation.
func TestCinder_Volume_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-dryrun
      description: Dry run delete
      resource: volume
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)
	resourceResults.AssertRemediationSkipped(t, "dry-run")

	// TODO: Verify the resource still exists after dry-run (e.g. GET returns 200)
}

// TestCinder_Volume_AllowActionsFiltering verifies the allow-actions filter.
func TestCinder_Volume_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateVolume(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-allowlist
      description: Delete not in allowlist
      resource: volume
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("volume").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestCinder_Volume_Check_Encrypted tests the encrypted domain check.
func TestCinder_Volume_Check_Encrypted(t *testing.T) {
	// TODO: Implement domain-specific check test for encrypted
	// Description: Volume is not encrypted
	// Category: security, Severity: high
	t.Skip("Domain check test for encrypted not yet implemented")
}

// TestCinder_Volume_Check_Attached tests the attached domain check.
func TestCinder_Volume_Check_Attached(t *testing.T) {
	// TODO: Implement domain-specific check test for attached
	// Description: Volume is not attached to any instance
	// Category: cost, Severity: medium
	t.Skip("Domain check test for attached not yet implemented")
}

// TestCinder_Volume_Check_HasBackup tests the has_backup domain check.
func TestCinder_Volume_Check_HasBackup(t *testing.T) {
	// TODO: Implement domain-specific check test for has_backup
	// Description: Volume has no backup
	// Category: compliance, Severity: medium
	t.Skip("Domain check test for has_backup not yet implemented")
}

// TestCleanup_Volume cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/cinder/... -run TestCleanup
func TestCleanup_Volume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	CleanupOrphans(t, client)
}
