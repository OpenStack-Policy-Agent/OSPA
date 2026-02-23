//go:build e2e

package neutron

// =============================================================================
// Port E2E TESTS
// =============================================================================
//
// These tests verify OSPA's ability to discover and audit Neutron ports.
//
// TEST COVERAGE CHECKLIST:
// - [x] Status check (status: DOWN for unattached port)
// - [x] Age check (age_gt: 30d)
// - [x] Unused check (unused: true - port not attached to device)
// - [x] No security group check (no_security_group: true)
// - [x] Exempt names (exempt_names: [...])
// - [x] Discovery (multiple resources)
// - [x] Classification propagation (severity/category/guide_ref)
// - [x] Output JSON
// - [x] Output CSV
// - [x] Delete action
// - [x] Dry-run remediation skip
// - [x] Allow-actions filtering
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/neutron/... -v -run Port
//
// =============================================================================

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestNeutron_Port_StatusCheck verifies status-based auditing.
func TestNeutron_Port_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-status
      description: Find ports with DOWN status
      resource: port
      check:
        status: DOWN
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected port to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during audit: %d", resourceResults.Errors)
	}

	// An unattached port is DOWN, so status: DOWN should flag it
	if resourceResults.Violations == 0 {
		t.Error("Expected DOWN port to be flagged by status: DOWN check")
	} else {
		t.Log("Port correctly flagged by status check - test passed")
	}
}

// TestNeutron_Port_AgeGTCheck verifies age-based auditing.
func TestNeutron_Port_AgeGTCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-age
      description: Find ports older than 30 days
      resource: port
      check:
        age_gt: 30d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected port to be scanned")
	}
	// Freshly created port should be compliant
	if resourceResults.Violations > 0 {
		t.Error("Freshly created port should not be flagged by age_gt: 30d")
	}
}

// TestNeutron_Port_UnusedCheck verifies unused detection (no device attached).
func TestNeutron_Port_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-unused
      description: Find ports not attached to any device
      resource: port
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected port to be scanned")
	}

	// CreatePort creates a port without a device, so it should be flagged
	if resourceResults.Violations == 0 {
		t.Error("Unattached port should be flagged as unused")
	} else {
		t.Log("Port correctly flagged as unused (no device) - test passed")
	}
}

// TestNeutron_Port_NoSecurityGroupCheck verifies no_security_group detection.
func TestNeutron_Port_NoSecurityGroupCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-no-sg
      description: Find ports with no security groups
      resource: port
      check:
        no_security_group: true
      action: log
      severity: high
      category: security`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected port to be scanned")
	}

	// CreatePort creates a port with empty security groups
	if resourceResults.Violations == 0 {
		t.Error("Port with no security groups should be flagged")
	} else {
		t.Log("Port correctly flagged as having no security groups - test passed")
	}
}

// TestNeutron_Port_ExemptNames verifies name exemptions work.
func TestNeutron_Port_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-exempt
      description: Test exemption by name prefix
      resource: port
      check:
        unused: true
        exempt_names:
          - ospa-e2e-*
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected port to be scanned")
	}

	if resourceResults.Violations > 0 {
		t.Error("Expected port to be exempt by name pattern, but it was flagged")
	} else {
		t.Log("Port correctly exempted by name pattern - test passed")
	}
}

// TestNeutron_Port_MultipleDiscovery verifies batch discovery of ports.
func TestNeutron_Port_MultipleDiscovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	id1, cleanup1 := CreatePort(t, client)
	defer cleanup1()
	id2, cleanup2 := CreatePort(t, client)
	defer cleanup2()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-discovery
      description: Discover multiple ports
      resource: port
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	r1 := results.FilterByService("neutron").FilterByResourceType("port").FilterByResourceID(id1)
	r2 := results.FilterByService("neutron").FilterByResourceType("port").FilterByResourceID(id2)

	if r1.Scanned == 0 {
		t.Error("First port was not discovered")
	}
	if r2.Scanned == 0 {
		t.Error("Second port was not discovered")
	}

	t.Log("Batch discovery test passed: both ports found")
}

// TestNeutron_Port_Classification verifies severity/category/guide_ref propagation.
func TestNeutron_Port_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-classify
      description: Test classification fields
      resource: port
      check:
        unused: true
      action: log
      severity: high
      category: cost
      guide_ref: OSPA-PORT-001`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	if resourceResults.Scanned == 0 {
		t.Error("Expected port to be scanned")
	}
	resourceResults.AssertClassification(t, "high", "cost", "OSPA-PORT-001")
}

// TestNeutron_Port_OutputJSON verifies JSON output contains required fields.
func TestNeutron_Port_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-json
      description: Output format test
      resource: port
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

// TestNeutron_Port_OutputCSV verifies CSV output has the correct headers.
func TestNeutron_Port_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-csv
      description: Output format test
      resource: port
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

// TestNeutron_Port_DeleteAction verifies delete remediation.
func TestNeutron_Port_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-delete
      description: Delete unused ports
      resource: port
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}
}

// TestNeutron_Port_DryRunSkip verifies dry-run skips remediation.
func TestNeutron_Port_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-dryrun
      description: Dry run delete
      resource: port
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)
	resourceResults.AssertRemediationSkipped(t, "dry-run")
}

// TestNeutron_Port_AllowActionsFiltering verifies the allow-actions filter.
func TestNeutron_Port_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreatePort(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-port-allowlist
      description: Delete not in allowlist
      resource: port
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("port").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected port to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestCleanup_Port cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/neutron/... -run TestCleanup -v
func TestCleanup_Port(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	CleanupOrphans(t, client)
}
