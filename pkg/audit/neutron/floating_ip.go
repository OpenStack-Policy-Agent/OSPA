package neutron

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/common"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
)

type floatingIpAdapter struct{ f floatingips.FloatingIP }

func (a floatingIpAdapter) GetID() string           { return a.f.ID }
func (a floatingIpAdapter) GetName() string         { return a.f.Description }
func (a floatingIpAdapter) GetProjectID() string    { return a.f.TenantID }
func (a floatingIpAdapter) GetStatus() string       { return a.f.Status }
func (a floatingIpAdapter) GetCreatedAt() time.Time { return a.f.CreatedAt }
func (a floatingIpAdapter) GetUpdatedAt() time.Time { return a.f.UpdatedAt }

// FloatingIpAuditor audits neutron/floating_ip resources.
//
// Allowed checks: status, age_gt, unused, unassociated, exempt_names
// Allowed actions: log, delete, tag
//
// FloatingIP has no Name field; exempt_names matches against Description.
// Both unused and unassociated flag floating IPs with PortID == ""
// (not attached to any port).
type FloatingIpAuditor struct{}

func (a *FloatingIpAuditor) ResourceType() string {
	return "floating_ip"
}

func (a *FloatingIpAuditor) ImplementedChecks() []string {
	return []string{"status", "age_gt", "unused", "unassociated", "exempt_names"}
}

func (a *FloatingIpAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	fip, ok := resource.(floatingips.FloatingIP)
	if !ok {
		return nil, fmt.Errorf("expected floatingips.FloatingIP, got %T", resource)
	}

	adapter := floatingIpAdapter{f: fip}
	result := common.BuildBaseResult(adapter, rule)

	exempt, err := common.RunCommonChecks(adapter, rule, result)
	if exempt || err != nil {
		return result, err
	}

	if rule.Check.Unused {
		if fip.PortID == "" {
			result.Compliant = false
			result.Observation = "floating IP is not associated with any port"
		}
	}

	if rule.Check.Unassociated {
		if fip.PortID == "" {
			result.Compliant = false
			result.Observation = "floating IP is not associated with any port"
		}
	}

	return result, nil
}

func (a *FloatingIpAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx

	if rule.Action == "log" {
		return nil
	}

	c, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	fip, ok := resource.(floatingips.FloatingIP)
	if !ok {
		return fmt.Errorf("expected floatingips.FloatingIP, got %T", resource)
	}

	switch rule.Action {
	case "delete":
		if err := floatingips.Delete(c, fip.ID).ExtractErr(); err != nil {
			return fmt.Errorf("deleting floating IP %s: %w", fip.ID, err)
		}
		return nil

	case "tag":
		return fmt.Errorf("neutron/floating_ip: tag action not yet implemented")

	default:
		return fmt.Errorf("neutron/floating_ip: action %q not implemented", rule.Action)
	}
}
