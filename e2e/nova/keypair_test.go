//go:build e2e

package nova

// =============================================================================
// Keypair E2E TESTS
// =============================================================================
//
// BEFORE WRITING TESTS:
// 1. Implement CreateKeypair() in resource_creator.go
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
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/nova/... -v -run Keypair
//
// =============================================================================

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestNova_Keypair_StatusCheck verifies status-based auditing.
func TestNova_Keypair_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-status
      description: Find keypair by status
      resource: keypair
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
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

// TestNova_Keypair_AgeGTCheck verifies age-based auditing.
func TestNova_Keypair_AgeGTCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-age
      description: Find keypair older than 30 days
      resource: keypair
      check:
        age_gt: 30d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
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

// TestNova_Keypair_UnusedCheck verifies unused detection.
func TestNova_Keypair_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-unused
      description: Find unused keypair
      resource: keypair
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// TODO: Assert based on resource-specific unused semantics
}

// TestNova_Keypair_ExemptNames verifies name exemptions work.
func TestNova_Keypair_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-exempt
      description: Test exemption by name prefix
      resource: keypair
      check:
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
		FilterByResourceID(resourceID)

	if resourceResults.Violations > 0 {
		t.Error("Expected resource to be exempt by name")
	}
}

// TestNova_Keypair_Discovery verifies batch discovery.
func TestNova_Keypair_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	id1, cleanup1 := CreateKeypair(t, client)
	defer cleanup1()
	id2, cleanup2 := CreateKeypair(t, client)
	defer cleanup2()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-discovery
      description: Discover keypair resources
      resource: keypair
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	r1 := results.FilterByService("nova").FilterByResourceType("keypair").FilterByResourceID(id1)
	r2 := results.FilterByService("nova").FilterByResourceType("keypair").FilterByResourceID(id2)

	if r1.Scanned == 0 {
		t.Error("First resource was not discovered")
	}
	if r2.Scanned == 0 {
		t.Error("Second resource was not discovered")
	}
}

// TestNova_Keypair_Classification verifies severity/category/guide_ref propagation.
func TestNova_Keypair_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-classify
      description: Test classification fields
      resource: keypair
      check:
        status: ACTIVE
      action: log
      severity: high
      category: security
      guide_ref: TEST-001`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
		FilterByResourceID(resourceID)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	resourceResults.AssertClassification(t, "high", "security", "TEST-001")

	// TODO: Replace TEST-001 with a real OpenStack Security Guide reference for this resource
}

// TestNova_Keypair_OutputJSON verifies JSON output contains required fields.
func TestNova_Keypair_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	_, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-json
      description: Output format test
      resource: keypair
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

// TestNova_Keypair_OutputCSV verifies CSV output has the correct headers.
func TestNova_Keypair_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	_, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-csv
      description: Output format test
      resource: keypair
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

// TestNova_Keypair_DeleteAction verifies delete remediation.
func TestNova_Keypair_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-delete
      description: Delete keypair resources
      resource: keypair
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}

	// TODO: Verify the resource was actually deleted (e.g. GET returns 404)
}

// TestNova_Keypair_TagAction verifies tag remediation.
func TestNova_Keypair_TagAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-tag
      description: Tag keypair resources
      resource: keypair
      check:
        status: ACTIVE
      action: tag
      tag_name: ospa-e2e-tagged`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during tag: %d", resourceResults.Errors)
	}

	// TODO: Verify the tag was actually applied (e.g. GET resource, check tags list)
}

// TestNova_Keypair_DryRunSkip verifies dry-run skips remediation.
func TestNova_Keypair_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-dryrun
      description: Dry run delete
      resource: keypair
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)
	resourceResults.AssertRemediationSkipped(t, "dry-run")

	// TODO: Verify the resource still exists after dry-run (e.g. GET returns 200)
}

// TestNova_Keypair_AllowActionsFiltering verifies the allow-actions filter.
func TestNova_Keypair_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetComputeClient(t)

	resourceID, cleanup := CreateKeypair(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-allowlist
      description: Delete not in allowlist
      resource: keypair
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("nova").
		FilterByResourceType("keypair").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestCleanup_Keypair cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/nova/... -run TestCleanup
func TestCleanup_Keypair(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetComputeClient(t)

	CleanupOrphans(t, client)
}
