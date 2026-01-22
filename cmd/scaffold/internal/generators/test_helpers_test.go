package generators

import (
	"os"
	"path/filepath"
)

// setupRepoPrereqs creates minimal files that the generators expect to exist in a real OSPA repo.
// Some generators update existing files (e.g., pkg/auth/auth.go).
func setupRepoPrereqs(baseDir string) error {
	// Minimal auth.go
	authContent := `package auth

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

type Session struct {
	CloudName string
}

// Placeholder client method referenced by generated services
func (s *Session) GetDummyClient() (*gophercloud.ServiceClient, error) {
	client, err := clientconfig.NewServiceClient("dummy", &clientconfig.ClientOpts{Cloud: s.CloudName})
	if err != nil {
		return nil, fmt.Errorf("failed to create dummy client: %w", err)
	}
	return client, nil
}
`
	if _, err := createTempAuthFile(baseDir, authContent); err != nil {
		return err
	}
	return nil
}

// createTempAuthFile creates a temporary auth.go file with given content
func createTempAuthFile(baseDir, content string) (string, error) {
	authDir := filepath.Join(baseDir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(authDir, "auth.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", err
	}

	return filePath, nil
}
