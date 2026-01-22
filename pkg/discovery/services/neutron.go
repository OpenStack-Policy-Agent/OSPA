package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
	// TODO: Import the correct gophercloud package for neutron.
	// Example for Nova: "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	// See: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
)


// NeutronSecurityGroupDiscoverer discovers neutron/security_group resources.
//
// TODO: Implement Discover() using gophercloud to list security_group resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/neutron
//
// Discovery hints from registry:
//   pagination: false
//   all_tenants: false
//   regions: false
type NeutronSecurityGroupDiscoverer struct{}

func (d *NeutronSecurityGroupDiscoverer) ResourceType() string {
	return "security_group"
}

func (d *NeutronSecurityGroupDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List security_group resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "neutron", ResourceType: "security_group", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
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
//   pagination: false
//   all_tenants: false
//   regions: false
type NeutronSecurityGroupRuleDiscoverer struct{}

func (d *NeutronSecurityGroupRuleDiscoverer) ResourceType() string {
	return "security_group_rule"
}

func (d *NeutronSecurityGroupRuleDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List security_group_rule resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "neutron", ResourceType: "security_group_rule", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

	return ch, nil
}


// NeutronFloatingIpDiscoverer discovers neutron/floating_ip resources.
//
// TODO: Implement Discover() using gophercloud to list floating_ip resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/neutron
//
// Discovery hints from registry:
//   pagination: false
//   all_tenants: false
//   regions: false
type NeutronFloatingIpDiscoverer struct{}

func (d *NeutronFloatingIpDiscoverer) ResourceType() string {
	return "floating_ip"
}

func (d *NeutronFloatingIpDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List floating_ip resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "neutron", ResourceType: "floating_ip", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

	return ch, nil
}


