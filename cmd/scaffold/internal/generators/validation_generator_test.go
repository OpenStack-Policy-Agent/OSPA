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
	err = GenerateValidationFile(tmpDir, "testservice", "TestService", resources, false)
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

	// Verify package declaration
	if !strings.Contains(contentStr, "package validation") {
		t.Error("Generated file missing package declaration")
	}

	// Verify imports
	if !strings.Contains(contentStr, "fmt") {
		t.Error("Generated file missing fmt import")
	}
	if !strings.Contains(contentStr, "policy") {
		t.Error("Generated file missing policy import")
	}

	// Verify validator struct
	if !strings.Contains(contentStr, "type TestServiceValidator struct{}") {
		t.Error("Generated file missing validator struct")
	}

	// Verify init() with registration call
	if !strings.Contains(contentStr, "func init()") {
		t.Error("Generated file missing init function")
	}
	if !strings.Contains(contentStr, "policy.RegisterValidator(&TestServiceValidator{})") {
		t.Error("Generated file missing policy.RegisterValidator call")
	}

	// Verify ServiceName() method
	if !strings.Contains(contentStr, "func (v *TestServiceValidator) ServiceName()") {
		t.Error("Generated file missing ServiceName method")
	}
	if !strings.Contains(contentStr, `return "testservice"`) {
		t.Error("Generated file missing service name return")
	}

	// Verify ValidateResource() switch with all resources
	if !strings.Contains(contentStr, "func (v *TestServiceValidator) ValidateResource(") {
		t.Error("Generated file missing ValidateResource method")
	}
	for _, res := range resources {
		if !strings.Contains(contentStr, `case "`+res+`":`) {
			t.Errorf("Generated file missing case for resource %q", res)
		}
	}

	// Verify Go syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

func TestGenerateValidationFile_ExistingFile_Update(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validation_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	validationDir := filepath.Join(tmpDir, "pkg", "policy", "validation")
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		t.Fatalf("Failed to create validation dir: %v", err)
	}

	// Create existing validation file
	existingFile := filepath.Join(validationDir, "testservice.go")
	existingContent := `package validation

import (
	"fmt"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

type TestServiceValidator struct{}

func init() {
	Register(&TestServiceValidator{})
}

func (v *TestServiceValidator) ServiceName() string {
	return "testservice"
}

func (v *TestServiceValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
	case "resource1":
		// TODO: Add validation
		return nil
	default:
		return fmt.Errorf("unsupported resource type %q", resourceType)
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Try to generate with existing resource and new resource
	resources := []string{"resource1", "resource2"}
	err = GenerateValidationFile(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateValidationFile() = %v, want nil", err)
	}

	// Verify file was updated
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(content)

	// Verify both resources are present
	if !strings.Contains(contentStr, `case "resource1":`) {
		t.Error("Updated file missing existing resource1")
	}
	if !strings.Contains(contentStr, `case "resource2":`) {
		t.Error("Updated file missing new resource2")
	}
}

func TestGenerateValidationFile_ExistingFile_Force(t *testing.T) {
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

	// Create existing file
	existingFile := filepath.Join(validationDir, "testservice.go")
	existingContent := "package validation\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GenerateValidationFile(tmpDir, "testservice", "TestService", resources, true)
	if err != nil {
		t.Fatalf("GenerateValidationFile() = %v, want nil", err)
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

func TestUpdateValidationFile_AddResource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validation_update_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	validationDir := filepath.Join(tmpDir, "pkg", "policy", "validation")
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		t.Fatalf("Failed to create validation dir: %v", err)
	}

	// Create existing validation file
	existingFile := filepath.Join(validationDir, "testservice.go")
	existingContent := `package validation

func (v *TestServiceValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
	case "resource1":
		return nil
	default:
		return fmt.Errorf("unsupported")
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with new resource
	err = updateValidationFile(existingFile, "testservice", "TestService", []string{"resource2"})
	if err != nil {
		t.Fatalf("updateValidationFile() = %v, want nil", err)
	}

	// Verify new resource was added
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `case "resource2":`) {
		t.Error("New resource case was not added")
	}
}

func TestUpdateValidationFile_ResourceExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validation_update_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	validationDir := filepath.Join(tmpDir, "pkg", "policy", "validation")
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		t.Fatalf("Failed to create validation dir: %v", err)
	}

	// Create existing validation file
	existingFile := filepath.Join(validationDir, "testservice.go")
	existingContent := `package validation

func (v *TestServiceValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
	case "resource1":
		return nil
	default:
		return fmt.Errorf("unsupported")
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Try to update with existing resource
	err = updateValidationFile(existingFile, "testservice", "TestService", []string{"resource1"})
	if err != nil {
		t.Fatalf("updateValidationFile() = %v, want nil", err)
	}

	// Verify no duplicate was created
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	count := strings.Count(contentStr, `case "resource1":`)
	if count != 1 {
		t.Errorf("Resource case appears %d times, want 1", count)
	}
}

func TestUpdateValidationFile_InsertBeforeDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validation_update_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	validationDir := filepath.Join(tmpDir, "pkg", "policy", "validation")
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		t.Fatalf("Failed to create validation dir: %v", err)
	}

	// Create existing validation file with default case
	existingFile := filepath.Join(validationDir, "testservice.go")
	existingContent := `package validation

func (v *TestServiceValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
	case "resource1":
		return nil
	default:
		return fmt.Errorf("unsupported")
	}
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Update with new resource
	err = updateValidationFile(existingFile, "testservice", "TestService", []string{"resource2"})
	if err != nil {
		t.Fatalf("updateValidationFile() = %v, want nil", err)
	}

	// Verify new resource was inserted before default
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	resource2Index := strings.Index(contentStr, `case "resource2":`)
	defaultIndex := strings.Index(contentStr, "\n\tdefault:")
	if resource2Index == -1 || defaultIndex == -1 {
		t.Error("Could not find resource2 case or default case")
	}
	if resource2Index > defaultIndex {
		t.Error("New resource case was not inserted before default case")
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
	err = GenerateValidationFile(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateValidationFile() = %v, want nil", err)
	}

	filePath := filepath.Join(validationDir, "testservice.go")
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

