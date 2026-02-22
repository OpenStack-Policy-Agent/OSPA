package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// CinderValidator validates Cinder service policies.
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
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names", "encrypted", "attached", "has_backup"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	case "snapshot":
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names", "encrypted"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for cinder service", ruleName, resourceType)
	}

	return nil
}
