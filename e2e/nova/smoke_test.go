//go:build e2e

package nova

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// TestNova_PolicyLoad verifies that loading a Nova policy does not panic.
func TestNova_PolicyLoad(t *testing.T) {
	engine := e2e.NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: smoke-instance
      description: Smoke test for Nova instance policy
      resource: instance
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	if policy == nil {
		t.Fatal("Expected policy to be loaded, got nil")
	}

	t.Log("Nova policy load smoke test passed")
}

// TestNova_AuditSmoke verifies a basic Nova audit runs without panicking.
func TestNova_AuditSmoke(t *testing.T) {
	engine := e2e.NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - nova:
    - name: smoke-instance-audit
      description: Smoke test audit for Nova instances
      resource: instance
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	t.Logf("Nova audit smoke: scanned=%d, violations=%d, errors=%d",
		results.Scanned, results.Violations, results.Errors)
}
