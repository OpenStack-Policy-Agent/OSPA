package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)


// NovaInstanceDiscoverer discovers nova resources of type instance.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing nova instance resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
type NovaInstanceDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *NovaInstanceDiscoverer) ResourceType() string {
	return "instance"
}

// Discover discovers resources and sends them to the returned channel
func (d *NovaInstanceDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}


// NovaKeypairDiscoverer discovers nova resources of type keypair.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing nova keypair resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
type NovaKeypairDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *NovaKeypairDiscoverer) ResourceType() string {
	return "keypair"
}

// Discover discovers resources and sends them to the returned channel
func (d *NovaKeypairDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}


