package services

import (
	"context"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/pagination"
)

// ComputeInstanceDiscoverer discovers Nova instances/servers
type ComputeInstanceDiscoverer struct{}

func (d *ComputeInstanceDiscoverer) ResourceType() string {
	return "instance"
}

func (d *ComputeInstanceDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	opts := servers.ListOpts{
		AllTenants: allTenants,
	}
	pager := servers.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(serverList))
		for i := range serverList {
			resources[i] = serverList[i]
		}
		return resources, nil
	}

	createJob := discovery.SimpleJobCreator(
		"nova",
		func(r interface{}) string {
			return r.(servers.Server).ID
		},
		func(r interface{}) string {
			return r.(servers.Server).TenantID
		},
	)

	return discovery.DiscoverPaged(ctx, client, "nova", d.ResourceType(), pager, extract, createJob)
}

// ComputeKeypairDiscoverer discovers Nova keypairs
type ComputeKeypairDiscoverer struct{}

func (d *ComputeKeypairDiscoverer) ResourceType() string {
	return "keypair"
}

func (d *ComputeKeypairDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	opts := keypairs.ListOpts{}
	// Note: keypairs.List doesn't support allTenants directly
	// This would need to be handled differently if cross-tenant scanning is needed
	pager := keypairs.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		keypairList, err := keypairs.ExtractKeyPairs(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(keypairList))
		for i := range keypairList {
			resources[i] = keypairList[i]
		}
		return resources, nil
	}

	createJob := discovery.SimpleJobCreator(
		"nova",
		func(r interface{}) string {
			return r.(keypairs.KeyPair).Name
		},
		func(r interface{}) string {
			// Keypairs don't have TenantID, use UserID
			return r.(keypairs.KeyPair).UserID
		},
	)

	return discovery.DiscoverPaged(ctx, client, "nova", d.ResourceType(), pager, extract, createJob)
}

