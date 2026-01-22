package services

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/neutron"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	discovery_services "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
	rootservices "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/gophercloud/gophercloud"
)

// NeutronService implements the Service interface for Neutron
//
// Supported resources:
//   - security_group_rule: Neutron security_group_rule resources
//   - floating_ip: Neutron floating_ip resources
//   - security_group: Neutron security_group resources
//
// TODO(OSPA): Ensure pkg/auth/auth.go has GetNeutronClient() that returns a
// *gophercloud.ServiceClient for service type "network". The scaffold tool adds this
// method automatically, but verify it is correct for your cloud/provider.
//
// To add support for a new resource type:
//  1. Create a discoverer in pkg/discovery/services/neutron.go
//  2. Create an auditor in pkg/audit/neutron/
//  3. Add cases in GetResourceAuditor() and GetResourceDiscoverer() below
//  4. Register the resource in init() using RegisterResource()
type NeutronService struct{}

func init() {
	rootservices.MustRegister(&NeutronService{})
	// Register all supported resources for automatic validation
	rootservices.RegisterResource("neutron", "security_group_rule")
	rootservices.RegisterResource("neutron", "floating_ip")
	rootservices.RegisterResource("neutron", "security_group")
}

// Name returns the service name
func (s *NeutronService) Name() string {
	return "neutron"
}

// GetClient returns an authenticated service client
func (s *NeutronService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetNeutronClient()
}

// GetResourceAuditor returns an auditor for the given resource type
func (s *NeutronService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "security_group_rule":
		return &neutron.SecurityGroupRuleAuditor{}, nil
	case "floating_ip":
		return &neutron.FloatingIpAuditor{}, nil
	case "security_group":
		return &neutron.SecurityGroupAuditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

// GetResourceDiscoverer returns a discoverer for the given resource type
func (s *NeutronService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "security_group_rule":
		return &discovery_services.NeutronSecurityGroupRuleDiscoverer{}, nil
	case "floating_ip":
		return &discovery_services.NeutronFloatingIpDiscoverer{}, nil
	case "security_group":
		return &discovery_services.NeutronSecurityGroupDiscoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}
