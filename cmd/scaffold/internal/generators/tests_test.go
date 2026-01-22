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
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1", "resource2"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources)
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
			t.Errorf("Failed to read test file: %v", err)
			continue
		}

		contentStr := string(content)

		if !strings.Contains(contentStr, "package testservice") {
			t.Errorf("Generated file missing package declaration: %q", filePath)
		}

		testName := "Test" + ToPascal(res) + "Auditor_ResourceType"
		if !strings.Contains(contentStr, testName) {
			t.Errorf("Generated file missing ResourceType test: %q", filePath)
		}

		checkTestName := "Test" + ToPascal(res) + "Auditor_Check"
		if !strings.Contains(contentStr, checkTestName) {
			t.Errorf("Generated file missing Check test: %q", filePath)
		}

		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Generated file has invalid Go syntax: %v, file: %q", err, filePath)
		}
	}
}

func TestGenerateUnitTests_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	existingFile := filepath.Join(auditDir, "resource1_test.go")
	existingContent := "package testservice\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "old content") {
		t.Error("File was not overwritten")
	}
}

func TestGenerateUnitTests_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_tests_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1", "resource2"}
	err = GenerateUnitTests(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateUnitTests() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	for _, res := range resources {
		filePath := filepath.Join(auditDir, res+"_test.go")
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
			continue
		}

		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Generated file has invalid Go syntax: %v", err)
		}
	}
}
