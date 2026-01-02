package policy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/services"

	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation"
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services"
)

func TestValidate_UsesServiceSpecificValidator(t *testing.T) {
	// This test should not rely on any built-in service validator semantics.
	// Instead, we register a test validator that always errors, and verify that
	// policy.Load returns that error for matching service/resource rules.

	serviceName := "testsvc_" + policyTestSafeName(t.Name())
	resourceName := "testresource_" + policyTestSafeName(t.Name())

	services.RegisterResource(serviceName, resourceName)

	wantErr := "validator_was_called"
	policy.RegisterValidator(&testValidator{
		serviceName: serviceName,
		err:         fmt.Errorf("%s", wantErr),
	})

	b := []byte(fmt.Sprintf(`version: v1
defaults:
  workers: 1
policies:
  - %s:
    - name: test-rule
      description: test rule
      service: %s
      resource: %s
      check:
        status: active
      action: log
`, serviceName, serviceName, resourceName))

	dir := t.TempDir()
	p := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(p, b, 0644); err != nil {
		t.Fatalf("write policy: %v", err)
	}

	_, err := policy.Load(p)
	if err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
	if err.Error() != wantErr {
		t.Fatalf("Load() error = %q, want %q", err.Error(), wantErr)
	}
}

type testValidator struct {
	serviceName string
	err         error
}

func (v *testValidator) ServiceName() string { return v.serviceName }

func (v *testValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	return v.err
}

func policyTestSafeName(s string) string {
	// Keep only ASCII letters/numbers for registry keys / YAML identifiers.
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r+('a'-'A'))
		case r >= '0' && r <= '9':
			out = append(out, r)
		default:
			// drop
		}
	}
	if len(out) == 0 {
		return "x"
	}
	return string(out)
}


