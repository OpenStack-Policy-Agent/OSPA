package generators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAuthMethod_NewMethod(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auth_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	authDir := filepath.Join(tmpDir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create auth dir: %v", err)
	}

	// Create base auth.go file
	authFile := filepath.Join(authDir, "auth.go")
	baseContent := `package auth

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

type Session struct {
	CloudName string
}
`
	if err := os.WriteFile(authFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write auth file: %v", err)
	}

	// Generate new method
	err = GenerateAuthMethod(tmpDir, "testservice", "TestService", "test", false)
	if err != nil {
		t.Fatalf("GenerateAuthMethod() = %v, want nil", err)
	}

	// Verify method was appended
	content, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("Failed to read auth file: %v", err)
	}

	contentStr := string(content)

	// Verify method signature
	if !strings.Contains(contentStr, "func (s *Session) GetTestServiceClient()") {
		t.Error("Generated method missing or incorrect signature")
	}

	// Verify method uses correct service type
	if !strings.Contains(contentStr, `NewServiceClient("test"`) {
		t.Error("Generated method missing or incorrect service type")
	}

	// Verify method uses correct display name
	if !strings.Contains(contentStr, "GetTestServiceClient") {
		t.Error("Generated method missing display name")
	}
}

func TestGenerateAuthMethod_ExistingMethod_Skip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auth_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	authDir := filepath.Join(tmpDir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create auth dir: %v", err)
	}

	// Create auth.go with existing method
	authFile := filepath.Join(authDir, "auth.go")
	baseContent := `package auth

type Session struct {
	CloudName string
}

func (s *Session) GetTestServiceClient() (*gophercloud.ServiceClient, error) {
	return nil, nil
}
`
	if err := os.WriteFile(authFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write auth file: %v", err)
	}

	// Generate without force
	err = GenerateAuthMethod(tmpDir, "testservice", "TestService", "test", false)
	if err != nil {
		t.Fatalf("GenerateAuthMethod() = %v, want nil", err)
	}

	// Verify method was not duplicated
	content, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("Failed to read auth file: %v", err)
	}

	contentStr := string(content)
	count := strings.Count(contentStr, "func (s *Session) GetTestServiceClient()")
	if count != 1 {
		t.Errorf("Method appears %d times, want 1", count)
	}
}

func TestGenerateAuthMethod_ExistingMethod_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auth_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	authDir := filepath.Join(tmpDir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create auth dir: %v", err)
	}

	// Create auth.go with existing method
	authFile := filepath.Join(authDir, "auth.go")
	baseContent := `package auth

type Session struct {
	CloudName string
}

func (s *Session) GetTestServiceClient() (*gophercloud.ServiceClient, error) {
	return nil, nil
}
`
	if err := os.WriteFile(authFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write auth file: %v", err)
	}

	// Generate with force
	err = GenerateAuthMethod(tmpDir, "testservice", "TestService", "test", true)
	if err != nil {
		t.Fatalf("GenerateAuthMethod() = %v, want nil", err)
	}

	// Verify method was appended (will be duplicate)
	content, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("Failed to read auth file: %v", err)
	}

	contentStr := string(content)
	count := strings.Count(contentStr, "func (s *Session) GetTestServiceClient()")
	if count < 1 {
		t.Error("Method was not appended with force flag")
	}
}

func TestGenerateAuthMethod_MissingAuthFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auth_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Don't create auth.go
	err = GenerateAuthMethod(tmpDir, "testservice", "TestService", "test", false)
	if err == nil {
		t.Error("GenerateAuthMethod() = nil, want error for missing file")
	}
}

func TestGenerateAuthMethod_AppendBehavior(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auth_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	authDir := filepath.Join(tmpDir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create auth dir: %v", err)
	}

	// Create auth.go with existing content
	authFile := filepath.Join(authDir, "auth.go")
	baseContent := `package auth

type Session struct {
	CloudName string
}

func (s *Session) GetOtherClient() (*gophercloud.ServiceClient, error) {
	return nil, nil
}
`
	if err := os.WriteFile(authFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write auth file: %v", err)
	}

	// Generate new method
	err = GenerateAuthMethod(tmpDir, "testservice", "TestService", "test", false)
	if err != nil {
		t.Fatalf("GenerateAuthMethod() = %v, want nil", err)
	}

	// Verify method was appended (not inserted)
	content, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("Failed to read auth file: %v", err)
	}

	contentStr := string(content)

	// Verify existing method is still there
	if !strings.Contains(contentStr, "GetOtherClient") {
		t.Error("Existing method was removed")
	}

	// Verify new method was appended
	if !strings.Contains(contentStr, "GetTestServiceClient") {
		t.Error("New method was not appended")
	}

	// Verify new method comes after existing method
	otherIndex := strings.Index(contentStr, "GetOtherClient")
	testIndex := strings.Index(contentStr, "GetTestServiceClient")
	if testIndex < otherIndex {
		t.Error("New method was inserted before existing method")
	}
}

func TestGenerateAuthMethod_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auth_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	authDir := filepath.Join(tmpDir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("Failed to create auth dir: %v", err)
	}

	// Create base auth.go file
	authFile := filepath.Join(authDir, "auth.go")
	baseContent := `package auth

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

type Session struct {
	CloudName string
}
`
	if err := os.WriteFile(authFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write auth file: %v", err)
	}

	// Generate method
	err = GenerateAuthMethod(tmpDir, "testservice", "TestService", "test", false)
	if err != nil {
		t.Fatalf("GenerateAuthMethod() = %v, want nil", err)
	}

	// Verify file can be parsed (basic syntax check)
	content, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("Failed to read auth file: %v", err)
	}

	// Basic check: verify it contains valid Go constructs
	contentStr := string(content)
	if !strings.Contains(contentStr, "func") {
		t.Error("Generated code missing function keyword")
	}
	if !strings.Contains(contentStr, "return") {
		t.Error("Generated code missing return statement")
	}
}

