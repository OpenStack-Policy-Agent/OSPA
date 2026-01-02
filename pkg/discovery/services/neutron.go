package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)


// NeutronSecurityGroupRuleDiscoverer discovers neutron resources of type security_group_rule.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing neutron security_group_rule resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
type NeutronSecurityGroupRuleDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *NeutronSecurityGroupRuleDiscoverer) ResourceType() string {
	return "security_group_rule"
}

// Discover discovers resources and sends them to the returned channel
func (d *NeutronSecurityGroupRuleDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}


// NeutronFloatingIpDiscoverer discovers neutron resources of type floating_ip.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing neutron floating_ip resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
type NeutronFloatingIpDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *NeutronFloatingIpDiscoverer) ResourceType() string {
	return "floating_ip"
}

// Discover discovers resources and sends them to the returned channel
func (d *NeutronFloatingIpDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}


// NeutronSecurityGroupDiscoverer discovers neutron resources of type security_group.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing neutron security_group resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
type NeutronSecurityGroupDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *NeutronSecurityGroupDiscoverer) ResourceType() string {
	return "security_group"
}

// Discover discovers resources and sends them to the returned channel
func (d *NeutronSecurityGroupDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}


