//go:build e2e

package neutron

// =============================================================================
// SecurityGroupRule E2E TESTS
// =============================================================================
//
// These tests verify OSPA's ability to discover and audit Neutron security group rules.
// Security group rules are audited for dangerous configurations like:
// - SSH (port 22) open to the world (0.0.0.0/0)
// - RDP (port 3389) open to the world
// - ICMP open to the world
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/neutron/... -v -run SecurityGroupRule
//
// =============================================================================

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

// TestNeutron_SecurityGroupRule_SSHOpenToWorld verifies detection of SSH rules open to 0.0.0.0/0.
func TestNeutron_SecurityGroupRule_SSHOpenToWorld(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create a "dangerous" SSH rule - port 22 open to the world
	resourceID, cleanup := CreateSecurityGroupRule(t, client)
	defer cleanup()

	// Policy: find SSH (port 22) ingress rules open to 0.0.0.0/0
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-rule-ssh-open
      description: Find SSH rules open to world
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        ethertype: IPv4
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter to our specific resource
	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected security group rule to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during audit: %d", resourceResults.Errors)
	}

	// The rule should be flagged as a violation (SSH open to world)
	if resourceResults.Violations == 0 {
		t.Error("Expected SSH open to world rule to be flagged as a violation")
	} else {
		t.Log("SSH open to world rule correctly identified as violation - test passed")
	}
}

// TestNeutron_SecurityGroupRule_RDPOpenToWorld verifies detection of RDP rules open to 0.0.0.0/0.
func TestNeutron_SecurityGroupRule_RDPOpenToWorld(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create a "dangerous" RDP rule - port 3389 open to the world
	ruleOpts := rules.CreateOpts{
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   3389,
		PortRangeMax:   3389,
		RemoteIPPrefix: "0.0.0.0/0",
		Description:    "OSPA e2e test rule - RDP from anywhere - safe to delete",
	}

	resourceID, _, cleanup := CreateSecurityGroupRuleWithOptions(t, client, ruleOpts)
	defer cleanup()

	// Policy: find RDP (port 3389) ingress rules open to 0.0.0.0/0
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-rule-rdp-open
      description: Find RDP rules open to world
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        ethertype: IPv4
        protocol: tcp
        port: 3389
        remote_ip_prefix: 0.0.0.0/0
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected security group rule to be scanned but it wasn't discovered")
	}

	// The rule should be flagged as a violation (RDP open to world)
	if resourceResults.Violations == 0 {
		t.Error("Expected RDP open to world rule to be flagged as a violation")
	} else {
		t.Log("RDP open to world rule correctly identified as violation - test passed")
	}
}

// TestNeutron_SecurityGroupRule_SafeRule verifies that safe rules are not flagged.
func TestNeutron_SecurityGroupRule_SafeRule(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create a "safe" rule - SSH from a specific IP range
	ruleOpts := rules.CreateOpts{
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   22,
		PortRangeMax:   22,
		RemoteIPPrefix: "10.0.0.0/8", // Private network, not open to world
		Description:    "OSPA e2e test rule - SSH from private network - safe to delete",
	}

	resourceID, _, cleanup := CreateSecurityGroupRuleWithOptions(t, client, ruleOpts)
	defer cleanup()

	// Policy: find SSH (port 22) ingress rules open to 0.0.0.0/0
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-rule-ssh-open
      description: Find SSH rules open to world
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        ethertype: IPv4
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected security group rule to be scanned but it wasn't discovered")
	}

	// The rule should NOT be flagged (remote_ip_prefix doesn't match 0.0.0.0/0)
	if resourceResults.Violations > 0 {
		t.Error("Expected safe rule (private network) to NOT be flagged as a violation")
	} else {
		t.Log("Safe rule correctly identified as compliant - test passed")
	}
}

// TestNeutron_SecurityGroupRule_Discovery verifies basic discovery works.
func TestNeutron_SecurityGroupRule_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	// Create multiple rules to test batch discovery
	ruleID1, cleanup1 := CreateSecurityGroupRule(t, client)
	defer cleanup1()

	ruleID2, cleanup2 := CreateSecurityGroupRule(t, client)
	defer cleanup2()

	// Simple policy to discover rules
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-discovery
      description: Discover security group rules
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Check both rules were discovered
	results1 := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(ruleID1)

	results2 := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(ruleID2)

	if results1.Scanned == 0 {
		t.Error("First security group rule was not discovered")
	}
	if results2.Scanned == 0 {
		t.Error("Second security group rule was not discovered")
	}

	allRuleResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule")
	t.Logf("Discovery test passed: found %d total security group rules", allRuleResults.Scanned)
}

// TestNeutron_SecurityGroupRule_DeleteAction verifies the delete remediation action works.
func TestNeutron_SecurityGroupRule_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true // Enable remediation
	client := engine.GetNetworkClient(t)

	// Create a "dangerous" SSH rule
	resourceID, cleanup := CreateSecurityGroupRule(t, client)
	defer cleanup() // Cleanup will handle any remaining resources

	// Policy: delete SSH rules open to world
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-rule-delete
      description: Delete SSH rules open to world
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        ethertype: IPv4
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected security group rule to be scanned but it wasn't discovered")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete action: %d", resourceResults.Errors)
	}

	t.Log("Security group rule delete action test completed")
}

// TestNeutron_SecurityGroupRule_DryRunSkip verifies that dry-run mode skips remediation.
func TestNeutron_SecurityGroupRule_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroupRule(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sgr-dryrun
      description: Dry run delete on SSH open to world
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected security group rule to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "dry-run")
}

// TestNeutron_SecurityGroupRule_AllowActionsFiltering verifies the allow-actions filter.
func TestNeutron_SecurityGroupRule_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroupRule(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sgr-allowlist
      description: Delete not in allowlist
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: delete`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected security group rule to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}

// TestNeutron_SecurityGroupRule_Classification verifies classification on SG rule violations.
func TestNeutron_SecurityGroupRule_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	resourceID, cleanup := CreateSecurityGroupRule(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sgr-classification
      description: SSH open to world with classification
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: log
      severity: high
      category: security
      guide_ref: CIS-4.1`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("neutron").
		FilterByResourceType("security_group_rule").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Fatal("Expected security group rule to be scanned")
	}

	resourceResults.AssertClassification(t, "high", "security", "CIS-4.1")
}

// TestNeutron_SecurityGroupRule_OutputJSON verifies JSON output contains required fields.
func TestNeutron_SecurityGroupRule_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateSecurityGroupRule(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sgr-json
      description: Output format test
      resource: security_group_rule
      check:
        direction: ingress
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

// TestNeutron_SecurityGroupRule_OutputCSV verifies CSV output has the correct headers.
func TestNeutron_SecurityGroupRule_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	_, cleanup := CreateSecurityGroupRule(t, client)
	defer cleanup()

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sgr-csv
      description: Output format test
      resource: security_group_rule
      check:
        direction: ingress
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

// TestCleanup_SecurityGroupRule cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/neutron/... -run TestCleanup -v
func TestCleanup_SecurityGroupRule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.GetNetworkClient(t)

	CleanupOrphans(t, client)
}
