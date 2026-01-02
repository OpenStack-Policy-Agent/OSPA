package network

import (
	"context"
	"fmt"
	"strings"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
)

// SecurityGroupAuditor audits security groups
type SecurityGroupAuditor struct{}

func (a *SecurityGroupAuditor) ResourceType() string {
	return "security_group"
}

func (a *SecurityGroupAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	sg, ok := resource.(groups.SecGroup)
	if !ok {
		return nil, fmt.Errorf("expected groups.SecGroup, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   sg.ID,
		ResourceName: sg.Name,
		ProjectID:    sg.TenantID,
		Compliant:    true,
		Rule:         rule,
		Status:       "active",
	}

	check := rule.Check

	// Check if unused (no instances attached)
	if check.Unused {
		// Check if security group is attached to any ports
		// For now, we'll check if it's in exempt_names first
		for _, exemptName := range check.ExemptNames {
			if strings.EqualFold(sg.Name, exemptName) {
				return result, nil // Exempt, so compliant
			}
		}

		// TODO: Actually check if security group is attached to any instances
		// For now, we'll assume it's unused if it has no rules (heuristic)
		// In a real implementation, we'd query Nova for instances using this SG
		if len(sg.Rules) == 0 {
			result.Compliant = false
			result.Observation = fmt.Sprintf("Security group %s appears to be unused (no rules)", sg.Name)
			return result, nil
		}

		// For now, we'll mark as compliant if it has rules
		// A proper implementation would check Nova for instances using this SG
	}

	return result, nil
}

func (a *SecurityGroupAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	if rule.Action != "delete" {
		return nil
	}

	sg, ok := resource.(groups.SecGroup)
	if !ok {
		return fmt.Errorf("expected groups.SecGroup, got %T", resource)
	}

	serviceClient, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	return groups.Delete(serviceClient, sg.ID).ExtractErr()
}

