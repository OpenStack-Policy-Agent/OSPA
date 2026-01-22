package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

// CinderVolumeDiscoverer discovers cinder resources of type volume.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing cinder volume resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
type CinderVolumeDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *CinderVolumeDiscoverer) ResourceType() string {
	return "volume"
}

// Discover discovers resources and sends them to the returned channel
func (d *CinderVolumeDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}

// CinderSnapshotDiscoverer discovers cinder resources of type snapshot.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing cinder snapshot resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
type CinderSnapshotDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *CinderSnapshotDiscoverer) ResourceType() string {
	return "snapshot"
}

// Discover discovers resources and sends them to the returned channel
func (d *CinderSnapshotDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}
