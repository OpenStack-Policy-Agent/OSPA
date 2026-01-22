package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateAuthMethod appends an auth client method to auth.go if it does not already exist.
func GenerateAuthMethod(baseDir, serviceName, displayName, serviceType string) error {
	authFile := filepath.Join(baseDir, "pkg", "auth", "auth.go")

	content, err := os.ReadFile(authFile)
	if err != nil {
		return fmt.Errorf("reading auth.go: %w", err)
	}

	methodName := fmt.Sprintf("Get%sClient", displayName)
	if strings.Contains(string(content), methodName) {
		// Method already exists; nothing to do.
		return nil
	}

	methodCode := fmt.Sprintf(`
// Get%sClient returns a client for %s.
func (s *Session) Get%sClient() (*gophercloud.ServiceClient, error) {
	client, err := clientconfig.NewServiceClient("%s", &clientconfig.ClientOpts{
		Cloud: s.CloudName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %%w", err)
	}
	return client, nil
}
`, displayName, displayName, displayName, serviceType, serviceName)

	file, err := os.OpenFile(authFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening auth.go: %w", err)
	}

	if _, err := file.WriteString(methodCode); err != nil {
		_ = file.Close()
		return fmt.Errorf("writing to auth.go: %w", err)
	}
	return file.Close()
}
