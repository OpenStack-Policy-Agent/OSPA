package services

import (
	"fmt"

	rootservices "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/blockstorage"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	discovery_services "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
	"github.com/gophercloud/gophercloud"
)

// BlockStorageService implements the Service interface for Cinder (Block Storage)
//
// Supported resources:
//   - volume: Block storage volumes
//   - snapshot: Volume snapshots
//
// To add support for a new resource type:
//   1. Create a discoverer in pkg/discovery/services/blockstorage.go
//   2. Create an auditor in pkg/audit/blockstorage/
//   3. Add cases in GetResourceAuditor() and GetResourceDiscoverer()
type BlockStorageService struct{}

func init() {
	rootservices.MustRegister(&BlockStorageService{})
	// Register supported resources for automatic validation
	rootservices.RegisterResource("cinder", "volume")
	rootservices.RegisterResource("cinder", "snapshot")
}

// Name returns the service name
func (s *BlockStorageService) Name() string {
	return "cinder"
}

// GetClient returns a Cinder block storage client
func (s *BlockStorageService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetBlockStorageClient()
}

// GetResourceAuditor returns an auditor for the given resource type
func (s *BlockStorageService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "volume":
		return &blockstorage.VolumeAuditor{}, nil
	case "snapshot":
		return &blockstorage.SnapshotAuditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

// GetResourceDiscoverer returns a discoverer for the given resource type
func (s *BlockStorageService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "volume":
		return &discovery_services.BlockStorageVolumeDiscoverer{}, nil
	case "snapshot":
		return &discovery_services.BlockStorageSnapshotDiscoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

