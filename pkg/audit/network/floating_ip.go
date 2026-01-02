package network

import (
	"context"
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
)

// FloatingIPAuditor audits floating IPs
type FloatingIPAuditor struct{}

func (a *FloatingIPAuditor) ResourceType() string {
	return "floating_ip"
}

func (a *FloatingIPAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	fip, ok := resource.(floatingips.FloatingIP)
	if !ok {
		return nil, fmt.Errorf("expected floatingips.FloatingIP, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   fip.ID,
		ResourceName: fip.FloatingIP,
		ProjectID:    fip.TenantID,
		Compliant:    true,
		Rule:         rule,
		Status:       fip.Status,
	}

	check := rule.Check

	// Check status
	if check.Status != "" {
		if fip.Status != check.Status {
			return result, nil // Not a match, but compliant
		}
	}

	// Check if unassociated (DOWN status typically means unassociated)
	if check.Status == "DOWN" || (check.Status == "" && fip.Status == "DOWN") {
		if fip.PortID == "" {
			result.Compliant = false
			result.Observation = fmt.Sprintf("Floating IP %s is not associated with any instance (status: %s)", fip.FloatingIP, fip.Status)
			return result, nil
		}
	}

	// If status matches and it's DOWN/unassociated, it's a violation
	if check.Status == "DOWN" && fip.Status == "DOWN" && fip.PortID == "" {
		result.Compliant = false
		result.Observation = fmt.Sprintf("Floating IP %s is not associated with any instance", fip.FloatingIP)
	}

	return result, nil
}

func (a *FloatingIPAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	if rule.Action != "delete" {
		return nil
	}

	fip, ok := resource.(floatingips.FloatingIP)
	if !ok {
		return fmt.Errorf("expected floatingips.FloatingIP, got %T", resource)
	}

	serviceClient, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	return floatingips.Delete(serviceClient, fip.ID).ExtractErr()
}

