package services

import (
	"fmt"

	rootservices "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/cinder"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	discovery_services "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
	"github.com/gophercloud/gophercloud"
)

// CinderService implements the Service interface for Cinder
//
// Supported resources:
//   - volume: Cinder volume resources
//   - snapshot: Cinder snapshot resources
//
// TODO(OSPA): Ensure pkg/auth/auth.go has GetCinderClient() that returns a
// *gophercloud.ServiceClient for service type "volumev3". The scaffold tool adds this
// method automatically, but verify it is correct for your cloud/provider.
//
// To add support for a new resource type:
//   1. Create a discoverer in pkg/discovery/services/cinder.go
//   2. Create an auditor in pkg/audit/cinder/
//   3. Add cases in GetResourceAuditor() and GetResourceDiscoverer() below
//   4. Register the resource in init() using RegisterResource()
type CinderService struct{}

func init() {
	rootservices.MustRegister(&CinderService{})
	// Register all supported resources for automatic validation
	rootservices.RegisterResource("cinder", "volume")
	rootservices.RegisterResource("cinder", "snapshot")
}

// Name returns the service name
func (s *CinderService) Name() string {
	return "cinder"
}

// GetClient returns an authenticated service client
func (s *CinderService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetCinderClient()
}

// GetResourceAuditor returns an auditor for the given resource type
func (s *CinderService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "volume":
		return &cinder.VolumeAuditor{}, nil
	case "snapshot":
		return &cinder.SnapshotAuditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

// GetResourceDiscoverer returns a discoverer for the given resource type
func (s *CinderService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "volume":
		return &discovery_services.CinderVolumeDiscoverer{}, nil
	case "snapshot":
		return &discovery_services.CinderSnapshotDiscoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}
