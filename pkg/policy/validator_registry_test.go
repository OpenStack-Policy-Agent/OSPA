package policy_test

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"

	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation"
)

func TestValidatorRegistry_HasBuiltins(t *testing.T) {
	for _, svc := range []string{"nova", "neutron", "cinder"} {
		if _, err := policy.MustGetValidator(svc); err != nil {
			t.Fatalf("MustGetValidator(%q) error = %v", svc, err)
		}
	}
}
