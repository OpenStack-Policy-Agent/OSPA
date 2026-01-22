//go:build integration

package auth_test

import (
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
)

func TestIntegration_NewSession_AndCreateClients(t *testing.T) {
	// Gate on OpenStack config. If not present, skip (CI-safe).
	if os.Getenv("OS_CLIENT_CONFIG_FILE") == "" && os.Getenv("OS_CLOUD") == "" && os.Getenv("OS_AUTH_URL") == "" {
		t.Skip("OpenStack config not found (set OS_CLIENT_CONFIG_FILE/OS_CLOUD or OS_AUTH_URL et al.)")
	}

	s, err := auth.NewSession(os.Getenv("OS_CLOUD"))
	if err != nil {
		t.Fatalf("NewSession() = %v", err)
	}

	if _, err := s.GetComputeClient(); err != nil {
		t.Fatalf("GetComputeClient() = %v", err)
	}
	if _, err := s.GetNetworkClient(); err != nil {
		t.Fatalf("GetNetworkClient() = %v", err)
	}
	if _, err := s.GetBlockStorageClient(); err != nil {
		t.Fatalf("GetBlockStorageClient() = %v", err)
	}
}
