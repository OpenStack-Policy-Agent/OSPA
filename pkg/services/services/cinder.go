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

// CinderService implements the Service interface for OpenStack Cinder.
//
// Supported resources:
//   - volume: Block storage volumes
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
//   - snapshot: Volume snapshots
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
type CinderService struct{}

func init() {
	rootservices.MustRegister(&CinderService{})
	rootservices.RegisterResource("cinder", "volume")
	rootservices.RegisterResource("cinder", "snapshot")
}

func (s *CinderService) Name() string {
	return "cinder"
}

func (s *CinderService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetCinderClient()
}

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
