package generators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeService_NewService(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_analyze_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	analysis, err := AnalyzeService(tmpDir, "nonexistent", []string{"resource1", "resource2"})
	if err != nil {
		t.Fatalf("AnalyzeService() = %v, want nil", err)
	}

	if analysis.ServiceExists {
		t.Error("ServiceExists = true, want false")
	}
	if analysis.DiscoveryExists {
		t.Error("DiscoveryExists = true, want false")
	}
	if analysis.ValidationExists {
		t.Error("ValidationExists = true, want false")
	}
	if analysis.E2ETestExists {
		t.Error("E2ETestExists = true, want false")
	}
	if len(analysis.ExistingResources) != 0 {
		t.Errorf("ExistingResources = %v, want empty", analysis.ExistingResources)
	}
	if len(analysis.MissingResources) != 2 {
		t.Errorf("MissingResources = %v, want [resource1 resource2]", analysis.MissingResources)
	}
}

func TestAnalyzeService_ExistingService(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_analyze_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create service file structure
	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	serviceContent := `package services

import (
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/gophercloud/gophercloud"
)

type TestServiceService struct{}

func init() {
	MustRegister(&TestServiceService{})
	RegisterResource("testservice", "resource1")
	RegisterResource("testservice", "resource2")
}

func (s *TestServiceService) Name() string {
	return "testservice"
}

func (s *TestServiceService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return nil, nil
}
`
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service file: %v", err)
	}

	analysis, err := AnalyzeService(tmpDir, "testservice", []string{"resource1", "resource2"})
	if err != nil {
		t.Fatalf("AnalyzeService() = %v, want nil", err)
	}

	if !analysis.ServiceExists {
		t.Error("ServiceExists = false, want true")
	}
	if len(analysis.ExistingResources) != 2 {
		t.Errorf("ExistingResources = %v, want [resource1 resource2]", analysis.ExistingResources)
	}
	if len(analysis.MissingResources) != 0 {
		t.Errorf("MissingResources = %v, want empty", analysis.MissingResources)
	}
}

func TestAnalyzeService_PartialResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_analyze_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	serviceContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
}
`
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service file: %v", err)
	}

	analysis, err := AnalyzeService(tmpDir, "testservice", []string{"resource1", "resource2", "resource3"})
	if err != nil {
		t.Fatalf("AnalyzeService() = %v, want nil", err)
	}

	if len(analysis.ExistingResources) != 1 {
		t.Errorf("ExistingResources = %v, want [resource1]", analysis.ExistingResources)
	}
	if len(analysis.MissingResources) != 2 {
		t.Errorf("MissingResources = %v, want [resource2 resource3]", analysis.MissingResources)
	}
	if !contains(analysis.MissingResources, "resource2") || !contains(analysis.MissingResources, "resource3") {
		t.Errorf("MissingResources = %v, should contain resource2 and resource3", analysis.MissingResources)
	}
}

func TestAnalyzeService_AllResourcesExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_analyze_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	serviceContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
	RegisterResource("testservice", "resource2")
}
`
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service file: %v", err)
	}

	analysis, err := AnalyzeService(tmpDir, "testservice", []string{"resource1", "resource2"})
	if err != nil {
		t.Fatalf("AnalyzeService() = %v, want nil", err)
	}

	if len(analysis.MissingResources) != 0 {
		t.Errorf("MissingResources = %v, want empty", analysis.MissingResources)
	}
}

func TestExtractResourcesFromServiceFile_ValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_service_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	content := `package services

func init() {
	MustRegister(&TestService{})
	RegisterResource("testservice", "resource1")
	RegisterResource("testservice", "resource2")
	RegisterResource("testservice", "resource3")
}
`
	if err := os.WriteFile(tmpFile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	resources, err := extractResourcesFromServiceFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("extractResourcesFromServiceFile() = %v, want nil", err)
	}

	if len(resources) != 3 {
		t.Errorf("extractResourcesFromServiceFile() returned %d resources, want 3", len(resources))
	}

	expected := []string{"resource1", "resource2", "resource3"}
	for _, exp := range expected {
		if !contains(resources, exp) {
			t.Errorf("extractResourcesFromServiceFile() missing resource: %q", exp)
		}
	}
}

