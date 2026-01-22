package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
	// TODO: Import the correct gophercloud package for cinder.
	// Example for Nova: "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	// See: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
)


// CinderVolumeDiscoverer discovers cinder/volume resources.
//
// TODO: Implement Discover() using gophercloud to list volume resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/cinder
//
// Discovery hints from registry:
//   pagination: false
//   all_tenants: false
//   regions: false
type CinderVolumeDiscoverer struct{}

func (d *CinderVolumeDiscoverer) ResourceType() string {
	return "volume"
}

func (d *CinderVolumeDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List volume resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "cinder", ResourceType: "volume", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

	return ch, nil
}


// CinderSnapshotDiscoverer discovers cinder/snapshot resources.
//
// TODO: Implement Discover() using gophercloud to list snapshot resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/cinder
//
// Discovery hints from registry:
//   pagination: false
//   all_tenants: false
//   regions: false
type CinderSnapshotDiscoverer struct{}

func (d *CinderSnapshotDiscoverer) ResourceType() string {
	return "snapshot"
}

func (d *CinderSnapshotDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List snapshot resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "cinder", ResourceType: "snapshot", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

	return ch, nil
}


