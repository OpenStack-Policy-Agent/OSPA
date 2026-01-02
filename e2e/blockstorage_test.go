//go:build e2e

package e2e

import (
	"testing"
)

// TestBlockStorage_VolumeAudit tests Cinder volume auditing
func TestBlockStorage_VolumeAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-volume-unattached
      description: Test unattached volume check
      service: cinder
      resource: volume
      check:
        status: available
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	volumeResults := results.FilterByService("cinder").FilterByResourceType("volume")
	volumeResults.LogSummary(t)
}

// TestBlockStorage_SnapshotAudit tests Cinder snapshot auditing
func TestBlockStorage_SnapshotAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: test-snapshot-old
      description: Test old snapshot check
      service: cinder
      resource: snapshot
      check:
        age_gt: 7d
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	snapshotResults := results.FilterByService("cinder").FilterByResourceType("snapshot")
	snapshotResults.LogSummary(t)
}

