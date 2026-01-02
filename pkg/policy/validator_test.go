package policy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"

	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation"
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services"
)

func TestValidate_UsesServiceSpecificValidator(t *testing.T) {
	// Nova keypair validation requires check.unused == true (see ComputeValidator).
	b, err := os.ReadFile(filepath.Join("testdata", "invalid-keypair.yaml"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(p, b, 0644); err != nil {
		t.Fatalf("write policy: %v", err)
	}

	_, err = policy.Load(p)
	if err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unused") {
		t.Fatalf("Load() error = %q, want contains %q", err.Error(), "unused")
	}
}


