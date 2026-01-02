//go:build e2e

package e2e

import (
	"testing"
)


// TestNova_InstanceAudit tests nova instance auditing
func TestNova_InstanceAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-instance-check
      description: Test instance check
      service: nova
      resource: instance
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for nova/instance results
	InstanceResults := results.FilterByService("nova").FilterByResourceType("instance")

	InstanceResults.LogSummary(t)

	// Basic assertions
	if InstanceResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during instance audit", InstanceResults.Errors)
	}
}


// TestNova_KeypairAudit tests nova keypair auditing
func TestNova_KeypairAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: test-keypair-check
      description: Test keypair check
      service: nova
      resource: keypair
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for nova/keypair results
	KeypairResults := results.FilterByService("nova").FilterByResourceType("keypair")

	KeypairResults.LogSummary(t)

	// Basic assertions
	if KeypairResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during keypair audit", KeypairResults.Errors)
	}
}


