package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateServiceFile_NewService(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	filePath := filepath.Join(serviceDir, "testservice.go")
	if !fileExists(filePath) {
		t.Fatal("Service file was not created")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read service file: %v", err)
	}

	contentStr := string(content)

	// Verify package declaration
	if !strings.Contains(contentStr, "package services") {
		t.Error("Generated file missing package declaration")
	}

	// Verify service struct
	if !strings.Contains(contentStr, "type TestServiceService struct{}") {
		t.Error("Generated file missing service struct")
	}

	// Verify init() with RegisterResource calls
	for _, res := range resources {
		if !strings.Contains(contentStr, `RegisterResource("testservice", "`+res+`")`) {
			t.Errorf("Generated file missing RegisterResource call for %q", res)
		}
	}

	// Verify GetClient method
	if !strings.Contains(contentStr, "GetTestServiceClient()") {
		t.Error("Generated file missing GetClient method")
	}

	// Verify GetResourceAuditor switch
	if !strings.Contains(contentStr, "GetResourceAuditor") {
		t.Error("Generated file missing GetResourceAuditor method")
	}
	for _, res := range resources {
		if !strings.Contains(contentStr, `case "`+res+`":`) {
			t.Errorf("Generated file missing case for resource %q in GetResourceAuditor", res)
		}
	}

	// Verify GetResourceDiscoverer switch
	if !strings.Contains(contentStr, "GetResourceDiscoverer") {
		t.Error("Generated file missing GetResourceDiscoverer method")
	}
	for _, res := range resources {
		if !strings.Contains(contentStr, `case "`+res+`":`) {
			t.Errorf("Generated file missing case for resource %q in GetResourceDiscoverer", res)
		}
	}

	// Verify Go syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

func TestGenerateServiceFile_ExistingService_Update(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create existing service file with one resource
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

	// Try to generate with existing resource and new resource
	resources := []string{"resource1", "resource2"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	// Verify file was updated (not overwritten)
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(content)

	// Verify both resources are present
	if !strings.Contains(contentStr, `RegisterResource("testservice", "resource1")`) {
		t.Error("Updated file missing existing resource1")
	}
	if !strings.Contains(contentStr, `RegisterResource("testservice", "resource2")`) {
		t.Error("Updated file missing new resource2")
	}

	// Verify both cases in switches
	if !strings.Contains(contentStr, `case "resource1":`) {
		t.Error("Updated file missing case for resource1")
	}
	if !strings.Contains(contentStr, `case "resource2":`) {
		t.Error("Updated file missing case for resource2")
	}
}

func TestGenerateServiceFile_ExistingService_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create existing file
	existingFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := "package services\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, true)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "old content") {
		t.Error("File was not overwritten with force flag")
	}
}

func TestGenerateServiceFile_SingleResource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	filePath := filepath.Join(serviceDir, "testservice.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `RegisterResource("testservice", "resource1")`) {
		t.Error("Single resource not registered")
	}
}

func TestGenerateServiceFile_MultipleResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	resources := []string{"resource1", "resource2", "resource3", "resource4"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	filePath := filepath.Join(serviceDir, "testservice.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	for _, res := range resources {
		if !strings.Contains(contentStr, `RegisterResource("testservice", "`+res+`")`) {
			t.Errorf("Resource %q not registered", res)
		}
	}
}

func TestGenerateServiceFile_TemplateRendering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Test with various display names and service types
	testCases := []struct {
		serviceName string
		displayName string
		serviceType string
		resources   []string
	}{
		{"glance", "Glance", "image", []string{"image"}},
		{"keystone", "Keystone", "identity", []string{"user", "role"}},
		{"test-service", "TestService", "test-type", []string{"resource1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.serviceName, func(t *testing.T) {
			err := GenerateServiceFile(tmpDir, tc.serviceName, tc.displayName, tc.serviceType, tc.resources, false)
			if err != nil {
				t.Fatalf("GenerateServiceFile() = %v, want nil", err)
			}

			filePath := filepath.Join(serviceDir, tc.serviceName+".go")
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, tc.displayName+"Service") {
				t.Errorf("Template missing display name %q", tc.displayName)
			}
			if !strings.Contains(contentStr, `return "`+tc.serviceName+`"`) {
				t.Errorf("Template missing service name %q", tc.serviceName)
			}

			// Clean up for next iteration
			os.Remove(filePath)
		})
	}
}

func TestGenerateServiceFile_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_service_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	filePath := filepath.Join(serviceDir, "testservice.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify Go syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

