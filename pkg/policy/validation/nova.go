package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// NovaValidator validates Nova service policies
//
// TODO(OSPA): Tighten validation rules for nova over time:
// - Require at least one check condition per rule
// - Validate supported check fields per resource
// - Validate allowed enum values (status/protocol/ethertype/etc.)
type NovaValidator struct{}

func init() {
	policy.RegisterValidator(&NovaValidator{})
}

func (v *NovaValidator) ServiceName() string {
	return "nova"
}

func (v *NovaValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {

	case "instance":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for nova/instance.
		_ = check


	case "keypair":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for nova/keypair.
		_ = check


	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for nova service", ruleName, resourceType)
	}

	return nil
}
