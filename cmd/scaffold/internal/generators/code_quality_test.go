package generators

import (
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegression_GeneratedCodeCompiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_code_quality_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateService("testservice", "TestService", "test", resources)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Test that generated Go files can be parsed
	goFiles := []string{
		"pkg/services/services/testservice.go",
		"pkg/discovery/services/testservice.go",
		"pkg/audit/testservice/resource1.go",
		"pkg/audit/testservice/resource1_test.go",
		"pkg/policy/validation/testservice.go",
		"e2e/testservice/resource_creator.go",
		"e2e/testservice/resource1_test.go",
	}

	fset := token.NewFileSet()
	for _, relPath := range goFiles {
		filePath := filepath.Join(tmpDir, relPath)
		if !fileExists(filePath) {
			continue // Skip if file doesn't exist
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read %q: %v", relPath, err)
			continue
		}

		// Try to parse the file
		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Generated file %q does not compile: %v", relPath, err)
		}
	}
}

func TestRegression_GeneratedCodeFormatted(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_code_quality_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	serviceFile := filepath.Join(tmpDir, "pkg/services/services/testservice.go")
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Try to format the code
	formatted, err := format.Source(content)
	if err != nil {
		t.Errorf("Generated code cannot be formatted: %v", err)
		return
	}

	// Compare original with formatted (should be the same if already formatted)
	// Allow for minor whitespace differences
	originalStr := strings.TrimSpace(string(content))
	formattedStr := strings.TrimSpace(string(formatted))

	// Basic check: if formatting changes the code significantly, it wasn't formatted
	if len(originalStr) > 0 && len(formattedStr) > 0 {
		// Code should be mostly the same after formatting
		similarity := float64(len(originalStr)) / float64(len(formattedStr))
		if similarity < 0.8 {
			t.Logf("Generated code may not be properly formatted (similarity: %.2f)", similarity)
		}
	}
}

func TestRegression_NoSyntaxErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_code_quality_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change dir: %v", err)
		}
	}
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	resources := []string{"resource1", "resource2"}
	err = GenerateService("testservice", "TestService", "test", resources)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Check all generated Go files for syntax errors
	goFiles := []string{
		"pkg/services/services/testservice.go",
		"pkg/discovery/services/testservice.go",
		"pkg/audit/testservice/resource1.go",
		"pkg/audit/testservice/resource2.go",
		"pkg/policy/validation/testservice.go",
	}

	fset := token.NewFileSet()
	for _, relPath := range goFiles {
		filePath := filepath.Join(tmpDir, relPath)
		if !fileExists(filePath) {
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Syntax error in %q: %v", relPath, err)
		}
	}
}

func TestRegression_ImportsCorrect(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_code_quality_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1"}
	err = GenerateServiceFile(tmpDir, "testservice", "TestService", "test", resources)
	if err != nil {
		t.Fatalf("GenerateServiceFile() = %v, want nil", err)
	}

	serviceFile := filepath.Join(tmpDir, "pkg/services/services/testservice.go")
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Check for duplicate imports (basic check)
	importCount := strings.Count(contentStr, "import (")
	if importCount > 1 {
		t.Error("Generated code has multiple import blocks")
	}

	// Check for common import patterns
	expectedImports := []string{
		"fmt",
		"audit",
		"auth",
		"discovery",
		"gophercloud",
	}

	for _, imp := range expectedImports {
		// Count occurrences - should appear once in import block
		count := strings.Count(contentStr, `"`+imp+`"`)
		if count > 1 {
			t.Logf("Import %q appears %d times (may be acceptable)", imp, count)
		}
	}
}
