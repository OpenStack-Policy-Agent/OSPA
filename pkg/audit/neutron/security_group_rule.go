package neutron

import (
	"context"
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

// SecurityGroupRuleAuditor audits neutron/security_group_rule resources.
//
// Allowed checks: direction, ethertype, protocol, port, remote_ip_prefix, exempt_names
// Allowed actions: log, delete
type SecurityGroupRuleAuditor struct{}

func (a *SecurityGroupRuleAuditor) ResourceType() string {
	return "security_group_rule"
}

func (a *SecurityGroupRuleAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	sgRule, ok := resource.(rules.SecGroupRule)
	if !ok {
		return nil, fmt.Errorf("expected rules.SecGroupRule, got %T", resource)
	}

	// Build a descriptive name for the rule (rules don't have names)
	ruleName := buildRuleName(sgRule)

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   sgRule.ID,
		ResourceName: ruleName,
		ProjectID:    sgRule.TenantID,
		Status:       "ACTIVE", // Rules don't have status; they're always active
		Compliant:    true,
		Rule:         rule,
	}

	// Check exemptions first (by security group name pattern - not directly applicable to rules)
	// Rules don't have names, but we can exempt based on parent SG ID pattern
	if isExemptByName(sgRule.SecGroupID, rule.Check.ExemptNames) {
		result.Compliant = true
		result.Observation = "exempt by security group ID pattern"
		return result, nil
	}

	// Security group rule specific checks - all must match for non-compliance
	// This is used to find "dangerous" rules like SSH open to world
	allChecksMatch := true
	var observations []string

	// Direction check (ingress/egress)
	if rule.Check.Direction != "" {
		if sgRule.Direction != rule.Check.Direction {
			allChecksMatch = false
		} else {
			observations = append(observations, fmt.Sprintf("direction=%s", sgRule.Direction))
		}
	}

	// Ethertype check (IPv4/IPv6)
	if rule.Check.Ethertype != "" {
		if sgRule.EtherType != rule.Check.Ethertype {
			allChecksMatch = false
		} else {
			observations = append(observations, fmt.Sprintf("ethertype=%s", sgRule.EtherType))
		}
	}

	// Protocol check (tcp/udp/icmp/etc)
	if rule.Check.Protocol != "" {
		if sgRule.Protocol != rule.Check.Protocol {
			allChecksMatch = false
		} else {
			observations = append(observations, fmt.Sprintf("protocol=%s", sgRule.Protocol))
		}
	}

	// Port check - matches if the rule's port range includes the specified port
	if rule.Check.Port != 0 {
		if !portMatches(sgRule.PortRangeMin, sgRule.PortRangeMax, rule.Check.Port) {
			allChecksMatch = false
		} else {
			observations = append(observations, fmt.Sprintf("port=%d (range %d-%d)", rule.Check.Port, sgRule.PortRangeMin, sgRule.PortRangeMax))
		}
	}

	// Remote IP prefix check (e.g., 0.0.0.0/0 for "open to world")
	if rule.Check.RemoteIPPrefix != "" {
		if sgRule.RemoteIPPrefix != rule.Check.RemoteIPPrefix {
			allChecksMatch = false
		} else {
			observations = append(observations, fmt.Sprintf("remote_ip_prefix=%s", sgRule.RemoteIPPrefix))
		}
	}

	// If all specified checks match, the rule is non-compliant (it's a "dangerous" rule)
	if allChecksMatch && len(observations) > 0 {
		result.Compliant = false
		result.Observation = fmt.Sprintf("rule matches policy criteria: %v", observations)
	}

	return result, nil
}

func (a *SecurityGroupRuleAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx

	// Log action doesn't require client or resource validation
	if rule.Action == "log" {
		return nil
	}

	c, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	sgRule, ok := resource.(rules.SecGroupRule)
	if !ok {
		return fmt.Errorf("expected rules.SecGroupRule, got %T", resource)
	}

	switch rule.Action {
	case "delete":
		// Delete the security group rule
		if err := rules.Delete(c, sgRule.ID).ExtractErr(); err != nil {
			return fmt.Errorf("deleting security group rule %s: %w", sgRule.ID, err)
		}
		return nil

	default:
		return fmt.Errorf("neutron/security_group_rule: action %q not implemented", rule.Action)
	}
}

// buildRuleName creates a descriptive name for a security group rule
func buildRuleName(r rules.SecGroupRule) string {
	proto := r.Protocol
	if proto == "" {
		proto = "any"
	}

	portRange := ""
	if r.PortRangeMin > 0 || r.PortRangeMax > 0 {
		if r.PortRangeMin == r.PortRangeMax {
			portRange = fmt.Sprintf(":%d", r.PortRangeMin)
		} else {
			portRange = fmt.Sprintf(":%d-%d", r.PortRangeMin, r.PortRangeMax)
		}
	}

	remote := r.RemoteIPPrefix
	if remote == "" && r.RemoteGroupID != "" {
		remote = fmt.Sprintf("sg:%s", r.RemoteGroupID)
	}
	if remote == "" {
		remote = "any"
	}

	return fmt.Sprintf("%s/%s%s from %s", r.Direction, proto, portRange, remote)
}

// portMatches checks if a port falls within the rule's port range
func portMatches(min, max, port int) bool {
	// If both min and max are 0, the rule applies to all ports
	if min == 0 && max == 0 {
		return true
	}
	return port >= min && port <= max
}
