//go:build e2e

package e2e

import (
	"testing"
)


func TestNova_InstanceAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
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

	InstanceResults := results.FilterByService("nova").FilterByResourceType("instance")
	InstanceResults.LogSummary(t)

	if InstanceResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during instance audit", InstanceResults.Errors)
	}
}


func TestNova_KeypairAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
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

	KeypairResults := results.FilterByService("nova").FilterByResourceType("keypair")
	KeypairResults.LogSummary(t)

	if KeypairResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during keypair audit", KeypairResults.Errors)
	}
}


