package services

import (
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

// Service represents an OpenStack service
type Service interface {
	// Name returns the service name (e.g., "neutron", "nova", "cinder")
	Name() string

	// GetClient returns an authenticated service client
	GetClient(session *auth.Session) (*gophercloud.ServiceClient, error)

	// GetResourceAuditor returns an auditor for the given resource type
	GetResourceAuditor(resourceType string) (audit.Auditor, error)

	// GetResourceDiscoverer returns a discoverer for the given resource type
	GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error)
}
