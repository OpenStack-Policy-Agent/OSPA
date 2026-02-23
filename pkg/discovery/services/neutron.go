package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

// NeutronNetworkDiscoverer discovers neutron/network resources.
type NeutronNetworkDiscoverer struct{}

func (d *NeutronNetworkDiscoverer) ResourceType() string {
	return "network"
}

func (d *NeutronNetworkDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		opts := networks.ListOpts{}
		pages, err := networks.List(client, opts).AllPages()
		if err != nil {
			return
		}

		networkList, err := networks.ExtractNetworks(pages)
		if err != nil {
			return
		}

		for _, network := range networkList {
			select {
			case <-ctx.Done():
				return
			case ch <- discovery.Job{
				Service:      "neutron",
				ResourceType: "network",
				ResourceID:   network.ID,
				ProjectID:    network.TenantID,
				Resource:     network,
			}:
			}
		}
	}()

	return ch, nil
}

// NeutronSecurityGroupDiscoverer discovers neutron/security_group resources.
//
// TODO: Implement Discover() using gophercloud to list security_group resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/neutron
//
// Discovery hints from registry:
//
//	pagination: false
//	all_tenants: false
//	regions: false
type NeutronSecurityGroupDiscoverer struct{}

func (d *NeutronSecurityGroupDiscoverer) ResourceType() string {
	return "security_group"
}

func (d *NeutronSecurityGroupDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		opts := groups.ListOpts{}
		pages, err := groups.List(client, opts).AllPages()
		if err != nil {
			return
		}

		sgList, err := groups.ExtractGroups(pages)
		if err != nil {
			return
		}

		for _, sg := range sgList {
			select {
			case <-ctx.Done():
				return
			case ch <- discovery.Job{
				Service:      "neutron",
				ResourceType: "security_group",
				ResourceID:   sg.ID,
				ProjectID:    sg.TenantID,
				Resource:     sg,
			}:
			}
		}
		_ = allTenants
	}()

	return ch, nil
}

// NeutronSecurityGroupRuleDiscoverer discovers neutron/security_group_rule resources.
//
// TODO: Implement Discover() using gophercloud to list security_group_rule resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/neutron
//
// Discovery hints from registry:
//
//	pagination: false
//	all_tenants: false
//	regions: false
type NeutronSecurityGroupRuleDiscoverer struct{}

func (d *NeutronSecurityGroupRuleDiscoverer) ResourceType() string {
	return "security_group_rule"
}

func (d *NeutronSecurityGroupRuleDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		opts := rules.ListOpts{}
		pages, err := rules.List(client, opts).AllPages()
		if err != nil {
			return
		}

		ruleList, err := rules.ExtractRules(pages)
		if err != nil {
			return
		}

		for _, rule := range ruleList {
			select {
			case <-ctx.Done():
				return
			case ch <- discovery.Job{
				Service:      "neutron",
				ResourceType: "security_group_rule",
				ResourceID:   rule.ID,
				ProjectID:    rule.TenantID,
				Resource:     rule,
			}:
			}
		}
		_ = allTenants
	}()

	return ch, nil
}

// NeutronFloatingIpDiscoverer discovers neutron/floating_ip resources.
type NeutronFloatingIpDiscoverer struct{}

func (d *NeutronFloatingIpDiscoverer) ResourceType() string {
	return "floating_ip"
}

func (d *NeutronFloatingIpDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		pages, err := floatingips.List(client, floatingips.ListOpts{}).AllPages()
		if err != nil {
			return
		}

		fipList, err := floatingips.ExtractFloatingIPs(pages)
		if err != nil {
			return
		}

		for _, fip := range fipList {
			select {
			case <-ctx.Done():
				return
			case ch <- discovery.Job{
				Service:      "neutron",
				ResourceType: "floating_ip",
				ResourceID:   fip.ID,
				ProjectID:    fip.TenantID,
				Resource:     fip,
			}:
			}
		}
	}()

	return ch, nil
}


// NeutronSubnetDiscoverer discovers neutron/subnet resources.
type NeutronSubnetDiscoverer struct{}

func (d *NeutronSubnetDiscoverer) ResourceType() string {
	return "subnet"
}

func (d *NeutronSubnetDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		opts := subnets.ListOpts{}
		pages, err := subnets.List(client, opts).AllPages()
		if err != nil {
			return
		}

		subnetList, err := subnets.ExtractSubnets(pages)
		if err != nil {
			return
		}

		for _, subnet := range subnetList {
			select {
			case <-ctx.Done():
				return
			case ch <- discovery.Job{
				Service:      "neutron",
				ResourceType: "subnet",
				ResourceID:   subnet.ID,
				ProjectID:    subnet.TenantID,
				Resource:     subnet,
			}:
			}
		}
	}()

	return ch, nil
}


// NeutronRouterDiscoverer discovers neutron/router resources.
type NeutronRouterDiscoverer struct{}

func (d *NeutronRouterDiscoverer) ResourceType() string {
	return "router"
}

func (d *NeutronRouterDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		pages, err := routers.List(client, routers.ListOpts{}).AllPages()
		if err != nil {
			return
		}

		routerList, err := routers.ExtractRouters(pages)
		if err != nil {
			return
		}

		for _, router := range routerList {
			select {
			case <-ctx.Done():
				return
			case ch <- discovery.Job{
				Service:      "neutron",
				ResourceType: "router",
				ResourceID:   router.ID,
				ProjectID:    router.TenantID,
				Resource:     router,
			}:
			}
		}
	}()

	return ch, nil
}
