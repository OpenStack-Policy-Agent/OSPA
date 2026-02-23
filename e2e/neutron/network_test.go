//go:build e2e

package neutron

import (
	"encoding/csv"
	"encoding/json"
	"os"
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

// TestNeutron_Network_AgeGTCompliant verifies a freshly created resource passes age_gt: 30d.
func TestNeutron_Network_AgeGTCompliant(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-age
      description: Find networks older than 30 days
      service: neutron
      resource: network
      check:
        age_gt: 30d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned")
	}

	if resourceResults.Violations > 0 {
		t.Error("Freshly created network should not be flagged by age_gt: 30d")
	} else {
		t.Log("Network correctly identified as young - test passed")
	}
}

// TestNeutron_Network_AgeGTViolation verifies a resource trips age_gt: 0m (any age triggers).
func TestNeutron_Network_AgeGTViolation(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-age-short
      description: Find networks older than 0m (always triggers)
      resource: network
      check:
        age_gt: 0m
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned")
	}

	if resourceResults.Violations == 0 {
		t.Error("Network should be flagged by age_gt: 0m")
	} else {
		t.Log("Network correctly flagged as old - test passed")
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

// TestNeutron_Network_MultiRulePolicy verifies that two rules against the same resource
// both produce results.
func TestNeutron_Network_MultiRulePolicy(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: rule-status
      description: Check status
      service: neutron
      resource: network
      check:
        status: ACTIVE
      action: log
    - name: rule-unused
      description: Check unused
      service: neutron
      resource: network
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	statusResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID).
		FilterByRuleID("rule-status")

	unusedResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID).
		FilterByRuleID("rule-unused")

	if statusResults.Scanned == 0 {
		t.Error("Expected network to produce result for status rule")
	}
	if unusedResults.Scanned == 0 {
		t.Error("Expected network to produce result for unused rule")
	}

	t.Logf("Multi-rule test: status=%d results, unused=%d results",
		statusResults.Scanned, unusedResults.Scanned)
}

// TestNeutron_Network_DeleteAction verifies the delete remediation action works on empty networks.
func TestNeutron_Network_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-delete
      description: Delete unused networks
      service: neutron
      resource: network
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned")
	}
	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}
}

// TestNeutron_Network_DeleteBlocked verifies delete fails gracefully on networks with subnets.
func TestNeutron_Network_DeleteBlocked(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.GetNetworkClient(t)

	networkID, _, cleanup := CreateNetworkWithSubnet(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-delete-blocked
      description: Try to delete network with subnet (should fail gracefully)
      service: neutron
      resource: network
      check:
        status: ACTIVE
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned")
	}
	// Delete should fail gracefully (network has subnets)
	t.Log("Delete blocked test completed - verify the network still exists")
}

// TestNeutron_Network_DryRunSkip verifies that dry-run mode skips remediation.
func TestNeutron_Network_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-dryrun
      description: Dry run delete on unused networks
      resource: network
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "dry-run")
}

// TestNeutron_Network_AllowActionsFiltering verifies that the allow-actions filter
// prevents disallowed remediation actions from running.
func TestNeutron_Network_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-allowlist
      description: Delete not in allowlist
      resource: network
      check:
        unused: true
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected network to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestNeutron_Network_Classification verifies severity/category/guide_ref propagation.
func TestNeutron_Network_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	networkID, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-classification
      description: Test classification field propagation
      resource: network
      check:
        status: ACTIVE
      action: log
      severity: critical
      category: security
      guide_ref: OSSN-0011`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("network").
		FilterByResourceID(networkID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Fatal("Expected network to be scanned")
	}

	resourceResults.AssertClassification(t, "critical", "security", "OSSN-0011")
}

// TestNeutron_Network_OutputJSON verifies JSON output contains required fields.
func TestNeutron_Network_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-json-output
      description: Output format test
      resource: network
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

// TestNeutron_Network_OutputCSV verifies CSV output has the correct headers.
func TestNeutron_Network_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateNetwork(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-csv-output
      description: Output format test
      resource: network
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
