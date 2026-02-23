//go:build e2e

// Package cinder contains e2e tests for the Cinder service.
//
// =============================================================================
// RESOURCE CREATOR - READ THIS FIRST
// =============================================================================
//
// This file provides helper functions to create test resources for e2e tests.
// Each resource may have dependencies that must be created first.
//
// HOW TO USE:
// 1. Implement the Create<Resource>() functions below
// 2. Each function should create the resource AND its dependencies
// 3. Return a cleanup function that deletes resources in reverse order
// 4. Use these functions in the corresponding <resource>_test.go files
//
// DEPENDENCY GRAPH FOR Cinder:
// =============================================================================

// Volume:
//   Description: Block storage volumes
//   Gophercloud: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
//   OpenStack API: https://docs.openstack.org/api-ref/cinder

// Snapshot:
//   Description: Volume snapshots
//   Gophercloud: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
//   OpenStack API: https://docs.openstack.org/api-ref/cinder

// =============================================================================

package cinder

import (
	"testing"

	"github.com/gophercloud/gophercloud"
	// TODO: Import the specific gophercloud packages you need:
	// "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	// "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
)

const testPrefix = "ospa-e2e-"

// =============================================================================
// RESOURCE CREATORS - IMPLEMENT THESE
// =============================================================================


// CreateVolume creates a test volume and returns:
//   - resourceID: The ID of the created resource (for filtering audit results)
//   - cleanup: A function to delete the resource and its dependencies
func CreateVolume(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
	t.Helper()
	
	// TODO: Implement resource creation
	// See the example above and the gophercloud documentation
	
	t.Skip("CreateVolume not implemented - implement in resource_creator.go")
	return "", func() {}
}


// CreateSnapshot creates a test snapshot and returns:
//   - resourceID: The ID of the created resource (for filtering audit results)
//   - cleanup: A function to delete the resource and its dependencies
func CreateSnapshot(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
	t.Helper()
	
	// TODO: Implement resource creation
	// See the example above and the gophercloud documentation
	
	t.Skip("CreateSnapshot not implemented - implement in resource_creator.go")
	return "", func() {}
}


// =============================================================================
// CLEANUP HELPER
// =============================================================================

// CleanupOrphans deletes any leaked test resources (those with testPrefix).
// Run this manually if tests fail and leave resources behind:
//   go test -tags=e2e ./e2e/cinder/... -run TestCleanupOrphans
func CleanupOrphans(t *testing.T, client *gophercloud.ServiceClient) {
	t.Helper()
	
	// TODO: Implement cleanup for orphaned resources
	// List all resources, filter by testPrefix, delete them
	
	t.Log("TODO: Implement orphan cleanup")
}
