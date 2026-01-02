package network

import (
	"context"
	"fmt"
	"strings"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

// SecurityGroupRuleAuditor audits security group rules
type SecurityGroupRuleAuditor struct{}

func (a *SecurityGroupRuleAuditor) ResourceType() string {
	return "security_group_rule"
}

func (a *SecurityGroupRuleAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	sgRule, ok := resource.(rules.SecGroupRule)
	if !ok {
		return nil, fmt.Errorf("expected rules.SecGroupRule, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   sgRule.ID,
		ResourceName: fmt.Sprintf("sg-rule-%s", sgRule.ID),
		ProjectID:    sgRule.TenantID,
		Compliant:    true,
		Rule:         rule,
		Status:       "active",
	}

	check := rule.Check

	// Check direction
	if check.Direction != "" {
		if strings.ToLower(sgRule.Direction) != strings.ToLower(check.Direction) {
			return result, nil // Not a match, but compliant
		}
	}

	// Check ethertype
	if check.Ethertype != "" {
		if sgRule.EtherType != check.Ethertype {
			return result, nil // Not a match, but compliant
		}
	}

	// Check protocol
	if check.Protocol != "" {
		if strings.ToLower(sgRule.Protocol) != strings.ToLower(check.Protocol) {
			return result, nil // Not a match, but compliant
		}
	}

	// Check port
	if check.Port != 0 {
		if sgRule.PortRangeMin == nil || *sgRule.PortRangeMin != check.Port {
			return result, nil // Not a match, but compliant
		}
	}

	// Check remote IP prefix
	if check.RemoteIPPrefix != "" {
		if sgRule.RemoteIPPrefix == nil || *sgRule.RemoteIPPrefix != check.RemoteIPPrefix {
			return result, nil // Not a match, but compliant
		}
	}

	// If we get here, all conditions match - this is a violation
	result.Compliant = false
	result.Observation = fmt.Sprintf("Security group rule matches policy violation criteria: direction=%s, protocol=%s, port=%d, remote_ip_prefix=%s",
		sgRule.Direction, sgRule.Protocol, check.Port, check.RemoteIPPrefix)

	return result, nil
}

func (a *SecurityGroupRuleAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	// Security group rule violations are typically logged, not fixed
	// If delete action is specified, we would delete the rule here
	if rule.Action == "delete" {
		// Implementation would delete the security group rule
		// For now, return an error indicating it's not implemented
		return fmt.Errorf("delete action for security_group_rule is not yet implemented")
	}
	return nil
}

