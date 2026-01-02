package discovery

import (
	"context"

	"github.com/gophercloud/gophercloud"
)

// Discoverer discovers resources of a specific type
type Discoverer interface {
	// Discover discovers resources and sends them to the returned channel
	Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error)

	// ResourceType returns the resource type this discoverer handles
	ResourceType() string
}

