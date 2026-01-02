package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateDiscoveryFile_AddSingleDiscoverer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	// Create existing discovery file
	existingFile := filepath.Join(discoveryDir, "testservice.go")
	existingContent := `package services

import (
	"context"
	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

type TestServiceResource1Discoverer struct{}

func (d *TestServiceResource1Discoverer) ResourceType() string {
	return "resource1"
}

func (d *TestServiceResource1Discoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with new discoverer
	err = UpdateDiscoveryFile(tmpDir, "testservice", "TestService", []string{"resource2"})
	if err != nil {
		t.Fatalf("UpdateDiscoveryFile() = %v, want nil", err)
	}

	// Verify new discoverer was added
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(content)

	// Verify new discoverer struct
	if !strings.Contains(contentStr, "type TestServiceResource2Discoverer struct{}") {
		t.Error("New discoverer struct was not added")
	}

	// Verify ResourceType method
	if !strings.Contains(contentStr, `return "resource2"`) {
		t.Error("ResourceType method was not added for new discoverer")
	}

	// Verify Discover method
	if !strings.Contains(contentStr, "func (d *TestServiceResource2Discoverer) Discover(") {
		t.Error("Discover method was not added for new discoverer")
	}

	// Verify Go syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, existingFile, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Updated file has invalid Go syntax: %v", err)
	}
}

func TestUpdateDiscoveryFile_AddMultipleDiscoverers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	// Create existing discovery file
	existingFile := filepath.Join(discoveryDir, "testservice.go")
	existingContent := `package services

type TestServiceResource1Discoverer struct{}

func (d *TestServiceResource1Discoverer) ResourceType() string {
	return "resource1"
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with multiple new discoverers
	err = UpdateDiscoveryFile(tmpDir, "testservice", "TestService", []string{"resource2", "resource3"})
	if err != nil {
		t.Fatalf("UpdateDiscoveryFile() = %v, want nil", err)
	}

	// Verify all discoverers were added
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	for _, res := range []string{"resource2", "resource3"} {
		discovererName := "TestService" + ToPascal(res) + "Discoverer"
		if !strings.Contains(contentStr, "type "+discovererName) {
			t.Errorf("Discoverer struct was not added for %q", res)
		}
	}
}

func TestUpdateDiscoveryFile_DiscovererAlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	// Create existing discovery file
	existingFile := filepath.Join(discoveryDir, "testservice.go")
	existingContent := `package services

type TestServiceResource1Discoverer struct{}

func (d *TestServiceResource1Discoverer) ResourceType() string {
	return "resource1"
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

	// Try to update with existing discoverer
	err = UpdateDiscoveryFile(tmpDir, "testservice", "TestService", []string{"resource1"})
	if err != nil {
		t.Fatalf("UpdateDiscoveryFile() = %v, want nil", err)
	}

	// Verify file was not modified
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(originalContent) {
		t.Error("File was modified when discoverer already exists")
	}
}

func TestUpdateDiscoveryFile_MissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// File doesn't exist
	err = UpdateDiscoveryFile(tmpDir, "testservice", "TestService", []string{"resource1"})
	if err == nil {
		t.Error("UpdateDiscoveryFile() = nil, want error for missing file")
	}
}

func TestGenerateDiscovererCode_ValidCode(t *testing.T) {
	code := generateDiscovererCode("testservice", "TestService", []string{"resource1"})
	
	if code == "" {
		t.Error("generateDiscovererCode() returned empty string")
	}

	// Verify basic structure
	if !strings.Contains(code, "type TestServiceResource1Discoverer") {
		t.Error("Generated code missing discoverer struct")
	}
	if !strings.Contains(code, "ResourceType()") {
		t.Error("Generated code missing ResourceType method")
	}
	if !strings.Contains(code, "Discover(") {
		t.Error("Generated code missing Discover method")
	}
}

func TestGenerateDiscovererCode_PackageReferences(t *testing.T) {
	code := generateDiscovererCode("testservice", "TestService", []string{"resource1"})
	
	// Placeholder should not reference discovery helpers; it returns an empty channel.
	if !strings.Contains(code, "close(ch)") {
		t.Error("Generated code missing close(ch) placeholder")
	}
}

func TestUpdateDiscoveryFile_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_updater_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	// Create existing discovery file
	existingFile := filepath.Join(discoveryDir, "testservice.go")
	existingContent := `package services

import (
	"context"
	"github.com/gophercloud/gophercloud"
)

type TestServiceResource1Discoverer struct{}

func (d *TestServiceResource1Discoverer) ResourceType() string {
	return "resource1"
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with new discoverer
	err = UpdateDiscoveryFile(tmpDir, "testservice", "TestService", []string{"resource2"})
	if err != nil {
		t.Fatalf("UpdateDiscoveryFile() = %v, want nil", err)
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

