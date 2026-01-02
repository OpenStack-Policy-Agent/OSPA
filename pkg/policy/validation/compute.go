package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// ComputeValidator validates Nova (compute) service policies
type ComputeValidator struct{}

func init() {
	policy.RegisterValidator(&ComputeValidator{})
}

func (v *ComputeValidator) ServiceName() string {
	return "nova"
}

func (v *ComputeValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
	case "instance":
		// At least one check condition should be set
		if check.AgeGT == "" && len(check.ImageName) == 0 && check.Status == "" {
			return fmt.Errorf("rule %q: check must specify at least one of age_gt, image_name, or status", ruleName)
		}

	case "keypair":
		// Unused check is required
		if !check.Unused {
			return fmt.Errorf("rule %q: check.unused must be true for keypair resource", ruleName)
		}

	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for compute service", ruleName, resourceType)
	}

	return nil
}

