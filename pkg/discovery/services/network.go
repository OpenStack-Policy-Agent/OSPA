package services

import (
	"context"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/pagination"
)

// NetworkSecurityGroupRuleDiscoverer discovers security group rules
type NetworkSecurityGroupRuleDiscoverer struct{}

func (d *NetworkSecurityGroupRuleDiscoverer) ResourceType() string {
	return "security_group_rule"
}

func (d *NetworkSecurityGroupRuleDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	opts := rules.ListOpts{}
	if allTenants {
		opts.TenantID = ""
	}
	pager := rules.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		ruleList, err := rules.ExtractRules(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(ruleList))
		for i := range ruleList {
			resources[i] = ruleList[i]
		}
		return resources, nil
	}

	createJob := discovery.SimpleJobCreator(
		"neutron",
		func(r interface{}) string {
			return r.(rules.SecGroupRule).ID
		},
		func(r interface{}) string {
			return r.(rules.SecGroupRule).TenantID
		},
	)

	return discovery.DiscoverPaged(ctx, client, "neutron", d.ResourceType(), pager, extract, createJob)
}

// NetworkFloatingIPDiscoverer discovers floating IPs
type NetworkFloatingIPDiscoverer struct{}

func (d *NetworkFloatingIPDiscoverer) ResourceType() string {
	return "floating_ip"
}

func (d *NetworkFloatingIPDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	opts := floatingips.ListOpts{}
	if allTenants {
		opts.TenantID = ""
	}
	pager := floatingips.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		fipList, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(fipList))
		for i := range fipList {
			resources[i] = fipList[i]
		}
		return resources, nil
	}

	createJob := discovery.SimpleJobCreator(
		"neutron",
		func(r interface{}) string {
			return r.(floatingips.FloatingIP).ID
		},
		func(r interface{}) string {
			return r.(floatingips.FloatingIP).TenantID
		},
	)

	return discovery.DiscoverPaged(ctx, client, "neutron", d.ResourceType(), pager, extract, createJob)
}

// NetworkSecurityGroupDiscoverer discovers security groups
type NetworkSecurityGroupDiscoverer struct{}

func (d *NetworkSecurityGroupDiscoverer) ResourceType() string {
	return "security_group"
}

func (d *NetworkSecurityGroupDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	opts := groups.ListOpts{}
	if allTenants {
		opts.TenantID = ""
	}
	pager := groups.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		sgList, err := groups.ExtractGroups(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(sgList))
		for i := range sgList {
			resources[i] = sgList[i]
		}
		return resources, nil
	}

	createJob := discovery.SimpleJobCreator(
		"neutron",
		func(r interface{}) string {
			return r.(groups.SecGroup).ID
		},
		func(r interface{}) string {
			return r.(groups.SecGroup).TenantID
		},
	)

	return discovery.DiscoverPaged(ctx, client, "neutron", d.ResourceType(), pager, extract, createJob)
}

