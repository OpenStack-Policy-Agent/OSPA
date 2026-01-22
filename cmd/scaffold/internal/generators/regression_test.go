package generators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/registry"
)

func TestRegression_OriginalScaffoldBehavior(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_regression_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test that original scaffold behavior still works
	// Generate a new service (should work as before)
	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil (original behavior broken)", err)
	}

	// Verify all expected files are created
	expectedFiles := []string{
		"pkg/services/services/testservice.go",
		"pkg/discovery/services/testservice.go",
		"pkg/audit/testservice/resource1.go",
		"pkg/audit/testservice/resource2.go",
		"pkg/audit/testservice/resource1_test.go",
		"pkg/audit/testservice/resource2_test.go",
		"pkg/policy/validation/testservice.go",
		"e2e/testservice_test.go",
		"examples/policies/testservice-policy-guide.md",
	}

	for _, relPath := range expectedFiles {
		filePath := filepath.Join(tmpDir, relPath)
		if !fileExists(filePath) {
			t.Errorf("Original behavior broken: file not created: %q", relPath)
		}
	}
}

func TestRegression_AllFileTypesGenerated(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_regression_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Verify all file types are still generated
	fileTypes := map[string]string{
		"service":      "pkg/services/services/testservice.go",
		"discovery":    "pkg/discovery/services/testservice.go",
		"auditor":      "pkg/audit/testservice/resource1.go",
		"test":         "pkg/audit/testservice/resource1_test.go",
		"validation":   "pkg/policy/validation/testservice.go",
		"e2e":          "e2e/testservice_test.go",
		"policy_guide": "examples/policies/testservice-policy-guide.md",
	}

	for fileType, relPath := range fileTypes {
		filePath := filepath.Join(tmpDir, relPath)
		if !fileExists(filePath) {
			t.Errorf("File type %q not generated: %q", fileType, relPath)
		}
	}
}

func TestRegression_ValidationStillWorks(t *testing.T) {
	// Test that validation logic still works
	testCases := []struct {
		service   string
		resources []string
		wantErr   bool
	}{
		{"nova", []string{"instance", "keypair"}, false},
		{"invalid", []string{"resource1"}, true},
		{"nova", []string{"invalid_resource"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			err := registry.ValidateResources(tc.service, tc.resources)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateResources(%q, %v) = %v, want error = %v", tc.service, tc.resources, err, tc.wantErr)
			}
		})
	}
}

func TestRegression_ForceFlagBehavior(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_regression_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create existing file
	serviceDir := filepath.Join(tmpDir, "pkg/services/services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	originalContent := "package services\n// original content\n"
	if err := os.WriteFile(serviceFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Generate with force - should overwrite
	resources := []string{"resource1"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, true)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "original content") {
		t.Error("Force flag behavior broken: file was not overwritten")
	}
}

func TestRegression_GeneratedCodeStructure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_regression_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	serviceFile := filepath.Join(tmpDir, "pkg/services/services/testservice.go")
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Verify expected structure elements are present
	requiredElements := []string{
		"package services",
		"type TestServiceService struct{}",
		"func init()",
		"MustRegister",
		"RegisterResource",
		"func (s *TestServiceService) Name()",
		"func (s *TestServiceService) GetClient(",
		"func (s *TestServiceService) GetResourceAuditor(",
		"func (s *TestServiceService) GetResourceDiscoverer(",
	}

	for _, element := range requiredElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("Generated code structure broken: missing %q", element)
		}
	}
}
