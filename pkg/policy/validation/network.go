package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// NetworkValidator validates Neutron (network) service policies
type NetworkValidator struct{}

func init() {
	policy.RegisterValidator(&NetworkValidator{})
}

func (v *NetworkValidator) ServiceName() string {
	return "neutron"
}

func (v *NetworkValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
	case "security_group_rule":
		// At least one of direction, protocol, port, or remote_ip_prefix should be set
		if check.Direction == "" && check.Protocol == "" && check.Port == 0 && check.RemoteIPPrefix == "" {
			return fmt.Errorf("rule %q: check must specify at least one of direction, protocol, port, or remote_ip_prefix", ruleName)
		}
		if check.Direction != "" && check.Direction != "ingress" && check.Direction != "egress" {
			return fmt.Errorf("rule %q: check.direction must be 'ingress' or 'egress'", ruleName)
		}
		if check.Ethertype != "" && check.Ethertype != "IPv4" && check.Ethertype != "IPv6" {
			return fmt.Errorf("rule %q: check.ethertype must be 'IPv4' or 'IPv6'", ruleName)
		}

	case "floating_ip":
		// Status check is common
		if check.Status == "" && !check.Unused {
			return fmt.Errorf("rule %q: check must specify status or unused", ruleName)
		}

	case "security_group":
		// Unused check is required
		if !check.Unused {
			return fmt.Errorf("rule %q: check.unused must be true for security_group resource", ruleName)
		}

	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for network service", ruleName, resourceType)
	}

	return nil
}

