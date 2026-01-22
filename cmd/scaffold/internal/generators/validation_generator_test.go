package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateValidationFile_NewFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validation_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	validationDir := filepath.Join(tmpDir, "pkg", "policy", "validation")
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		t.Fatalf("Failed to create validation dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateValidationFile(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateValidationFile() = %v, want nil", err)
	}

	filePath := filepath.Join(validationDir, "testservice.go")
	if !fileExists(filePath) {
		t.Fatal("Validation file was not created")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read validation file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "package validation") {
		t.Error("Generated file missing package declaration")
	}
	if !strings.Contains(contentStr, "type TestServiceValidator struct{}") {
		t.Error("Generated file missing validator struct")
	}
	if !strings.Contains(contentStr, "func init()") {
		t.Error("Generated file missing init function")
	}
	if !strings.Contains(contentStr, "policy.RegisterValidator(&TestServiceValidator{})") {
		t.Error("Generated file missing policy.RegisterValidator call")
	}
	if !strings.Contains(contentStr, "func (v *TestServiceValidator) ServiceName()") {
		t.Error("Generated file missing ServiceName method")
	}

	for _, res := range resources {
		if !strings.Contains(contentStr, `case "`+res+`":`) {
			t.Errorf("Generated file missing case for resource %q", res)
		}
	}

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

func TestGenerateValidationFile_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validation_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	validationDir := filepath.Join(tmpDir, "pkg", "policy", "validation")
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		t.Fatalf("Failed to create validation dir: %v", err)
	}

	existingFile := filepath.Join(validationDir, "testservice.go")
	existingContent := "package validation\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateValidationFile(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateValidationFile() = %v, want nil", err)
	}

	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "old content") {
		t.Error("File was not overwritten")
	}
}

func TestGenerateValidationFile_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validation_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	validationDir := filepath.Join(tmpDir, "pkg", "policy", "validation")
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		t.Fatalf("Failed to create validation dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateValidationFile(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateValidationFile() = %v, want nil", err)
	}

	filePath := filepath.Join(validationDir, "testservice.go")
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
