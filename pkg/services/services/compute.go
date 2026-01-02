package services

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/compute"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

// ComputeService implements the Service interface for Nova (Compute)
// 
// Supported resources:
//   - instance: Nova server instances
//   - keypair: SSH keypairs
//
// To add support for a new resource type:
//   1. Create a discoverer in pkg/discovery/services/compute.go
//   2. Create an auditor in pkg/audit/compute/
//   3. Add cases in GetResourceAuditor() and GetResourceDiscoverer()
type ComputeService struct{}

func init() {
	MustRegister(&ComputeService{})
	// Register supported resources for automatic validation
	RegisterResource("nova", "instance")
	RegisterResource("nova", "keypair")
}

// Name returns the service name
func (s *ComputeService) Name() string {
	return "nova"
}

// GetClient returns a Nova compute client
func (s *ComputeService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetComputeClient()
}

// GetResourceAuditor returns an auditor for the given resource type
func (s *ComputeService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "instance":
		return &compute.InstanceAuditor{}, nil
	case "keypair":
		return &compute.KeypairAuditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

// GetResourceDiscoverer returns a discoverer for the given resource type
func (s *ComputeService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "instance":
		return &discovery.ComputeInstanceDiscoverer{}, nil
	case "keypair":
		return &discovery.ComputeKeypairDiscoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

