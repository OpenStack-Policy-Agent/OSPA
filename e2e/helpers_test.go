//go:build e2e

package e2e

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

var (
	computeOnce   sync.Once
	computeClient *gophercloud.ServiceClient
	computeErr    error
)

// getComputeClient authenticates once per `go test` invocation and returns a Nova client.
func getComputeClient(t *testing.T) *gophercloud.ServiceClient {
	t.Helper()

	cloud := os.Getenv("OS_CLOUD")
	if cloud == "" {
		t.Skip("OS_CLOUD not set; skipping OpenStack e2e tests")
	}

	computeOnce.Do(func() {
		session, err := auth.NewSession(cloud)
		if err != nil {
			computeErr = fmt.Errorf("auth.NewSession(%q): %w", cloud, err)
			return
		}
		c, err := session.GetComputeClient()
		if err != nil {
			computeErr = fmt.Errorf("GetComputeClient: %w", err)
			return
		}
		computeClient = c
	})

	if computeErr != nil {
		t.Fatalf("compute client init failed: %v", computeErr)
	}
	return computeClient
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("%s not set; skipping OpenStack e2e test", key)
	}
	return v
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func createServer(t *testing.T, client *gophercloud.ServiceClient, name string) servers.Server {
	t.Helper()

	imageID := requireEnv(t, "OSPA_E2E_IMAGE_ID")
	flavorID := requireEnv(t, "OSPA_E2E_FLAVOR_ID")
	networkID := requireEnv(t, "OSPA_E2E_NETWORK_ID")

	createOpts := servers.CreateOpts{
		Name:      name,
		ImageRef:  imageID,
		FlavorRef: flavorID,
		Networks: []servers.Network{
			{UUID: networkID},
		},
	}
	// NOTE: We intentionally keep CreateOpts minimal to stay compatible across OpenStack deployments
	// and gophercloud versions. If you want keypair/availability-zone/security-groups support,
	// add the appropriate CreateOpts extensions for your cloud.

	s, err := servers.Create(client, createOpts).Extract()
	if err != nil {
		t.Fatalf("servers.Create: %v", err)
	}

	// Cleanup best-effort if the test fails mid-way.
	t.Cleanup(func() {
		_ = servers.Delete(client, s.ID).ExtractErr()
	})

	waitForStatus(t, client, s.ID, "ACTIVE", 10*time.Minute)
	got, err := servers.Get(client, s.ID).Extract()
	if err != nil {
		t.Fatalf("servers.Get after create: %v", err)
	}
	return *got
}

func stopServer(t *testing.T, client *gophercloud.ServiceClient, id string) {
	t.Helper()
	if err := startstop.Stop(client, id).ExtractErr(); err != nil {
		t.Fatalf("stop server: %v", err)
	}
	waitForStatus(t, client, id, "SHUTOFF", 10*time.Minute)
}

func waitForStatus(t *testing.T, client *gophercloud.ServiceClient, id string, want string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s, err := servers.Get(client, id).Extract()
		if err != nil {
			t.Fatalf("servers.Get while waiting for status=%s: %v", want, err)
		}
		if s.Status == want {
			return
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatalf("timeout waiting for server %s to reach status=%s", id, want)
}

func waitForDeleted(t *testing.T, client *gophercloud.ServiceClient, id string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := servers.Get(client, id).Extract()
		if err != nil {
			// treat 404 as deleted
			if isNotFound(err) {
				return
			}
			// some clouds return 403 after delete, etc; keep it strict for now
		} else {
			time.Sleep(5 * time.Second)
			continue
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("timeout waiting for server %s to be deleted", id)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*gophercloud.ErrDefault404); ok {
		return true
	}
	var ue *gophercloud.ErrUnexpectedResponseCode
	if errors.As(err, &ue) && ue != nil {
		return ue.Actual == 404
	}
	return false
}


