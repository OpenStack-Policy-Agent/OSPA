//go:build e2e

package cinder

// =============================================================================
// Snapshot E2E TESTS
// =============================================================================
//
// BEFORE WRITING TESTS:
// 1. Implement CreateSnapshot() in resource_creator.go
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
// - [ ] Domain check: encrypted (Snapshot is not encrypted)
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/cinder/... -v -run Snapshot
//
// =============================================================================

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestCinder_Snapshot_StatusCheck verifies status-based auditing.
func TestCinder_Snapshot_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-status
      description: Find snapshot by status
      resource: snapshot
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
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

// TestCinder_Snapshot_AgeGTCheck verifies age-based auditing.
func TestCinder_Snapshot_AgeGTCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-age
      description: Find snapshot older than 30 days
      resource: snapshot
      check:
        age_gt: 30d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	// Freshly created resource should be compliant (younger than 30 days)
	if resourceResults.Violations > 0 {
		t.Error("Freshly created resource should not be flagged by age_gt: 30d")
	}

	// TODO: Add an AgeGTViolation test: create resource, sleep briefly, audit with age_gt: 1s, expect violation
}

// TestCinder_Snapshot_UnusedCheck verifies unused detection.
func TestCinder_Snapshot_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-unused
      description: Find unused snapshot
      resource: snapshot
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// TODO: Assert based on resource-specific unused semantics
}

// TestCinder_Snapshot_ExemptNames verifies name exemptions work.
func TestCinder_Snapshot_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-exempt
      description: Test exemption by name prefix
      resource: snapshot
      check:
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	if resourceResults.Violations > 0 {
		t.Error("Expected resource to be exempt by name")
	}
}

// TestCinder_Snapshot_Discovery verifies batch discovery.
func TestCinder_Snapshot_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	id1, cleanup1 := CreateSnapshot(t, client)
	defer cleanup1()
	id2, cleanup2 := CreateSnapshot(t, client)
	defer cleanup2()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-discovery
      description: Discover snapshot resources
      resource: snapshot
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	r1 := results.FilterByService("cinder").FilterByResourceType("snapshot").FilterByResourceID(id1)
	r2 := results.FilterByService("cinder").FilterByResourceType("snapshot").FilterByResourceID(id2)

	if r1.Scanned == 0 {
		t.Error("First resource was not discovered")
	}
	if r2.Scanned == 0 {
		t.Error("Second resource was not discovered")
	}
}

// TestCinder_Snapshot_Classification verifies severity/category/guide_ref propagation.
func TestCinder_Snapshot_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-classify
      description: Test classification fields
      resource: snapshot
      check:
        status: ACTIVE
      action: log
      severity: high
      category: security
      guide_ref: TEST-001`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	resourceResults.AssertClassification(t, "high", "security", "TEST-001")

	// TODO: Replace TEST-001 with a real OpenStack Security Guide reference for this resource
}

// TestCinder_Snapshot_OutputJSON verifies JSON output contains required fields.
func TestCinder_Snapshot_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	_, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-json
      description: Output format test
      resource: snapshot
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

// TestCinder_Snapshot_OutputCSV verifies CSV output has the correct headers.
func TestCinder_Snapshot_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	_, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-csv
      description: Output format test
      resource: snapshot
      check:
        status: ACTIVE
      action: log
      severity: low
      category: operations`

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

// TestCinder_Snapshot_DeleteAction verifies delete remediation.
func TestCinder_Snapshot_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-delete
      description: Delete snapshot resources
      resource: snapshot
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}

	// TODO: Verify the resource was actually deleted (e.g. GET returns 404)
}

// TestCinder_Snapshot_TagAction verifies tag remediation.
func TestCinder_Snapshot_TagAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-tag
      description: Tag snapshot resources
      resource: snapshot
      check:
        status: ACTIVE
      action: tag
      tag_name: ospa-e2e-tagged`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during tag: %d", resourceResults.Errors)
	}

	// TODO: Verify the tag was actually applied (e.g. GET resource, check tags list)
}

// TestCinder_Snapshot_DryRunSkip verifies dry-run skips remediation.
func TestCinder_Snapshot_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-dryrun
      description: Dry run delete
      resource: snapshot
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)
	resourceResults.AssertRemediationSkipped(t, "dry-run")

	// TODO: Verify the resource still exists after dry-run (e.g. GET returns 200)
}

// TestCinder_Snapshot_AllowActionsFiltering verifies the allow-actions filter.
func TestCinder_Snapshot_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetBlockStorageClient(t)

	resourceID, cleanup := CreateSnapshot(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-allowlist
      description: Delete not in allowlist
      resource: snapshot
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("cinder").
		FilterByResourceType("snapshot").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestCinder_Snapshot_Check_Encrypted tests the encrypted domain check.
func TestCinder_Snapshot_Check_Encrypted(t *testing.T) {
	// TODO: Implement domain-specific check test for encrypted
	// Description: Snapshot is not encrypted
	// Category: security, Severity: high
	t.Skip("Domain check test for encrypted not yet implemented")
}

// TestCleanup_Snapshot cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/cinder/... -run TestCleanup
func TestCleanup_Snapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetBlockStorageClient(t)

	CleanupOrphans(t, client)
}
