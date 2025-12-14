package auth

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

// Session holds the authenticated provider client and configuration options.
type Session struct {
	Provider *gophercloud.ProviderClient
	CloudName string
	Region    string
}

// NewSession creates a new OpenStack session based on a cloud name found in clouds.yaml
// If cloudName is empty, it looks for OS_CLOUD env var or standard env vars.
func NewSession(cloudName string) (*Session, error) {
	opts := &clientconfig.ClientOpts{
		Cloud: cloudName,
	}

	// This helper function looks for clouds.yaml in standard locations 
	// (~/.config/openstack, /etc/openstack, current dir)
	provider, err := clientconfig.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return &Session{
		Provider: provider,
		CloudName: cloudName,
	}, nil
}

// GetComputeClient returns a client for Nova (Compute)
func (s *Session) GetComputeClient() (*gophercloud.ServiceClient, error) {
	// clientconfig handles finding the right endpoint (public/internal) and region automatically
	client, err := clientconfig.NewServiceClient("compute", &clientconfig.ClientOpts{
		Cloud: s.CloudName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}
	return client, nil
}

// GetNetworkClient returns a client for Neutron (Network)
func (s *Session) GetNetworkClient() (*gophercloud.ServiceClient, error) {
	client, err := clientconfig.NewServiceClient("network", &clientconfig.ClientOpts{
		Cloud: s.CloudName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create network client: %w", err)
	}
	return client, nil
}
