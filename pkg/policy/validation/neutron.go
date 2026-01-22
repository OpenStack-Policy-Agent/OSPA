package validation

import (
	"fmt"
	"strings"

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

	case "security_group":
		if err := validateAllowedChecks(check, []string{ "status", "age_gt", "unused", "exempt_names" }); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}


	case "security_group_rule":
		if err := validateAllowedChecks(check, []string{ "status", "age_gt", "unused", "exempt_names" }); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}


	case "floating_ip":
		if err := validateAllowedChecks(check, []string{ "status", "age_gt", "unused", "exempt_names" }); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}


	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for neutron service", ruleName, resourceType)
	}

	return nil
}

func validateAllowedChecks(check *policy.CheckConditions, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = true
	}

	if !hasAnyCheck(check, allowedSet) {
		return fmt.Errorf("check must specify at least one of: %s", strings.Join(allowed, ", "))
	}

	disallowed := findDisallowedChecks(check, allowedSet)
	if len(disallowed) > 0 {
		return fmt.Errorf("check specifies unsupported fields: %s", strings.Join(disallowed, ", "))
	}

	return nil
}

func hasAnyCheck(check *policy.CheckConditions, allowed map[string]bool) bool {
	for name := range allowed {
		if isCheckSet(check, name) {
			return true
		}
	}
	return false
}

func findDisallowedChecks(check *policy.CheckConditions, allowed map[string]bool) []string {
	var disallowed []string
	for _, name := range []string{
		"direction",
		"ethertype",
		"protocol",
		"port",
		"remote_ip_prefix",
		"status",
		"age_gt",
		"unused",
		"exempt_names",
		"exempt_metadata",
		"image_name",
	} {
		if isCheckSet(check, name) && !allowed[name] {
			disallowed = append(disallowed, name)
		}
	}
	return disallowed
}

func isCheckSet(check *policy.CheckConditions, name string) bool {
	switch name {
	case "direction":
		return check.Direction != ""
	case "ethertype":
		return check.Ethertype != ""
	case "protocol":
		return check.Protocol != ""
	case "port":
		return check.Port != 0
	case "remote_ip_prefix":
		return check.RemoteIPPrefix != ""
	case "status":
		return check.Status != ""
	case "age_gt":
		return check.AgeGT != ""
	case "unused":
		return check.Unused
	case "exempt_names":
		return len(check.ExemptNames) > 0
	case "exempt_metadata":
		return check.ExemptMetadata != nil
	case "image_name":
		return len(check.ImageName) > 0
	default:
		return false
	}
}
