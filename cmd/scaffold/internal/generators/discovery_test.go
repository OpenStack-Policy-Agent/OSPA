package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDiscoveryFile_NewFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
	}

	filePath := filepath.Join(discoveryDir, "testservice.go")
	if !fileExists(filePath) {
		t.Fatal("Discovery file was not created")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read discovery file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "package services") {
		t.Error("Generated file missing package declaration")
	}
	if !strings.Contains(contentStr, "context") {
		t.Error("Generated file missing context import")
	}
	if !strings.Contains(contentStr, "gophercloud") {
		t.Error("Generated file missing gophercloud import")
	}

	for _, res := range resources {
		discovererName := "TestService" + ToPascal(res) + "Discoverer"
		if !strings.Contains(contentStr, "type "+discovererName) {
			t.Errorf("Generated file missing discoverer struct: %q", discovererName)
		}
		if !strings.Contains(contentStr, "func (d *"+discovererName+") ResourceType()") {
			t.Errorf("Generated file missing ResourceType() for %q", discovererName)
		}
		if !strings.Contains(contentStr, "func (d *"+discovererName+") Discover(") {
			t.Errorf("Generated file missing Discover() for %q", discovererName)
		}
	}

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

func TestGenerateDiscoveryFile_SkipsExistingDiscoverers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	// Create an existing file with a discoverer already defined
	existingFile := filepath.Join(discoveryDir, "testservice.go")
	existingContent := `package services

import (
	"context"
	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

// TestServiceResource1Discoverer already exists - should not be regenerated
type TestServiceResource1Discoverer struct{}

func (d *TestServiceResource1Discoverer) ResourceType() string {
	return "resource1"
}

func (d *TestServiceResource1Discoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)
	go func() {
		defer close(ch)
		// Real implementation here
	}()
	return ch, nil
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Try to generate for the same resource - should be skipped
	resources := []string{"resource1"}
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
	}

	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Existing discoverer should be preserved
	if !strings.Contains(string(content), "Real implementation here") {
		t.Error("Existing discoverer was overwritten when it should have been skipped")
	}

	// Should not have duplicate type declarations
	count := strings.Count(string(content), "type TestServiceResource1Discoverer struct{}")
	if count != 1 {
		t.Errorf("Expected 1 discoverer declaration, got %d", count)
	}
}

func TestGenerateDiscoveryFile_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
	}

	filePath := filepath.Join(discoveryDir, "testservice.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}
