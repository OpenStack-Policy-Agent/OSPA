package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// BlockStorageValidator validates Cinder (block storage) service policies
type BlockStorageValidator struct{}

func init() {
	Register(&BlockStorageValidator{})
}

func (v *BlockStorageValidator) ServiceName() string {
	return "cinder"
}

func (v *BlockStorageValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
	case "volume":
		// Status or age_gt should be set
		if check.Status == "" && check.AgeGT == "" {
			return fmt.Errorf("rule %q: check must specify at least one of status or age_gt", ruleName)
		}

	case "snapshot":
		// Age_gt is typically required
		if check.AgeGT == "" {
			return fmt.Errorf("rule %q: check.age_gt is required for snapshot resource", ruleName)
		}

	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for block storage service", ruleName, resourceType)
	}

	return nil
}

