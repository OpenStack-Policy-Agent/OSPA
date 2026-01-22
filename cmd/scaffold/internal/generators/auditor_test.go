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
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources)
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
			t.Errorf("Failed to read auditor file: %v", err)
			continue
		}

		contentStr := string(content)

		if !strings.Contains(contentStr, "package testservice") {
			t.Errorf("Generated file missing package declaration: %q", filePath)
		}

		auditorName := ToPascal(res) + "Auditor"
		if !strings.Contains(contentStr, "type "+auditorName) {
			t.Errorf("Generated file missing auditor struct: %q", filePath)
		}
		if !strings.Contains(contentStr, "func (a *"+auditorName+") ResourceType()") {
			t.Errorf("Generated file missing ResourceType() method: %q", filePath)
		}
		if !strings.Contains(contentStr, "func (a *"+auditorName+") Check(") {
			t.Errorf("Generated file missing Check() method: %q", filePath)
		}
		if !strings.Contains(contentStr, "func (a *"+auditorName+") Fix(") {
			t.Errorf("Generated file missing Fix() method: %q", filePath)
		}

		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("Generated file has invalid Go syntax: %v, file: %q", err, filePath)
		}
	}
}

func TestGenerateAuditorFiles_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	existingFile := filepath.Join(auditDir, "resource1.go")
	existingContent := "package testservice\n// old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	resources := []string{"resource1"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
	}

	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "old content") {
		t.Error("File was not overwritten")
	}
}

func TestGenerateAuditorFiles_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_auditor_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1", "resource2"}
	err = GenerateAuditorFiles(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateAuditorFiles() = %v, want nil", err)
	}

	auditDir := filepath.Join(tmpDir, "pkg", "audit", "testservice")
	for _, res := range resources {
		filePath := filepath.Join(auditDir, res+".go")
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
