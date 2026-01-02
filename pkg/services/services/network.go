package services

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/network"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

// NetworkService implements the Service interface for Neutron (Network)
//
// Supported resources:
//   - security_group_rule: Security group ingress/egress rules
//   - floating_ip: Floating IP addresses
//   - security_group: Security groups
//
// To add support for a new resource type:
//   1. Create a discoverer in pkg/discovery/services/network.go
//   2. Create an auditor in pkg/audit/network/
//   3. Add cases in GetResourceAuditor() and GetResourceDiscoverer()
type NetworkService struct{}

func init() {
	MustRegister(&NetworkService{})
	// Register supported resources for automatic validation
	RegisterResource("neutron", "security_group_rule")
	RegisterResource("neutron", "floating_ip")
	RegisterResource("neutron", "security_group")
}

// Name returns the service name
func (s *NetworkService) Name() string {
	return "neutron"
}

// GetClient returns a Neutron network client
func (s *NetworkService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetNetworkClient()
}

// GetResourceAuditor returns an auditor for the given resource type
func (s *NetworkService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "security_group_rule":
		return &network.SecurityGroupRuleAuditor{}, nil
	case "floating_ip":
		return &network.FloatingIPAuditor{}, nil
	case "security_group":
		return &network.SecurityGroupAuditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

// GetResourceDiscoverer returns a discoverer for the given resource type
func (s *NetworkService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "security_group_rule":
		return &discovery.NetworkSecurityGroupRuleDiscoverer{}, nil
	case "floating_ip":
		return &discovery.NetworkFloatingIPDiscoverer{}, nil
	case "security_group":
		return &discovery.NetworkSecurityGroupDiscoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

