package services

import (
	"fmt"

	rootservices "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/nova"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	discovery_services "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
	"github.com/gophercloud/gophercloud"
)

// NovaService implements the Service interface for Nova
//
// Supported resources:
//   - instance: Nova instance resources
//   - keypair: Nova keypair resources
//
// TODO(OSPA): Ensure pkg/auth/auth.go has GetNovaClient() that returns a
// *gophercloud.ServiceClient for service type "compute". The scaffold tool adds this
// method automatically, but verify it is correct for your cloud/provider.
//
// To add support for a new resource type:
//   1. Create a discoverer in pkg/discovery/services/nova.go
//   2. Create an auditor in pkg/audit/nova/
//   3. Add cases in GetResourceAuditor() and GetResourceDiscoverer() below
//   4. Register the resource in init() using RegisterResource()
type NovaService struct{}

func init() {
	rootservices.MustRegister(&NovaService{})
	// Register all supported resources for automatic validation
	rootservices.RegisterResource("nova", "instance")
	rootservices.RegisterResource("nova", "keypair")
}

// Name returns the service name
func (s *NovaService) Name() string {
	return "nova"
}

// GetClient returns an authenticated service client
func (s *NovaService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetNovaClient()
}

// GetResourceAuditor returns an auditor for the given resource type
func (s *NovaService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "instance":
		return &nova.InstanceAuditor{}, nil
	case "keypair":
		return &nova.KeypairAuditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

// GetResourceDiscoverer returns a discoverer for the given resource type
func (s *NovaService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "instance":
		return &discovery_services.NovaInstanceDiscoverer{}, nil
	case "keypair":
		return &discovery_services.NovaKeypairDiscoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}
