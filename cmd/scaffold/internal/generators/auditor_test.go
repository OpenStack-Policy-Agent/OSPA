package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAuditorFiles_NewFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1", "resource2"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")

	for _, res := range resources {
		filePath := filepath.Join(auditDir, res+".go")
		if !fileExists(filePath) {
			t.Errorf("Auditor file was not created: %q", filePath)
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read auditor file: %v", err)
		}

		contentStr := string(content)

		// Verify package declaration
		if !strings.Contains(contentStr, "package testservice") {
			t.Errorf("Generated file missing package declaration: %q", filePath)
		}

		// Verify imports (placeholder auditor should not import gophercloud/openstack SDK packages)
		requiredImports := []string{"context", "fmt", "audit", "policy"}
		for _, imp := range requiredImports {
			if !strings.Contains(contentStr, imp) {
				t.Errorf("Generated file missing import %q: %q", imp, filePath)
			}
		}
		if strings.Contains(contentStr, "gophercloud/gophercloud/openstack") {
			t.Errorf("Generated placeholder auditor must not import OpenStack SDK packages: %q", filePath)
		}

		// Verify auditor struct
		auditorName := ToPascal(res) + "Auditor"
		if !strings.Contains(contentStr, "type "+auditorName) {
			t.Errorf("Generated file missing auditor struct: %q", filePath)
		}

		// Verify ResourceType() method
		if !strings.Contains(contentStr, "func (a *"+auditorName+") ResourceType()") {
			t.Errorf("Generated file missing ResourceType() method: %q", filePath)
		}

		// Verify Check() method
		if !strings.Contains(contentStr, "func (a *"+auditorName+") Check(") {
			t.Errorf("Generated file missing Check() method: %q", filePath)
		}

		// Verify Fix() method
		if !strings.Contains(contentStr, "func (a *"+auditorName+") Fix(") {
			t.Errorf("Generated file missing Fix() method: %q", filePath)
		}

		// Verify Go syntax
		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Generated file has invalid Go syntax: %v, file: %q", err, filePath)
		}
	}
}

func TestGenerateAuditorFiles_ExistingFile_Skip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create existing auditor file
	existingFile := filepath.Join(auditDir, "resource1.go")
	existingContent := "package testservice\n// existing content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate without force
	resources := []string{"resource1", "resource2"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
	}

	// Verify existing file was not overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "existing content") {
		t.Error("Existing auditor file was overwritten without force flag")
	}

	// Verify new file was created
	newFile := filepath.Join(auditDir, "resource2.go")
	if !fileExists(newFile) {
		t.Error("New auditor file was not created")
	}
}

func TestGenerateAuditorFiles_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create existing file
	existingFile := filepath.Join(auditDir, "resource1.go")
	existingContent := "package testservice\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources, true)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
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

func TestGenerateAuditorFiles_DirectoryCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Directory doesn't exist yet
	resources := []string{"resource1"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if info, err := os.Stat(auditDir); err != nil || !info.IsDir() {
		t.Error("Audit directory was not created")
	}

	filePath := filepath.Join(auditDir, "resource1.go")
	if !fileExists(filePath) {
		t.Error("Auditor file was not created in new directory")
	}
}

func TestGenerateAuditorFiles_MultipleResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1", "resource2", "resource3", "resource4"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	for _, res := range resources {
		filePath := filepath.Join(auditDir, res+".go")
		if !fileExists(filePath) {
			t.Errorf("Auditor file was not created for resource %q", res)
		}
	}
}

func TestGenerateAuditorFiles_TemplateRendering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources, false)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
	}

	filePath := filepath.Join(tmpDir, "pkg", "audit", "testservice", "resource1.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Verify template variables are rendered
	if !strings.Contains(contentStr, "Resource1Auditor") {
		t.Error("Template missing Resource1Auditor")
	}
	if !strings.Contains(contentStr, `return "resource1"`) {
		t.Error("Template missing resource name")
	}
	if !strings.Contains(contentStr, "testservice") {
		t.Error("Template missing service name")
	}
}
