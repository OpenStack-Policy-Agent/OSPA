package auth_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
)

func TestNewSession_InvalidCloudConfigErrors(t *testing.T) {
	dir := t.TempDir()
	clouds := filepath.Join(dir, "clouds.yaml")

	// Intentionally incomplete: no auth section.
	if err := os.WriteFile(clouds, []byte(`
clouds:
  test:
    region_name: RegionOne
    interface: public
`), 0644); err != nil {
		t.Fatalf("write clouds.yaml: %v", err)
	}

	t.Setenv("OS_CLIENT_CONFIG_FILE", clouds)
	t.Setenv("OS_CLOUD", "test")
	// Ensure we don't accidentally use ambient OS_* creds.
	t.Setenv("OS_AUTH_URL", "")
	t.Setenv("OS_USERNAME", "")
	t.Setenv("OS_PASSWORD", "")
	t.Setenv("OS_PROJECT_NAME", "")

	_, err := auth.NewSession("test")
	if err == nil {
		t.Fatalf("NewSession() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to authenticate") {
		t.Fatalf("NewSession() error = %q, want contains %q", err.Error(), "failed to authenticate")
	}
}
