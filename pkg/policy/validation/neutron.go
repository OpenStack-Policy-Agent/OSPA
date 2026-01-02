package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// NeutronValidator validates Neutron service policies
//
// TODO(OSPA): Tighten validation rules for neutron over time:
// - Require at least one check condition per rule
// - Validate supported check fields per resource
// - Validate allowed enum values (status/protocol/ethertype/etc.)
type NeutronValidator struct{}

func init() {
	policy.RegisterValidator(&NeutronValidator{})
}

func (v *NeutronValidator) ServiceName() string {
	return "neutron"
}

func (v *NeutronValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {

	case "security_group_rule":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for neutron/security_group_rule.
		_ = check


	case "floating_ip":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for neutron/floating_ip.
		_ = check


	case "security_group":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for neutron/security_group.
		_ = check


	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for neutron service", ruleName, resourceType)
	}

	return nil
}
