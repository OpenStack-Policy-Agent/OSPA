package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateServiceFile_AddSingleResource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create existing service file
	existingFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := `package services

import (
	"fmt"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

type TestServiceService struct{}

func init() {
	MustRegister(&TestServiceService{})
	RegisterResource("testservice", "resource1")
}

func (s *TestServiceService) Name() string {
	return "testservice"
}

func (s *TestServiceService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.GetTestServiceClient()
}

func (s *TestServiceService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "resource1":
		return &testservice.Resource1Auditor{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q", resourceType)
	}
}

func (s *TestServiceService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "resource1":
		return &discovery.TestServiceResource1Discoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q", resourceType)
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with new resource
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{"resource2"})
	if err != nil {
		t.Fatalf("UpdateServiceFile() = %v, want nil", err)
	}

	// Verify all updates
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(content)

	// Verify RegisterResource call was added
	if !strings.Contains(contentStr, `RegisterResource("testservice", "resource2")`) {
		t.Error("RegisterResource call was not added")
	}

	// Verify case was added to GetResourceAuditor
	if !strings.Contains(contentStr, `case "resource2":`) {
		t.Error("Case was not added to GetResourceAuditor")
	}

	// Verify case was added to GetResourceDiscoverer
	discovererCases := strings.Count(contentStr, `case "resource2":`)
	if discovererCases < 1 {
		t.Error("Case was not added to GetResourceDiscoverer")
	}

	// Verify Go syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, existingFile, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Updated file has invalid Go syntax: %v", err)
	}
}

func TestUpdateServiceFile_AddMultipleResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create existing service file
	existingFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := `package services

func init() {
	MustRegister(&TestServiceService{})
	RegisterResource("testservice", "resource1")
}

func (s *TestServiceService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "resource1":
		return nil, nil
	default:
		return nil, nil
	}
}

func (s *TestServiceService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "resource1":
		return nil, nil
	default:
		return nil, nil
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with multiple new resources
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{"resource2", "resource3"})
	if err != nil {
		t.Fatalf("UpdateServiceFile() = %v, want nil", err)
	}

	// Verify all resources were added
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	for _, res := range []string{"resource2", "resource3"} {
		if !strings.Contains(contentStr, `RegisterResource("testservice", "`+res+`")`) {
			t.Errorf("RegisterResource call was not added for %q", res)
		}
		if !strings.Contains(contentStr, `case "`+res+`":`) {
			t.Errorf("Case was not added for %q", res)
		}
	}
}

func TestUpdateServiceFile_ResourceAlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create existing service file
	existingFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
}

func (s *TestServiceService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "resource1":
		return nil, nil
	default:
		return nil, nil
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Get original content
	originalContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	// Try to update with existing resource
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{"resource1"})
	if err != nil {
		t.Fatalf("UpdateServiceFile() = %v, want nil", err)
	}

	// Verify file was not modified
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(originalContent) {
		t.Error("File was modified when resource already exists")
	}
}

func TestUpdateServiceFile_NoResourcesToAdd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create existing service file
	existingFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Get original content
	originalContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	// Try to update with empty list
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{})
	if err != nil {
		t.Fatalf("UpdateServiceFile() = %v, want nil", err)
	}

	// Verify file was not modified
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(originalContent) {
		t.Error("File was modified when no resources to add")
	}
}

func TestUpdateServiceFile_MissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// File doesn't exist
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{"resource1"})
	if err == nil {
		t.Error("UpdateServiceFile() = nil, want error for missing file")
	}
}

func TestUpdateServiceFile_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create existing service file
	existingFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
}

func (s *TestServiceService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	case "resource1":
		return nil, nil
	default:
		return nil, nil
	}
}

func (s *TestServiceService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	case "resource1":
		return nil, nil
	default:
		return nil, nil
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with new resource
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{"resource2"})
	if err != nil {
		t.Fatalf("UpdateServiceFile() = %v, want nil", err)
	}

	// Verify Go syntax
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, existingFile, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Updated file has invalid Go syntax: %v", err)
	}
}
