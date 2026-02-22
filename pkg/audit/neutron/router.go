package neutron

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/common"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

type routerAdapter struct{ r routers.Router }

func (a routerAdapter) GetID() string           { return a.r.ID }
func (a routerAdapter) GetName() string         { return a.r.Name }
func (a routerAdapter) GetProjectID() string    { return a.r.TenantID }
func (a routerAdapter) GetStatus() string       { return a.r.Status }
func (a routerAdapter) GetCreatedAt() time.Time { return time.Time{} }
func (a routerAdapter) GetUpdatedAt() time.Time { return time.Time{} }

// RouterAuditor audits neutron/router resources.
//
// Allowed checks: status, age_gt, unused, exempt_names
// Allowed actions: log, delete, tag
//
// Note: routers.Router in gophercloud v1.14.1 has no timestamp fields.
// The age_gt check is accepted for policy consistency but is a no-op.
// The unused check flags routers with no external gateway configured
// (GatewayInfo.NetworkID is empty).
type RouterAuditor struct{}

func (a *RouterAuditor) ResourceType() string {
	return "router"
}

func (a *RouterAuditor) ImplementedChecks() []string {
	return []string{"status", "age_gt", "unused", "exempt_names"}
}

func (a *RouterAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	router, ok := resource.(routers.Router)
	if !ok {
		return nil, fmt.Errorf("expected routers.Router, got %T", resource)
	}

	adapter := routerAdapter{r: router}
	result := common.BuildBaseResult(adapter, rule)

	exempt, err := common.RunCommonChecks(adapter, rule, result)
	if exempt || err != nil {
		return result, err
	}

	if rule.Check.Unused {
		if router.GatewayInfo.NetworkID == "" {
			result.Compliant = false
			result.Observation = "router has no external gateway"
		}
	}

	return result, nil
}

func (a *RouterAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx

	if rule.Action == "log" {
		return nil
	}

	c, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	router, ok := resource.(routers.Router)
	if !ok {
		return fmt.Errorf("expected routers.Router, got %T", resource)
	}

	switch rule.Action {
	case "delete":
		// Safety: refuse to delete a router that still has ports attached
		portPages, err := ports.List(c, ports.ListOpts{DeviceID: router.ID}).AllPages()
		if err != nil {
			return fmt.Errorf("listing ports for router %s: %w", router.ID, err)
		}
		portList, err := ports.ExtractPorts(portPages)
		if err != nil {
			return fmt.Errorf("extracting ports: %w", err)
		}
		if len(portList) > 0 {
			return fmt.Errorf("cannot delete router %s: has %d attached ports", router.ID, len(portList))
		}

		if err := routers.Delete(c, router.ID).ExtractErr(); err != nil {
			return fmt.Errorf("deleting router %s: %w", router.ID, err)
		}
		return nil

	case "tag":
		return fmt.Errorf("neutron/router: tag action not yet implemented")

	default:
		return fmt.Errorf("neutron/router: action %q not implemented", rule.Action)
	}
}
