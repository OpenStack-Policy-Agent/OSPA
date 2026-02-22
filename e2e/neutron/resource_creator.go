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
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
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



// CreateSecurityGroup creates a test security_group and returns:
//   - resourceID: The ID of the created resource (for filtering audit results)
//   - cleanup: A function to delete the resource and its dependencies
func CreateSecurityGroup(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
	t.Helper()

	name := fmt.Sprintf("%ssecurity-group-%d", testPrefix, time.Now().UnixNano())

	createOpts := groups.CreateOpts{
		Name:        name,
		Description: "OSPA e2e test security group - safe to delete",
	}

	sg, err := groups.Create(client, createOpts).Extract()
	if err != nil {
		t.Fatalf("Failed to create test security group: %v", err)
	}

	t.Logf("Created test security group: %s (%s)", sg.Name, sg.ID)

	cleanup = func() {
		t.Logf("Cleaning up test security group: %s", sg.ID)
		if err := groups.Delete(client, sg.ID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test security group %s: %v", sg.ID, err)
		}
	}

	return sg.ID, cleanup
}



// CreateSecurityGroupRule creates a test security_group_rule and returns:
//   - resourceID: The ID of the created resource (for filtering audit results)
//   - cleanup: A function to delete the resource and its dependencies
//
// This creates a security group first, then adds a rule to it.
// The cleanup function will delete both the rule and the security group.
func CreateSecurityGroupRule(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
	t.Helper()

	// First create a security group to hold the rule
	sgName := fmt.Sprintf("%ssg-for-rule-%d", testPrefix, time.Now().UnixNano())
	sgOpts := groups.CreateOpts{
		Name:        sgName,
		Description: "OSPA e2e test security group for rule - safe to delete",
	}

	sg, err := groups.Create(client, sgOpts).Extract()
	if err != nil {
		t.Fatalf("Failed to create test security group for rule: %v", err)
	}
	t.Logf("Created test security group for rule: %s (%s)", sg.Name, sg.ID)

	// Create a security group rule - SSH ingress from anywhere (a "dangerous" rule for testing)
	ruleOpts := rules.CreateOpts{
		SecGroupID:     sg.ID,
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   22,
		PortRangeMax:   22,
		RemoteIPPrefix: "0.0.0.0/0",
		Description:    "OSPA e2e test rule - SSH from anywhere - safe to delete",
	}

	rule, err := rules.Create(client, ruleOpts).Extract()
	if err != nil {
		// Clean up the security group on failure
		_ = groups.Delete(client, sg.ID).ExtractErr()
		t.Fatalf("Failed to create test security group rule: %v", err)
	}
	t.Logf("Created test security group rule: %s (direction=%s, protocol=%s, port=%d, remote=%s)",
		rule.ID, rule.Direction, rule.Protocol, rule.PortRangeMin, rule.RemoteIPPrefix)

	cleanup = func() {
		// Delete the rule first
		t.Logf("Cleaning up test security group rule: %s", rule.ID)
		if err := rules.Delete(client, rule.ID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test security group rule %s: %v", rule.ID, err)
		}

		// Then delete the security group
		t.Logf("Cleaning up test security group: %s", sg.ID)
		if err := groups.Delete(client, sg.ID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test security group %s: %v", sg.ID, err)
		}
	}

	return rule.ID, cleanup
}

// CreateSecurityGroupRuleWithOptions creates a security group rule with custom options.
// This is useful for testing specific rule configurations.
func CreateSecurityGroupRuleWithOptions(t *testing.T, client *gophercloud.ServiceClient, opts rules.CreateOpts) (ruleID, sgID string, cleanup func()) {
	t.Helper()

	// If no security group ID is provided, create one
	if opts.SecGroupID == "" {
		sgName := fmt.Sprintf("%ssg-for-rule-%d", testPrefix, time.Now().UnixNano())
		sgOpts := groups.CreateOpts{
			Name:        sgName,
			Description: "OSPA e2e test security group for rule - safe to delete",
		}

		sg, err := groups.Create(client, sgOpts).Extract()
		if err != nil {
			t.Fatalf("Failed to create test security group for rule: %v", err)
		}
		t.Logf("Created test security group for rule: %s (%s)", sg.Name, sg.ID)
		opts.SecGroupID = sg.ID
	}

	rule, err := rules.Create(client, opts).Extract()
	if err != nil {
		// Clean up the security group on failure
		_ = groups.Delete(client, opts.SecGroupID).ExtractErr()
		t.Fatalf("Failed to create test security group rule: %v", err)
	}
	t.Logf("Created test security group rule: %s", rule.ID)

	cleanup = func() {
		t.Logf("Cleaning up test security group rule: %s", rule.ID)
		if err := rules.Delete(client, rule.ID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test security group rule %s: %v", rule.ID, err)
		}

		t.Logf("Cleaning up test security group: %s", opts.SecGroupID)
		if err := groups.Delete(client, opts.SecGroupID).ExtractErr(); err != nil {
			t.Logf("Warning: failed to delete test security group %s: %v", opts.SecGroupID, err)
		}
	}

	return rule.ID, opts.SecGroupID, cleanup
}



// CreateSubnet creates a test subnet (and its parent network) and returns:
//   - resourceID: The subnet ID (for filtering audit results)
//   - cleanup: A function to delete the subnet and its parent network
func CreateSubnet(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
	t.Helper()
	_, subnetID, cleanupAll := CreateNetworkWithSubnet(t, client)
	return subnetID, cleanupAll
}

// CleanupOrphans deletes any leaked test resources (those with testPrefix).
// Run this manually if tests fail and leave resources behind:
//
//	go test -tags=e2e ./e2e/neutron/... -run TestCleanup
func CleanupOrphans(t *testing.T, client *gophercloud.ServiceClient) {
	t.Helper()

	ctx := context.Background()
	_ = ctx // For future use with context-aware operations

	// Clean up orphaned security groups first (before networks, as SGs may be attached to ports)
	t.Log("Searching for orphaned test security groups...")
	sgPages, err := groups.List(client, groups.ListOpts{}).AllPages()
	if err != nil {
		t.Logf("Warning: failed to list security groups: %v", err)
	} else {
		allSGs, err := groups.ExtractGroups(sgPages)
		if err != nil {
			t.Logf("Warning: failed to extract security groups: %v", err)
		} else {
			var sgDeleted int
			for _, sg := range allSGs {
				if strings.HasPrefix(sg.Name, testPrefix) {
					t.Logf("Found orphaned security group: %s (%s)", sg.Name, sg.ID)
					if err := groups.Delete(client, sg.ID).ExtractErr(); err != nil {
						t.Logf("  Warning: failed to delete security group %s: %v", sg.ID, err)
					} else {
						sgDeleted++
					}
				}
			}
			t.Logf("Cleanup complete: deleted %d orphaned security groups", sgDeleted)
		}
	}

	// Clean up orphaned networks
	t.Log("Searching for orphaned test networks...")
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
