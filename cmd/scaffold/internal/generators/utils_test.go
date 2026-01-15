package generators

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"
)

func TestFileExists_ExistingFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	
	if !fileExists(tmpFile.Name()) {
		t.Errorf("fileExists(%q) = false, want true", tmpFile.Name())
	}
}

func TestFileExists_NonExistentFile(t *testing.T) {
	nonExistent := "/tmp/nonexistent_file_12345.txt"
	if fileExists(nonExistent) {
		t.Errorf("fileExists(%q) = true, want false", nonExistent)
	}
}

func TestFileExists_Directory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	
	// fileExists should return false for directories
	if fileExists(tmpDir) {
		t.Errorf("fileExists(%q) = true for directory, want false", tmpDir)
	}
}

func TestWriteFile_NewFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_write_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	
	filePath := filepath.Join(tmpDir, "test.go")
	
	tmpl, err := template.New("test").Parse("package test\n\nconst Value = {{.Value}}\n")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	
	data := struct {
		Value string
	}{
		Value: "\"test\"",
	}
	
	if err := writeFile(filePath, tmpl, data); err != nil {
		t.Fatalf("writeFile() = %v, want nil", err)
	}
	
	if !fileExists(filePath) {
		t.Error("writeFile() did not create file")
	}
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	
	expected := "package test\n\nconst Value = \"test\"\n"
	if string(content) != expected {
		t.Errorf("writeFile() content = %q, want %q", string(content), expected)
	}
}

func TestWriteFile_ExistingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_write_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	
	filePath := filepath.Join(tmpDir, "test.go")
	
	// Create initial file
	initialContent := "package test\n\nconst Old = \"old\"\n"
	if err := os.WriteFile(filePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}
	
	// Overwrite with template
	tmpl, err := template.New("test").Parse("package test\n\nconst New = {{.Value}}\n")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	
	data := struct {
		Value string
	}{
		Value: "\"new\"",
	}
	
	if err := writeFile(filePath, tmpl, data); err != nil {
		t.Fatalf("writeFile() = %v, want nil", err)
	}
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	
	expected := "package test\n\nconst New = \"new\"\n"
	if string(content) != expected {
		t.Errorf("writeFile() content = %q, want %q", string(content), expected)
	}
}

func TestWriteFile_DirectoryCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_write_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	
	// File path with nested directories that don't exist
	filePath := filepath.Join(tmpDir, "nested", "deep", "test.go")
	
	tmpl, err := template.New("test").Parse("package test\n")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	
	if err := writeFile(filePath, tmpl, nil); err != nil {
		t.Fatalf("writeFile() = %v, want nil", err)
	}
	
	if !fileExists(filePath) {
		t.Error("writeFile() did not create file in nested directory")
	}
}

func TestWriteFile_TemplateExecution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_write_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	
	filePath := filepath.Join(tmpDir, "test.go")
	
	tmpl, err := template.New("test").Parse(`
package {{.Package}}

type {{.Type}} struct {
	Name string
	Value int
}
`)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	
	data := struct {
		Package string
		Type    string
	}{
		Package: "test",
		Type:    "TestStruct",
	}
	
	if err := writeFile(filePath, tmpl, data); err != nil {
		t.Fatalf("writeFile() = %v, want nil", err)
	}
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	
	expected := `
package test

type TestStruct struct {
	Name string
	Value int
}
`
	if string(content) != expected {
		t.Errorf("writeFile() template execution = %q, want %q", string(content), expected)
	}
}

func TestWriteFile_Permissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_write_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	
	filePath := filepath.Join(tmpDir, "test.go")
	
	tmpl, err := template.New("test").Parse("package test\n")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	
	if err := writeFile(filePath, tmpl, nil); err != nil {
		t.Fatalf("writeFile() = %v, want nil", err)
	}
	
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	
	// Check that file is readable and writable by owner
	mode := info.Mode()
	if mode&0400 == 0 {
		t.Error("File is not readable by owner")
	}
	if mode&0200 == 0 {
		t.Error("File is not writable by owner")
	}
}

