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

func TestGenerateDiscoveryFile_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_discovery_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	discoveryDir := filepath.Join(tmpDir, "pkg", "discovery", "services")
	if err := os.MkdirAll(discoveryDir, 0755); err != nil {
		t.Fatalf("Failed to create discovery dir: %v", err)
	}

	existingFile := filepath.Join(discoveryDir, "testservice.go")
	existingContent := "package services\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateDiscoveryFile(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateDiscoveryFile() = %v, want nil", err)
	}

	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "old content") {
		t.Error("File was not overwritten")
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
