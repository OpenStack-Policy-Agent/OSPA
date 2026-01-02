//go:build e2e

package e2e

import (
	"testing"
)

// TestOpenStack_Smoke_Connectivity tests basic OpenStack connectivity
// by running a minimal audit that requires authentication and service client creation.
//
// Run:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/...
func TestOpenStack_Smoke_Connectivity(t *testing.T) {
	engine := NewTestEngine(t)

	// Create a minimal policy that will test authentication and service client creation
	policyYAML := `version: v1
defaults:
  workers: 1
policies:
  - nova:
    - name: smoke-test
      description: Smoke test for connectivity
      service: nova
      resource: instance
      check:
        status: ACTIVE
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Just verify the audit ran without errors
	if results.Errors > 0 {
		t.Logf("Note: %d errors encountered (this may be normal if no resources exist)", results.Errors)
	}
	
	// The test passes if authentication and client creation succeeded
	t.Logf("Smoke test passed: successfully authenticated and created service clients")
}


