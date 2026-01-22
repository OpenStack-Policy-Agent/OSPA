//go:build e2e

package e2e

import (
	"testing"
)


func TestCinder_VolumeAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-check
      description: Test volume check
      service: cinder
      resource: volume
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	VolumeResults := results.FilterByService("cinder").FilterByResourceType("volume")
	VolumeResults.LogSummary(t)

	if VolumeResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during volume audit", VolumeResults.Errors)
	}
}


func TestCinder_SnapshotAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-check
      description: Test snapshot check
      service: cinder
      resource: snapshot
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	SnapshotResults := results.FilterByService("cinder").FilterByResourceType("snapshot")
	SnapshotResults.LogSummary(t)

	if SnapshotResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during snapshot audit", SnapshotResults.Errors)
	}
}


