package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateE2ETest_NewFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateE2ETest(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	filePath := filepath.Join(e2eDir, "testservice_test.go")
	if !fileExists(filePath) {
		t.Fatal("E2E test file was not created")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read e2e test file: %v", err)
	}

	contentStr := string(content)

	// Verify build tag
	if !strings.Contains(contentStr, "//go:build e2e") {
		t.Error("Generated file missing build tag")
	}

	// Verify package declaration
	if !strings.Contains(contentStr, "package e2e") {
		t.Error("Generated file missing package declaration")
	}

	// Verify imports
	if !strings.Contains(contentStr, "testing") {
		t.Error("Generated file missing testing import")
	}

	// Verify test functions for each resource
	for _, res := range resources {
		testName := "TestTestService_" + ToPascal(res) + "Audit"
		if !strings.Contains(contentStr, "func "+testName) {
			t.Errorf("Generated file missing test function: %q", testName)
		}
	}

	// Verify test function structure
	if !strings.Contains(contentStr, "NewTestEngine") {
		t.Error("Generated file missing NewTestEngine call")
	}
	if !strings.Contains(contentStr, "LoadPolicyFromYAML") {
		t.Error("Generated file missing LoadPolicyFromYAML call")
	}
	if !strings.Contains(contentStr, "RunAudit") {
		t.Error("Generated file missing RunAudit call")
	}

	// Verify Go syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated file has invalid Go syntax: %v", err)
	}
}

func TestGenerateE2ETest_ExistingFile_Append(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	// Create existing e2e test file
	existingFile := filepath.Join(e2eDir, "testservice_test.go")
	existingContent := `//go:build e2e

package e2e

import (
	"testing"
)

func TestTestService_Resource1Audit(t *testing.T) {
	// Existing test
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Try to generate with existing resource and new resource
	resources := []string{"resource1", "resource2"}
	err = GenerateE2ETest(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	// Verify file was updated
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(content)

	// Verify existing test is still there
	if !strings.Contains(contentStr, "TestTestService_Resource1Audit") {
		t.Error("Updated file missing existing test function")
	}

	// Verify new test was appended
	if !strings.Contains(contentStr, "TestTestService_Resource2Audit") {
		t.Error("Updated file missing new test function")
	}
}

func TestGenerateE2ETest_ExistingFile_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	// Create existing file
	existingFile := filepath.Join(e2eDir, "testservice_test.go")
	existingContent := "//go:build e2e\npackage e2e\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GenerateE2ETest(tmpDir, "testservice", "TestService", resources, true)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
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

func TestGenerateE2ETest_ExistingTest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	// Create existing e2e test file with test for resource1
	existingFile := filepath.Join(e2eDir, "testservice_test.go")
	existingContent := `//go:build e2e

package e2e

import (
	"testing"
)

func TestTestService_Resource1Audit(t *testing.T) {
	// Existing test
}
`
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Try to generate with only resource1 (already exists)
	resources := []string{"resource1"}
	err = GenerateE2ETest(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	// Verify file was not modified (no duplicate)
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	count := strings.Count(contentStr, "TestTestService_Resource1Audit")
	if count != 1 {
		t.Errorf("Test function appears %d times, want 1", count)
	}
}

func TestGenerateE2ETest_MultipleResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	resources := []string{"resource1", "resource2", "resource3"}
	err = GenerateE2ETest(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	filePath := filepath.Join(e2eDir, "testservice_test.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	for _, res := range resources {
		testName := "TestTestService_" + ToPascal(res) + "Audit"
		if !strings.Contains(contentStr, testName) {
			t.Errorf("Test function missing for resource %q", res)
		}
	}
}

func TestGenerateE2ETest_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateE2ETest(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	filePath := filepath.Join(e2eDir, "testservice_test.go")
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