func TestExtractResourcesFromServiceFile_NoResources(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_service_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	content := `package services

func init() {
	MustRegister(&TestService{})
}
`
	if err := os.WriteFile(tmpFile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	resources, err := extractResourcesFromServiceFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("extractResourcesFromServiceFile() = %v, want nil", err)
	}

	if len(resources) != 0 {
		t.Errorf("extractResourcesFromServiceFile() returned %v, want empty", resources)
	}
}

func TestExtractResourcesFromServiceFile_InvalidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_service_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	content := `package services

func init() {
	Invalid syntax here!!!
}
`
	if err := os.WriteFile(tmpFile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = extractResourcesFromServiceFile(tmpFile.Name())
	if err == nil {
		t.Error("extractResourcesFromServiceFile() = nil, want error")
	}
}

func TestExtractResourcesFromServiceFile_MissingFile(t *testing.T) {
	_, err := extractResourcesFromServiceFile("/nonexistent/file.go")
	if err == nil {
		t.Error("extractResourcesFromServiceFile() = nil, want error")
	}
}

func TestCheckAuthMethodExists_Exists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_auth_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	content := `package auth

func (s *Session) GetTestClient() (*gophercloud.ServiceClient, error) {
	return nil, nil
}
`
	if err := os.WriteFile(tmpFile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if !checkAuthMethodExists(tmpFile.Name(), "GetTestClient") {
		t.Error("checkAuthMethodExists() = false, want true")
	}
}

func TestCheckAuthMethodExists_NotExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_auth_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	content := `package auth

func (s *Session) GetOtherClient() (*gophercloud.ServiceClient, error) {
	return nil, nil
}
`
	if err := os.WriteFile(tmpFile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if checkAuthMethodExists(tmpFile.Name(), "GetTestClient") {
		t.Error("checkAuthMethodExists() = true, want false")
	}
}

func TestCheckAuthMethodExists_MissingFile(t *testing.T) {
	if checkAuthMethodExists("/nonexistent/auth.go", "GetTestClient") {
		t.Error("checkAuthMethodExists() = true for missing file, want false")
	}
}

func TestGetDisplayName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"glance", "Glance"},
		{"nova", "Nova"},
		{"keystone", "Keystone"},
		{"neutron", "Neutron"},
		{"cinder", "Cinder"},
		{"a", "A"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := getDisplayName(tc.input)
			if got != tc.expected {
				t.Errorf("getDisplayName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestGetDisplayName_Empty(t *testing.T) {
	got := getDisplayName("")
	if got != "" {
		t.Errorf("getDisplayName(\"\") = %q, want \"\"", got)
	}
}

func TestAnalyzeService_MissingFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_analyze_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create service file but not other files
	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	serviceContent := `package services

func init() {
	RegisterResource("testservice", "resource1")
}
`
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service file: %v", err)
	}

	analysis, err := AnalyzeService(tmpDir, "testservice", []string{"resource1"})
	if err != nil {
		t.Fatalf("AnalyzeService() = %v, want nil", err)
	}

	// Should have missing files for auditor and test
	if len(analysis.MissingFiles) == 0 {
		t.Error("MissingFiles is empty, should contain auditor and test files")
	}

	// Check that missing files include auditor and test files
	hasAuditorFile := false
	hasTestFile := false
	for _, file := range analysis.MissingFiles {
		if strings.Contains(file, "audit") && strings.Contains(file, "resource1.go") {
			hasAuditorFile = true
		}
		if strings.Contains(file, "resource1_test.go") {
			hasTestFile = true
		}
	}

	if !hasAuditorFile {
		t.Error("MissingFiles should contain auditor file")
	}
	if !hasTestFile {
		t.Error("MissingFiles should contain test file")
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
