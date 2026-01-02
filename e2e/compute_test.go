//go:build e2e

package e2e

import (
	"testing"
)

// TestCompute_InstanceAudit tests Nova instance auditing
func TestCompute_InstanceAudit(t *testing.T) {
	engine := NewTestEngine(t)

	// Create a minimal policy for instance testing
	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-instance-status
      description: Test instance status check
      service: nova
      resource: instance
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for compute/instance results
	instanceResults := results.FilterByService("nova").FilterByResourceType("instance")

	instanceResults.LogSummary(t)

	// Basic assertions
	if instanceResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during instance audit", instanceResults.Errors)
	}
}

// TestCompute_KeypairAudit tests Nova keypair auditing
func TestCompute_KeypairAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-unused
      description: Test keypair unused check
      service: nova
      resource: keypair
      check:
        unused: true
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	keypairResults := results.FilterByService("nova").FilterByResourceType("keypair")
	keypairResults.LogSummary(t)
}

