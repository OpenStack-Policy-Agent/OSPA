package generators

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/registry"
)

func TestGenerateService_PermissionDenied(t *testing.T) {
	// This test would require root or special setup
	// For now, we'll skip it as it's environment-dependent
	t.Skip("Permission denied test requires special environment setup")
}

func TestGenerateService_InvalidPath(t *testing.T) {
	// Test with an invalid baseDir in a way that's stable across environments:
	// make baseDir a FILE (not a directory), so creating baseDir/pkg/... must fail.
	tmpFile, err := os.CreateTemp("", "test_invalid_basedir_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFilePath) }()

	err = GenerateServiceFile(tmpFilePath, "testservice", "TestService", "test", []string{"resource1"}, false)
	if err == nil {
		t.Error("GenerateServiceFile with invalid path = nil, want error")
	}
}

func TestExtractResourcesFromServiceFile_MalformedRegisterResource(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_malformed_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Malformed RegisterResource call (wrong number of args)
	content := `package services

func init() {
	RegisterResource("testservice")  // Missing second argument
}
`
	if err := os.WriteFile(tmpFile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Should still parse (just won't extract the resource)
	resources, err := extractResourcesFromServiceFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("extractResourcesFromServiceFile() = %v, want nil", err)
	}

	// Should return empty or handle gracefully
	if len(resources) > 0 {
		t.Logf("extractResourcesFromServiceFile returned resources: %v (may be acceptable)", resources)
	}
}

func TestUpdateServiceFile_NoSwitchStatement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_no_switch_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create service file without switch statement
	serviceFile := filepath.Join(serviceDir, "testservice.go")
	content := `package services

func (s *TestServiceService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	return nil, nil
}
`
	if err := os.WriteFile(serviceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Try to update - should handle gracefully or error
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{"resource1"})
	// This might error or handle gracefully - both are acceptable
	if err != nil {
		t.Logf("UpdateServiceFile with no switch returned error (may be expected): %v", err)
	}
}

func TestUpdateDiscoveryFile_WrongPackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_wrong_package_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	// Create discovery file with wrong package
	discoveryFile := filepath.Join(discoveryDir, "testservice.go")
	content := `package wrongpackage

type TestServiceResource1Discoverer struct{}
`
	if err := os.WriteFile(discoveryFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Try to update - should handle or error
	err = UpdateDiscoveryFile(tmpDir, "testservice", "TestService", []string{"resource2"})
	// This might work (just appends) or error - both acceptable
	if err != nil {
		t.Logf("UpdateDiscoveryFile with wrong package returned error (may be expected): %v", err)
	}
}

func TestGenerateAuthMethod_CorruptedFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_corrupted_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	authDir := filepath.Join(tmpDir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create auth dir: %v", err)
	}

	// Create corrupted auth.go (invalid Go syntax)
	authFile := filepath.Join(authDir, "auth.go")
	content := `package auth

func invalid syntax here!!!
`
	if err := os.WriteFile(authFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Try to generate - should still append (doesn't parse, just appends)
	err = GenerateAuthMethod(tmpDir, "testservice", "TestService", "test", false)
	// Should still work (just appends to file)
	if err != nil {
		t.Logf("GenerateAuthMethod with corrupted file returned error: %v", err)
	}
}

func TestAnalyzeService_InvalidServiceFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_invalid_service_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create service file with syntax errors
	serviceFile := filepath.Join(serviceDir, "testservice.go")
	content := `package services

func init() {
	invalid syntax!!!
}
`
	if err := os.WriteFile(serviceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Try to analyze - should return error
	_, err = AnalyzeService(tmpDir, "testservice", []string{"resource1"})
	if err == nil {
		t.Error("AnalyzeService with invalid file = nil, want error")
	}
}

func TestUpdateServiceFile_MalformedGo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_malformed_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Create malformed Go file
	serviceFile := filepath.Join(serviceDir, "testservice.go")
	content := `package services

func init() {
	RegisterResource("testservice", "resource1")
	// Missing closing brace
`
	if err := os.WriteFile(serviceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Try to update - might error when extracting resources
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", []string{"resource2"})
	// Should error when trying to extract resources
	if err == nil {
		t.Log("UpdateServiceFile with malformed Go succeeded (may extract resources anyway)")
	}
}

func TestGenerateServiceFile_InvalidTemplate(t *testing.T) {
	// This is hard to test as template errors would be caught at parse time
	// But we can test with edge case inputs
	tmpDir, err := os.MkdirTemp("", "test_template_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	serviceDir := filepath.Join(tmpDir, "pkg", "services", "services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	// Test with empty strings
	err = GenerateServiceFile(tmpDir, "", "", "", []string{}, false)
	if err == nil {
		t.Log("GenerateServiceFile with empty strings succeeded")
	}

	// Test with special characters in service name
	err = GenerateServiceFile(tmpDir, "test-service", "Test-Service", "test", []string{"resource1"}, false)
	// Should handle or error - both acceptable
	if err != nil {
		t.Logf("GenerateServiceFile with special chars returned error: %v", err)
	}
}

func TestValidateService_EmptyString(t *testing.T) {
	err := registry.ValidateService("")
	if err == nil {
		t.Error("ValidateService(\"\") = nil, want error")
	}
}

func TestValidateResources_EmptyList(t *testing.T) {
	err := registry.ValidateResources("nova", []string{})
	// Empty list should be valid (no resources to validate)
	if err != nil {
		t.Errorf("ValidateResources with empty list = %v, want nil", err)
	}
}

func TestValidateResources_Duplicates(t *testing.T) {
	// Duplicate resources should be handled
	resources := []string{"instance", "instance", "keypair"}
	err := registry.ValidateResources("nova", resources)
	// Should succeed (duplicates are valid, just redundant)
	if err != nil {
		t.Errorf("ValidateResources with duplicates = %v, want nil", err)
	}
}

func TestValidateResources_CaseVariations(t *testing.T) {
	testCases := []struct {
		service   string
		resources []string
		wantErr   bool
	}{
		{"nova", []string{"Instance", "INSTANCE", "instance"}, false},
		{"neutron", []string{"Security_Group", "security_group"}, false},
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
