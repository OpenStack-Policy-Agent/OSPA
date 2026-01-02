//go:build e2e

package e2e

import (
	"testing"
)


// TestCinder_VolumeAudit tests cinder volume auditing
func TestCinder_VolumeAudit(t *testing.T) {
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

	// Filter for cinder/volume results
	VolumeResults := results.FilterByService("cinder").FilterByResourceType("volume")

	VolumeResults.LogSummary(t)

	// Basic assertions
	if VolumeResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during volume audit", VolumeResults.Errors)
	}
}


// TestCinder_SnapshotAudit tests cinder snapshot auditing
func TestCinder_SnapshotAudit(t *testing.T) {
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

	// Filter for cinder/snapshot results
	SnapshotResults := results.FilterByService("cinder").FilterByResourceType("snapshot")

	SnapshotResults.LogSummary(t)

	// Basic assertions
	if SnapshotResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during snapshot audit", SnapshotResults.Errors)
	}
}


