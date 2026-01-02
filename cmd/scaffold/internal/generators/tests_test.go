package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateUnitTests_NewFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	resources := []string{"resource1", "resource2"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")

	for _, res := range resources {
		filePath := filepath.Join(auditDir, res+"_test.go")
		if !fileExists(filePath) {
			t.Errorf("Test file was not created: %q", filePath)
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		contentStr := string(content)

		// Verify package declaration
		if !strings.Contains(contentStr, "package testservice") {
			t.Errorf("Generated file missing package declaration: %q", filePath)
		}

		// Verify imports
		requiredImports := []string{"context", "testing", "time", "policy"}
		for _, imp := range requiredImports {
			if !strings.Contains(contentStr, imp) {
				t.Errorf("Generated file missing import %q: %q", imp, filePath)
			}
		}

		// Verify test functions
		testName := strings.Title(res)
		if !strings.Contains(contentStr, "Test"+testName+"Auditor_ResourceType") {
			t.Errorf("Generated file missing TestResourceType function: %q", filePath)
		}
		if !strings.Contains(contentStr, "Test"+testName+"Auditor_Check") {
			t.Errorf("Generated file missing TestCheck function: %q", filePath)
		}
		if !strings.Contains(contentStr, "Test"+testName+"Auditor_Check_AgeGT") {
			t.Errorf("Generated file missing TestCheck_AgeGT function: %q", filePath)
		}
		if !strings.Contains(contentStr, "Test"+testName+"Auditor_Fix") {
			t.Errorf("Generated file missing TestFix function: %q", filePath)
		}

		// Verify Go syntax
		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Generated file has invalid Go syntax: %v, file: %q", err, filePath)
		}
	}
}

func TestGenerateUnitTests_ExistingFile_Skip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create existing test file
	existingFile := filepath.Join(auditDir, "resource1_test.go")
	existingContent := "package testservice\n// existing test content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate without force
	resources := []string{"resource1", "resource2"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	// Verify existing file was not overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "existing test content") {
		t.Error("Existing test file was overwritten without force flag")
	}

	// Verify new file was created
	newFile := filepath.Join(auditDir, "resource2_test.go")
	if !fileExists(newFile) {
		t.Error("New test file was not created")
	}
}

func TestGenerateUnitTests_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create existing file
	existingFile := filepath.Join(auditDir, "resource1_test.go")
	existingContent := "package testservice\n// old test content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources, true)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "old test content") {
		t.Error("File was not overwritten with force flag")
	}
}

func TestGenerateUnitTests_MultipleResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	resources := []string{"resource1", "resource2", "resource3", "resource4"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	for _, res := range resources {
		filePath := filepath.Join(auditDir, res+"_test.go")
		if !fileExists(filePath) {
			t.Errorf("Test file was not created for resource %q", res)
		}
	}
}

func TestGenerateUnitTests_DirectoryCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Directory doesn't exist yet
	resources := []string{"resource1"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if info, err := os.Stat(auditDir); err != nil || !info.IsDir() {
		t.Error("Audit directory was not created")
	}

	filePath := filepath.Join(auditDir, "resource1_test.go")
	if !fileExists(filePath) {
		t.Error("Test file was not created in new directory")
	}
}

func TestGenerateUnitTests_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	resources := []string{"resource1"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	filePath := filepath.Join(tmpDir, "pkg", "audit", "testservice", "resource1_test.go")
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

