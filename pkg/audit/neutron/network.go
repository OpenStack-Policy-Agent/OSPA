package neutron

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/common"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

type networkAdapter struct{ n networks.Network }

func (a networkAdapter) GetID() string           { return a.n.ID }
func (a networkAdapter) GetName() string         { return a.n.Name }
func (a networkAdapter) GetProjectID() string    { return a.n.TenantID }
func (a networkAdapter) GetStatus() string       { return a.n.Status }
func (a networkAdapter) GetCreatedAt() time.Time { return a.n.CreatedAt }
func (a networkAdapter) GetUpdatedAt() time.Time { return a.n.UpdatedAt }

// NetworkAuditor audits neutron/network resources.
//
// Allowed checks: status, age_gt, unused, exempt_names, shared_network
// Allowed actions: log, delete, tag
type NetworkAuditor struct{}

func (a *NetworkAuditor) ResourceType() string {
	return "network"
}

func (a *NetworkAuditor) ImplementedChecks() []string {
	return []string{"status", "age_gt", "unused", "exempt_names"}
}

func (a *NetworkAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	network, ok := resource.(networks.Network)
	if !ok {
		return nil, fmt.Errorf("expected networks.Network, got %T", resource)
	}

	adapter := networkAdapter{n: network}
	result := common.BuildBaseResult(adapter, rule)

	exempt, err := common.RunCommonChecks(adapter, rule, result)
	if exempt || err != nil {
		return result, err
	}

	if rule.Check.Unused {
		if len(network.Subnets) == 0 {
			result.Compliant = false
			result.Observation = "network has no subnets"
		}
	}

	return result, nil
}

func (a *NetworkAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx

	// Log action doesn't require client or resource validation
	if rule.Action == "log" {
		return nil
	}

	c, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	network, ok := resource.(networks.Network)
	if !ok {
		return fmt.Errorf("expected networks.Network, got %T", resource)
	}

	switch rule.Action {

	case "delete":
		// First check if there are any ports on this network
		portPages, err := ports.List(c, ports.ListOpts{NetworkID: network.ID}).AllPages()
		if err != nil {
			return fmt.Errorf("listing ports for network %s: %w", network.ID, err)
		}
		portList, err := ports.ExtractPorts(portPages)
		if err != nil {
			return fmt.Errorf("extracting ports: %w", err)
		}
		if len(portList) > 0 {
			return fmt.Errorf("cannot delete network %s: has %d attached ports", network.ID, len(portList))
		}

		// Delete the network
		if err := networks.Delete(c, network.ID).ExtractErr(); err != nil {
			return fmt.Errorf("deleting network %s: %w", network.ID, err)
		}
		return nil

	case "tag":
		// Neutron networks support tags via the tags extension
		// For now, return not implemented - would need tags extension
		return fmt.Errorf("neutron/network: tag action not yet implemented")

	default:
		return fmt.Errorf("neutron/network: action %q not implemented", rule.Action)
	}
}

// isExemptByName checks if the resource name matches any exempt pattern.
// Supports glob patterns like "ospa-e2e-*" and "test-*-network".
func isExemptByName(name string, exemptNames []string) bool {
	for _, pattern := range exemptNames {
		// Try exact match first
		if name == pattern {
			return true
		}
		// Try glob pattern match
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}
