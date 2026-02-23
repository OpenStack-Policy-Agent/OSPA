//go:build e2e

package cinder

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestCinder_PolicyLoad verifies that loading a Cinder policy does not panic.
func TestCinder_PolicyLoad(t *testing.T) {
	engine := e2e.NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: smoke-volume
      description: Smoke test for Cinder volume policy
      resource: volume
      check:
        status: available
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	if policy == nil {
		t.Fatal("Expected policy to be loaded, got nil")
	}

	t.Log("Cinder policy load smoke test passed")
}

// TestCinder_AuditSmoke verifies a basic Cinder audit runs without panicking.
func TestCinder_AuditSmoke(t *testing.T) {
	engine := e2e.NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - cinder:
    - name: smoke-volume-audit
      description: Smoke test audit for Cinder volumes
      resource: volume
      check:
        status: available
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	t.Logf("Cinder audit smoke: scanned=%d, violations=%d, errors=%d",
		results.Scanned, results.Violations, results.Errors)
}
