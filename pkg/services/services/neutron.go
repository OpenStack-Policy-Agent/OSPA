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

// NeutronService implements the Service interface for OpenStack Neutron.
//
// Supported resources:
//   - network: Networks
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
//   - security_group: Security groups
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
//   - security_group_rule: Security group rules
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
//   - floating_ip: Floating IP addresses
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
//   - subnet: Subnets
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
type NeutronService struct{}

func init() {
	rootservices.MustRegister(&NeutronService{})
	rootservices.RegisterResource("neutron", "network")
	rootservices.RegisterResource("neutron", "security_group")
	rootservices.RegisterResource("neutron", "security_group_rule")
	rootservices.RegisterResource("neutron", "floating_ip")
	rootservices.RegisterResource("neutron", "subnet")
}

func (s *NeutronService) Name() string {
	return "neutron"
}

func (s *NeutronService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetNeutronClient()
}

func (s *NeutronService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "network":
		return &neutron.NetworkAuditor{}, nil
	case "security_group":
		return &neutron.SecurityGroupAuditor{}, nil
	case "security_group_rule":
		return &neutron.SecurityGroupRuleAuditor{}, nil
	case "floating_ip":
		return &neutron.FloatingIpAuditor{}, nil
	case "subnet":
		return &neutron.SubnetAuditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

func (s *NeutronService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "network":
		return &discovery_services.NeutronNetworkDiscoverer{}, nil
	case "security_group":
		return &discovery_services.NeutronSecurityGroupDiscoverer{}, nil
	case "security_group_rule":
		return &discovery_services.NeutronSecurityGroupRuleDiscoverer{}, nil
	case "floating_ip":
		return &discovery_services.NeutronFloatingIpDiscoverer{}, nil
	case "subnet":
		return &discovery_services.NeutronSubnetDiscoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}
