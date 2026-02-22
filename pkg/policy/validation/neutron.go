package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// NeutronValidator validates Neutron service policies.
type NeutronValidator struct{}

func init() {
	policy.RegisterValidator(&NeutronValidator{})
}

func (v *NeutronValidator) ServiceName() string {
	return "neutron"
}

func (v *NeutronValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {

	case "network":
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names", "shared_network"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	case "security_group":
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	case "security_group_rule":
		if err := validateAllowedChecks(check, []string{"direction", "ethertype", "protocol", "port", "remote_ip_prefix", "port_range_wide", "exempt_names"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	case "floating_ip":
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "unassociated", "exempt_names"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	case "subnet":
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	case "port":
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names", "no_security_group"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	case "router":
		if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names"}); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for neutron service", ruleName, resourceType)
	}

	return nil
}
