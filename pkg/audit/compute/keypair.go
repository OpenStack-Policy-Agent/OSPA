package compute

import (
	"context"
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
)

// KeypairAuditor audits Nova keypairs
type KeypairAuditor struct{}

func (a *KeypairAuditor) ResourceType() string {
	return "keypair"
}

func (a *KeypairAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	kp, ok := resource.(keypairs.KeyPair)
	if !ok {
		return nil, fmt.Errorf("expected keypairs.KeyPair, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   kp.Name,
		ResourceName: kp.Name,
		ProjectID:    kp.UserID,
		Compliant:    true,
		Rule:         rule,
		Status:       "active",
	}

	check := rule.Check

	// Check if unused
	if check.Unused {
		// TODO: Actually check if keypair is used by any instances
		// This would require querying all instances and checking their key_name
		// For now, we'll mark as compliant (not unused)
		// In a real implementation, we'd need to cross-reference with instances
		result.Compliant = false
		result.Observation = fmt.Sprintf("Keypair %s may be unused (unused check requires instance cross-reference)", kp.Name)
	}

	return result, nil
}

func (a *KeypairAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	// Keypair violations are typically logged, not fixed
	// If delete action is specified, we would delete the keypair here
	if rule.Action == "delete" {
		kp, ok := resource.(keypairs.KeyPair)
		if !ok {
			return fmt.Errorf("expected keypairs.KeyPair, got %T", resource)
		}

		serviceClient, ok := client.(*gophercloud.ServiceClient)
		if !ok {
			return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
		}

		return keypairs.Delete(serviceClient, kp.Name, nil).ExtractErr()
	}
	return nil
}

