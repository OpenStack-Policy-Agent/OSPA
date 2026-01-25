package policy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"

	// Ensure services/resources are registered for validation.
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services"
	// Ensure service-specific check validation is registered.
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation"
)

func writeTempPolicy(t *testing.T, yaml string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(p, []byte(yaml), 0644); err != nil {
		t.Fatalf("write policy file: %v", err)
	}
	return p
}

func TestLoad_ServiceKeyedStructure_SetsRuleServiceFromParent(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "valid-policy.yaml"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	path := writeTempPolicy(t, string(b))

	p, err := policy.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(p.Policies) != 1 {
		t.Fatalf("Policies len = %d, want 1", len(p.Policies))
	}
	if got := p.Policies[0].Service; got != "nova" {
		t.Fatalf("Policies[0].Service = %q, want %q", got, "nova")
	}
	if got := p.Policies[0].Rules[0].Service; got != "nova" {
		t.Fatalf("Policies[0].Rules[0].Service = %q, want %q", got, "nova")
	}
}

func TestLoad_RejectsUnsupportedResource(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "invalid-resource.yaml"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	path := writeTempPolicy(t, string(b))

	_, err = policy.Load(path)
	if err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unsupported resource") {
		t.Fatalf("Load() error = %q, want contains %q", err.Error(), "unsupported resource")
	}
}

func TestLoad_RejectsDuplicateRuleNames(t *testing.T) {
	path := writeTempPolicy(t, `
version: v1
defaults:
  workers: 1
  output: out.json
policies:
  - nova:
    - name: dup
      description: one
      resource: instance
      check:
        status: ACTIVE
      action: log
  - neutron:
    - name: dup
      description: two
      resource: security_group
      check:
        unused: true
      action: log
`)

	_, err := policy.Load(path)
	if err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "duplicate rule name") {
		t.Fatalf("Load() error = %q, want contains %q", err.Error(), "duplicate rule name")
	}
}
