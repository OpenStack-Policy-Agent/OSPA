package neutron

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/common"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

type subnetAdapter struct{ s subnets.Subnet }

func (a subnetAdapter) GetID() string           { return a.s.ID }
func (a subnetAdapter) GetName() string         { return a.s.Name }
func (a subnetAdapter) GetProjectID() string    { return a.s.TenantID }
func (a subnetAdapter) GetStatus() string       { return "" }
func (a subnetAdapter) GetCreatedAt() time.Time { return time.Time{} }
func (a subnetAdapter) GetUpdatedAt() time.Time { return time.Time{} }

// SubnetAuditor audits neutron/subnet resources.
//
// Allowed checks: status, age_gt, unused, exempt_names
// Allowed actions: log, delete, tag
//
// Note: subnets in Neutron have no Status or timestamp fields. The status
// and age_gt checks are accepted for policy consistency but are no-ops.
// The unused check flags subnets with empty allocation pools (no IP ranges
// available for port allocation).
type SubnetAuditor struct{}

func (a *SubnetAuditor) ResourceType() string {
	return "subnet"
}

func (a *SubnetAuditor) ImplementedChecks() []string {
	return []string{"status", "age_gt", "unused", "exempt_names"}
}

func (a *SubnetAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	subnet, ok := resource.(subnets.Subnet)
	if !ok {
		return nil, fmt.Errorf("expected subnets.Subnet, got %T", resource)
	}

	adapter := subnetAdapter{s: subnet}
	result := common.BuildBaseResult(adapter, rule)

	exempt, err := common.RunCommonChecks(adapter, rule, result)
	if exempt || err != nil {
		return result, err
	}

	if rule.Check.Unused {
		if len(subnet.AllocationPools) == 0 {
			result.Compliant = false
			result.Observation = "subnet has no allocation pools"
		}
	}

	return result, nil
}

func (a *SubnetAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx

	if rule.Action == "log" {
		return nil
	}

	c, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	subnet, ok := resource.(subnets.Subnet)
	if !ok {
		return fmt.Errorf("expected subnets.Subnet, got %T", resource)
	}

	switch rule.Action {

	case "delete":
		portPages, err := ports.List(c, ports.ListOpts{NetworkID: subnet.NetworkID}).AllPages()
		if err != nil {
			return fmt.Errorf("listing ports for subnet %s: %w", subnet.ID, err)
		}
		portList, err := ports.ExtractPorts(portPages)
		if err != nil {
			return fmt.Errorf("extracting ports: %w", err)
		}
		for _, p := range portList {
			for _, ip := range p.FixedIPs {
				if ip.SubnetID == subnet.ID {
					return fmt.Errorf("cannot delete subnet %s: port %s has a fixed IP on it", subnet.ID, p.ID)
				}
			}
		}

		if err := subnets.Delete(c, subnet.ID).ExtractErr(); err != nil {
			return fmt.Errorf("deleting subnet %s: %w", subnet.ID, err)
		}
		return nil

	case "tag":
		return fmt.Errorf("neutron/subnet: tag action not yet implemented")

	default:
		return fmt.Errorf("neutron/subnet: action %q not implemented", rule.Action)
	}
}
