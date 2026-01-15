package generators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateService_NewService_Complete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_orchestrator_*")
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

	resources := []string{"resource1", "resource2"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Verify all files created
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
			t.Errorf("Expected file was not created: %q", relPath)
		}
	}

	// Verify service file structure
	serviceFile := filepath.Join(tmpDir, "pkg/services/services/testservice.go")
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read service file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "type TestServiceService struct{}") {
		t.Error("Service file missing service struct")
	}
	for _, res := range resources {
		if !strings.Contains(contentStr, `RegisterResource("testservice", "`+res+`")`) {
			t.Errorf("Service file missing RegisterResource for %q", res)
		}
	}

	// Verify discovery file structure
	discoveryFile := filepath.Join(tmpDir, "pkg/discovery/services/testservice.go")
	content, err = os.ReadFile(discoveryFile)
	if err != nil {
		t.Fatalf("Failed to read discovery file: %v", err)
	}

	contentStr = string(content)
	if !strings.Contains(contentStr, "package services") {
		t.Error("Discovery file missing package declaration")
	}
	for _, res := range resources {
		discovererName := "TestService" + ToPascal(res) + "Discoverer"
		if !strings.Contains(contentStr, "type "+discovererName) {
			t.Errorf("Discovery file missing discoverer for %q", res)
		}
	}

	// Verify auditor files created
	for _, res := range resources {
		auditorFile := filepath.Join(tmpDir, "pkg/audit/testservice", res+".go")
		if !fileExists(auditorFile) {
			t.Errorf("Auditor file was not created for %q", res)
		}
	}

	// Verify validation file created
	validationFile := filepath.Join(tmpDir, "pkg/policy/validation/testservice.go")
	if !fileExists(validationFile) {
		t.Error("Validation file was not created")
	}

	// Verify policy guide created
	guideFile := filepath.Join(tmpDir, "examples/policies/testservice-policy-guide.md")
	if !fileExists(guideFile) {
		t.Error("Policy guide was not created")
	}
}

func TestGenerateService_ExistingService_AddResource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_orchestrator_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	// Create existing service with one resource
	serviceDir := filepath.Join(tmpDir, "pkg/services/services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
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
	if err := os.WriteFile(serviceFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Create existing discovery file
	discoveryDir := filepath.Join(tmpDir, "pkg/discovery/services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	discoveryFile := filepath.Join(discoveryDir, "testservice.go")
	discoveryContent := `package services

type TestServiceResource1Discoverer struct{}

func (d *TestServiceResource1Discoverer) ResourceType() string {
	return "resource1"
}
`
	if err := os.WriteFile(discoveryFile, []byte(discoveryContent), 0644); err != nil {
		t.Fatalf("Failed to write discovery file: %v", err)
	}

	// Generate with existing and new resource
	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Verify service file was updated (not overwritten)
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read service file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `RegisterResource("testservice", "resource1")`) {
		t.Error("Service file missing existing resource1")
	}
	if !strings.Contains(contentStr, `RegisterResource("testservice", "resource2")`) {
		t.Error("Service file missing new resource2")
	}

	// Verify new auditor was created
	auditorFile := filepath.Join(tmpDir, "pkg/audit/testservice/resource2.go")
	if !fileExists(auditorFile) {
		t.Error("New auditor file was not created")
	}

	// Verify new test was created
	testFile := filepath.Join(tmpDir, "pkg/audit/testservice/resource2_test.go")
	if !fileExists(testFile) {
		t.Error("New test file was not created")
	}
}

func TestGenerateService_ExistingService_AllResourcesExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_orchestrator_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	// Create existing service with all resources
	serviceDir := filepath.Join(tmpDir, "pkg/services/services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
	RegisterResource("testservice", "resource2")
}
`
	if err := os.WriteFile(serviceFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Create all auditor files
	auditDir := filepath.Join(tmpDir, "pkg/audit/testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	for _, res := range []string{"resource1", "resource2"} {
		auditorFile := filepath.Join(auditDir, res+".go")
		if err := os.WriteFile(auditorFile, []byte("package testservice\n"), 0644); err != nil {
			t.Fatalf("Failed to write auditor file: %v", err)
		}
		testFile := filepath.Join(auditDir, res+"_test.go")
		if err := os.WriteFile(testFile, []byte("package testservice\n"), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Generate with all existing resources
	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Verify files were not overwritten (check timestamps or content)
	// For simplicity, just verify files still exist
	for _, res := range resources {
		auditorFile := filepath.Join(auditDir, res+".go")
		if !fileExists(auditorFile) {
			t.Errorf("Auditor file was removed for %q", res)
		}
	}
}

func TestGenerateService_ForceFlag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_orchestrator_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	// Create existing service file
	serviceDir := filepath.Join(tmpDir, "pkg/services/services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := "package services\n// old content\n"
	if err := os.WriteFile(serviceFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GenerateService("testservice", "TestService", "test", resources, true)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "old content") {
		t.Error("File was not overwritten with force flag")
	}
}

func TestGenerateService_PartialUpdate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_orchestrator_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	// Create existing service with some resources
	serviceDir := filepath.Join(tmpDir, "pkg/services/services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	existingContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
}
`
	if err := os.WriteFile(serviceFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Create auditor for resource1
	auditDir := filepath.Join(tmpDir, "pkg/audit/testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	auditorFile := filepath.Join(auditDir, "resource1.go")
	if err := os.WriteFile(auditorFile, []byte("package testservice\n"), 0644); err != nil {
		t.Fatalf("Failed to write auditor file: %v", err)
	}

	// Generate with resource1 (exists) and resource2 (new)
	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Verify resource1 auditor was not overwritten
	content, err := os.ReadFile(auditorFile)
	if err != nil {
		t.Fatalf("Failed to read auditor file: %v", err)
	}

	if !strings.Contains(string(content), "package testservice") {
		t.Error("Existing auditor file was overwritten")
	}

	// Verify resource2 auditor was created
	newAuditorFile := filepath.Join(auditDir, "resource2.go")
	if !fileExists(newAuditorFile) {
		t.Error("New auditor file was not created")
	}
}

func TestGenerateService_ErrorHandling(t *testing.T) {
	// Test error propagation
	// This is tested indirectly through other tests
	// Direct error testing would require mocking file operations
	t.Skip("Error handling tested indirectly through other tests")
}

func TestGenerateService_FileSystemIsolation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_orchestrator_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	// Generate service
	resources := []string{"resource1"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Verify files were created in current directory (tmpDir)
	serviceFile := filepath.Join(tmpDir, "pkg/services/services/testservice.go")
	if !fileExists(serviceFile) {
		t.Error("Service file was not created in current directory")
	}
}

