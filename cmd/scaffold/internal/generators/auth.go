package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateAuthMethod generates and appends the auth client method to auth.go
func GenerateAuthMethod(baseDir, serviceName, displayName, serviceType string, force bool) error {
	authFile := filepath.Join(baseDir, "pkg", "auth", "auth.go")
	
	// Read existing file
	content, err := os.ReadFile(authFile)
	if err != nil {
		return fmt.Errorf("reading auth.go: %w", err)
	}

	// Check if method already exists
	methodName := fmt.Sprintf("Get%sClient", displayName)
	// Always be idempotent: never append duplicate methods, even with --force.
	// (Overwriting methods safely would require Go AST rewriting; we avoid that here.)
	if strings.Contains(string(content), methodName) {
		fmt.Printf("Warning: Auth method %s already exists in auth.go, skipping\n", methodName)
		return nil
	}

	// Generate method code
	methodCode := fmt.Sprintf(`
// Get%sClient returns a client for %s
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

	// Append to file
	file, err := os.OpenFile(authFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening auth.go: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(methodCode); err != nil {
		return fmt.Errorf("writing to auth.go: %w", err)
	}

	return nil
}

