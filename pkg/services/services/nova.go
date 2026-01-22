package services

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/nova"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	discovery_services "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
	rootservices "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/gophercloud/gophercloud"
)

// NovaService implements the Service interface for OpenStack Nova.
//
// Supported resources:
//   - instance: Server instances
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
//   - keypair: SSH keypairs
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
type NovaService struct{}

func init() {
	rootservices.MustRegister(&NovaService{})
	rootservices.RegisterResource("nova", "instance")
	rootservices.RegisterResource("nova", "keypair")
}

func (s *NovaService) Name() string {
	return "nova"
}

func (s *NovaService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetNovaClient()
}

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
