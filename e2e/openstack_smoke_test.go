//go:build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
)

// OpenStack e2e smoke test.
// It authenticates via clouds.yaml (OS_CLOUD) and performs a minimal Nova API call.
//
// Run:
//   OS_CLOUD=mycloud go test -tags=e2e ./...
func TestOpenStack_Smoke_AuthAndListServersFirstPage(t *testing.T) {
	client := getComputeClient(t)

	// Fetch only the first page and stop early to keep it fast and safe.
	pager := servers.List(client, servers.ListOpts{
		AllTenants: os.Getenv("OSPA_E2E_ALL_TENANTS") == "true",
	})

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		_, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		return false, nil // stop after first page
	})
	if err != nil {
		t.Fatalf("servers.List first page failed: %v", err)
	}
}


