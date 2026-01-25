//go:build e2e

// Package neutron contains e2e tests for the Neutron service.
package neutron

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

const testPrefix = "ospa-e2e-"

// =============================================================================
// RESOURCE CREATORS
// =============================================================================

// CreateNetwork creates a test network and returns the network ID and cleanup function.
func CreateNetwork(t *testing.T, client *gophercloud.ServiceClient) (networkID string, cleanup func()) {
	t.Helper()

	name := fmt.Sprintf("%snetwork-%d", testPrefix, time.Now().UnixNano())
	adminStateUp := true

	createOpts := networks.CreateOpts{
		Name:         name,
		AdminStateUp: &adminStateUp,
		Description:  "OSPA e2e test network - safe to delete",
	}

	network, err := networks.Create(client, createOpts).Extract()
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}

	t.Logf("Created test network: %s (%s)", network.Name, network.ID)

	cleanup = func() {
		t.Logf("Cleaning up test network: %s", network.ID)
		if err := networks.Delete(client, network.ID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test network %s: %v", network.ID, err)
		}
	}

	return network.ID, cleanup
}

// CreateNetworkWithSubnet creates a test network with a subnet attached.
// Use this when testing resources that require a subnet (ports, routers, etc.)
func CreateNetworkWithSubnet(t *testing.T, client *gophercloud.ServiceClient) (networkID, subnetID string, cleanup func()) {
	t.Helper()

	// Create network first
	networkName := fmt.Sprintf("%snetwork-%d", testPrefix, time.Now().UnixNano())
	adminStateUp := true
	networkOpts := networks.CreateOpts{
		Name:         networkName,
		AdminStateUp: &adminStateUp,
		Description:  "OSPA e2e test network with subnet - safe to delete",
	}

	network, err := networks.Create(client, networkOpts).Extract()
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}
	t.Logf("Created test network: %s (%s)", network.Name, network.ID)

	// Create subnet
	subnetName := fmt.Sprintf("%ssubnet-%d", testPrefix, time.Now().UnixNano())
	cidr := generateCIDR()
	subnetOpts := subnets.CreateOpts{
		Name:      subnetName,
		NetworkID: network.ID,
		CIDR:      cidr,
		IPVersion: gophercloud.IPv4,
	}

	subnet, err := subnets.Create(client, subnetOpts).Extract()
	if err != nil {
		// Cleanup network on subnet creation failure
		_ = networks.Delete(client, network.ID).ExtractErr()
		t.Fatalf("Failed to create test subnet: %v", err)
	}
	t.Logf("Created test subnet: %s (%s) with CIDR %s", subnet.Name, subnet.ID, cidr)

	cleanup = func() {
		// Delete subnet first, then network
		t.Logf("Cleaning up test subnet: %s", subnet.ID)
		if err := subnets.Delete(client, subnet.ID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test subnet %s: %v", subnet.ID, err)
		}

		t.Logf("Cleaning up test network: %s", network.ID)
		if err := networks.Delete(client, network.ID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test network %s: %v", network.ID, err)
		}
	}

	return network.ID, subnet.ID, cleanup
}

// generateCIDR generates a unique CIDR to avoid conflicts.
func generateCIDR() string {
	now := time.Now().UnixNano()
	second := (now / 1e9) % 256
	third := (now / 1e6) % 256
	return fmt.Sprintf("10.%d.%d.0/24", second, third)
}

// =============================================================================
// CLEANUP HELPER
// =============================================================================

// CleanupOrphans deletes any leaked test resources (those with testPrefix).
// Run this manually if tests fail and leave resources behind:
//
//	go test -tags=e2e ./e2e/neutron/... -run TestCleanup
func CleanupOrphans(t *testing.T, client *gophercloud.ServiceClient) {
	t.Helper()

	t.Log("Searching for orphaned test networks...")
	ctx := context.Background()
	_ = ctx // For future use with context-aware operations

	// List all networks
	allPages, err := networks.List(client, networks.ListOpts{}).AllPages()
	if err != nil {
		t.Fatalf("Failed to list networks: %v", err)
	}

	allNetworks, err := networks.ExtractNetworks(allPages)
	if err != nil {
		t.Fatalf("Failed to extract networks: %v", err)
	}

	// Find and delete orphaned test networks
	var deleted int
	for _, network := range allNetworks {
		if strings.HasPrefix(network.Name, testPrefix) {
			t.Logf("Found orphaned network: %s (%s)", network.Name, network.ID)

			// First delete any subnets
			for _, subnetID := range network.Subnets {
				t.Logf("  Deleting subnet: %s", subnetID)
				if err := subnets.Delete(client, subnetID).ExtractErr(); err != nil {
					t.Logf("  Warning: failed to delete subnet %s: %v", subnetID, err)
				}
			}

			// Then delete the network
			if err := networks.Delete(client, network.ID).ExtractErr(); err != nil {
				t.Logf("  Warning: failed to delete network %s: %v", network.ID, err)
			} else {
				deleted++
			}
		}
	}

	t.Logf("Cleanup complete: deleted %d orphaned networks", deleted)
}
