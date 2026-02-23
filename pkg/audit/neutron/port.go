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
)

type portAdapter struct{ p ports.Port }

func (a portAdapter) GetID() string           { return a.p.ID }
func (a portAdapter) GetName() string         { return a.p.Name }
func (a portAdapter) GetProjectID() string    { return a.p.TenantID }
func (a portAdapter) GetStatus() string       { return a.p.Status }
func (a portAdapter) GetCreatedAt() time.Time { return a.p.CreatedAt }
func (a portAdapter) GetUpdatedAt() time.Time { return a.p.UpdatedAt }

// PortAuditor audits neutron/port resources.
//
// Allowed checks: status, age_gt, unused, exempt_names, no_security_group
// Allowed actions: log, delete, tag
//
// The unused check flags ports not attached to any device (DeviceID is empty).
// The no_security_group check flags ports with no security groups attached.
type PortAuditor struct{}

func (a *PortAuditor) ResourceType() string {
	return "port"
}

func (a *PortAuditor) ImplementedChecks() []string {
	return []string{"status", "age_gt", "unused", "exempt_names", "no_security_group"}
}

func (a *PortAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	port, ok := resource.(ports.Port)
	if !ok {
		return nil, fmt.Errorf("expected ports.Port, got %T", resource)
	}

	adapter := portAdapter{p: port}
	result := common.BuildBaseResult(adapter, rule)

	exempt, err := common.RunCommonChecks(adapter, rule, result)
	if exempt || err != nil {
		return result, err
	}

	if rule.Check.Unused {
		if port.DeviceID == "" {
			result.Compliant = false
			result.Observation = "port is not attached to any device"
		}
	}

	if rule.Check.NoSecurityGroup {
		if len(port.SecurityGroups) == 0 {
			result.Compliant = false
			result.Observation = "port has no security groups attached"
		}
	}

	return result, nil
}

func (a *PortAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx

	if rule.Action == "log" {
		return nil
	}

	c, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	port, ok := resource.(ports.Port)
	if !ok {
		return fmt.Errorf("expected ports.Port, got %T", resource)
	}

	switch rule.Action {
	case "delete":
		if err := ports.Delete(c, port.ID).ExtractErr(); err != nil {
			return fmt.Errorf("deleting port %s: %w", port.ID, err)
		}
		return nil

	case "tag":
		return fmt.Errorf("neutron/port: tag action not yet implemented")

	default:
		return fmt.Errorf("neutron/port: action %q not implemented", rule.Action)
	}
}
