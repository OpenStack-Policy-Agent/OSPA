package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
	// TODO: Import the correct gophercloud package for nova.
	// Example for Nova: "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	// See: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
)


// NovaInstanceDiscoverer discovers nova/instance resources.
//
// TODO: Implement Discover() using gophercloud to list instance resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/nova
//
// Discovery hints from registry:
//   pagination: false
//   all_tenants: false
//   regions: false
type NovaInstanceDiscoverer struct{}

func (d *NovaInstanceDiscoverer) ResourceType() string {
	return "instance"
}

func (d *NovaInstanceDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List instance resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "nova", ResourceType: "instance", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

	return ch, nil
}


// NovaKeypairDiscoverer discovers nova/keypair resources.
//
// TODO: Implement Discover() using gophercloud to list keypair resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/nova
//
// Discovery hints from registry:
//   pagination: false
//   all_tenants: false
//   regions: false
type NovaKeypairDiscoverer struct{}

func (d *NovaKeypairDiscoverer) ResourceType() string {
	return "keypair"
}

func (d *NovaKeypairDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List keypair resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "nova", ResourceType: "keypair", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

	return ch, nil
}


