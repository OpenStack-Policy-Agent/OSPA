package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// CinderValidator validates Cinder service policies
//
// TODO(OSPA): Tighten validation rules for cinder over time:
// - Require at least one check condition per rule
// - Validate supported check fields per resource
// - Validate allowed enum values (status/protocol/ethertype/etc.)
type CinderValidator struct{}

func init() {
	policy.RegisterValidator(&CinderValidator{})
}

func (v *CinderValidator) ServiceName() string {
	return "cinder"
}

func (v *CinderValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {

	case "volume":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for cinder/volume.
		_ = check

	case "snapshot":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for cinder/snapshot.
		_ = check

	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for cinder service", ruleName, resourceType)
	}

	return nil
}
