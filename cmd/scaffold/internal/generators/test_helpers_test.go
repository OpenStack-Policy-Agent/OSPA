package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// setupRepoPrereqs creates minimal files that the generators expect to exist in a real OSPA repo.
// Many generators *update* existing files (e.g., pkg/auth/auth.go, pkg/policy/validator.go).
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

	// Minimal pkg/policy/validator.go (used by validation generator to add blank imports)
	validatorDir := filepath.Join(baseDir, "pkg", "policy")
	if err := os.MkdirAll(validatorDir, 0755); err != nil {
		return err
	}
	validatorPath := filepath.Join(validatorDir, "validator.go")
	validatorContent := `package policy

import (
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation"

	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation/compute"
)

// Dummy usage to avoid unused import in this minimal test file.
var _ = validation.List
`
	return os.WriteFile(validatorPath, []byte(validatorContent), 0644)
}

// createTempProjectStructure creates a temporary project structure for testing
func createTempProjectStructure() (string, error) {
	tmpDir, err := os.MkdirTemp("", "test_scaffold_*")
	if err != nil {
		return "", err
	}

	// Create directory structure
	dirs := []string{
		"pkg/services/services",
		"pkg/discovery/services",
		"pkg/audit",
		"pkg/auth",
		"pkg/policy/validation",
		"e2e",
		"examples/policies",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}
	}

	return tmpDir, nil
}

// createTempServiceFile creates a temporary service file with given content
func createTempServiceFile(baseDir, serviceName, content string) (string, error) {
	serviceDir := filepath.Join(baseDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(serviceDir, serviceName+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

// createTempDiscoveryFile creates a temporary discovery file with given content
func createTempDiscoveryFile(baseDir, serviceName, content string) (string, error) {
	discoveryDir := filepath.Join(baseDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(discoveryDir, serviceName+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", err
	}

	return filePath, nil
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

// cleanupTempFiles removes temporary files and directories
func cleanupTempFiles(tmpDir string) error {
	return os.RemoveAll(tmpDir)
}

// verifyServiceFileStructure verifies that a service file has the expected structure
func verifyServiceFileStructure(filePath string, expectedResources []string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	contentStr := string(content)

	// Check for required elements
	required := []string{
		"package services",
		"func init()",
		"MustRegister",
	}

	for _, req := range required {
		if !containsString(contentStr, req) {
			return false
		}
	}

	// Check for RegisterResource calls
	for _, res := range expectedResources {
		if !containsString(contentStr, `RegisterResource("`+res+`")`) {
			return false
		}
	}

	return true
}

// verifyGoCompiles checks if a Go file can be parsed (basic compilation check)
func verifyGoCompiles(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	// Use go/parser to check syntax
	// This is a simplified check - full compilation would require imports
	_, err = parseGoFile(filePath, content)
	return err == nil
}

// Helper function
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Simple Go file parser (uses go/parser internally)
func parseGoFile(filePath string, content []byte) (interface{}, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, filePath, content, parser.ParseComments)
}

