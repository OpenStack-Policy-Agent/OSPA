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
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources, false)
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

	// Verify package declaration
	if !strings.Contains(contentStr, "package services") {
		t.Error("Generated file missing package declaration")
	}

	// Verify imports (placeholder should not pull in pagination/openstack SDK packages)
	if !strings.Contains(contentStr, "context") {
		t.Error("Generated file missing context import")
	}
	if !strings.Contains(contentStr, "gophercloud") {
		t.Error("Generated file missing gophercloud import")
	}
	// The placeholder may mention pagination in TODOs, but must not import the pagination package.
	if strings.Contains(contentStr, "gophercloud/gophercloud/pagination") {
		t.Error("Generated placeholder file must not import pagination package")
	}

	// Verify discoverer structs for each resource
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

	// Verify placeholder behavior (should return an empty channel)
	if !strings.Contains(contentStr, "close(ch)") {
		t.Error("Generated placeholder discovery should close(ch) and return it")
	}

	// Verify Go syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

func TestGenerateDiscoveryFile_ExistingFile_Update(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	// Existing implementation
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Try to generate with existing resource and new resource
	resources := []string{"resource1", "resource2"}
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
	}

	// Verify file was updated
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(content)

	// Verify both discoverers are present
	if !strings.Contains(contentStr, "TestServiceResource1Discoverer") {
		t.Error("Updated file missing existing Resource1Discoverer")
	}
	if !strings.Contains(contentStr, "TestServiceResource2Discoverer") {
		t.Error("Updated file missing new Resource2Discoverer")
	}
}

func TestGenerateDiscoveryFile_ExistingFile_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	// Create existing file
	existingFile := filepath.Join(discoveryDir, "testservice.go")
	existingContent := "package services\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources, true)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
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

func TestGenerateDiscoveryFile_PlaceholderDoesNotImportOpenstackSDK(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
	}

	filePath := filepath.Join(discoveryDir, "testservice.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "gophercloud/gophercloud/openstack") || strings.Contains(contentStr, "openstack/testservice/") {
		t.Fatalf("Generated placeholder discovery must not import OpenStack SDK packages; got:\n%s", contentStr)
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
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
	}

	filePath := filepath.Join(discoveryDir, "testservice.go")
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

