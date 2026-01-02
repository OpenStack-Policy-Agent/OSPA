package generators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePolicyGuide_NewFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	resources := []string{"resource1", "resource2"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	guideFile := filepath.Join(tmpDir, "examples", "policies", "testservice-policy-guide.md")
	if !fileExists(guideFile) {
		t.Fatal("Policy guide file was not created")
	}

	content, err := os.ReadFile(guideFile)
	if err != nil {
		t.Fatalf("Failed to read policy guide file: %v", err)
	}

	contentStr := string(content)

	// Verify markdown structure
	if !strings.Contains(contentStr, "# Policy Guide:") {
		t.Error("Generated guide missing title")
	}

	// Verify service overview section
	if !strings.Contains(contentStr, "Service Overview") {
		t.Error("Generated guide missing service overview section")
	}
	if !strings.Contains(contentStr, "testservice") {
		t.Error("Generated guide missing service name")
	}
	if !strings.Contains(contentStr, "TestService") {
		t.Error("Generated guide missing display name")
	}
	if !strings.Contains(contentStr, "test") {
		t.Error("Generated guide missing service type")
	}

	// Verify supported resources section
	if !strings.Contains(contentStr, "Supported Resources") {
		t.Error("Generated guide missing supported resources section")
	}
	for _, res := range resources {
		if !strings.Contains(contentStr, res) {
			t.Errorf("Generated guide missing resource %q", res)
		}
	}

	// Verify policy structure section
	if !strings.Contains(contentStr, "Policy Structure") {
		t.Error("Generated guide missing policy structure section")
	}

	// Verify examples for each resource
	for _, res := range resources {
		if !strings.Contains(contentStr, res) {
			t.Errorf("Generated guide missing example for resource %q", res)
		}
	}
}

func TestGeneratePolicyGuide_ExistingFile_Skip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	examplesDir := filepath.Join(tmpDir, "examples", "policies")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		t.Fatalf("Failed to create examples dir: %v", err)
	}

	// Create existing guide file
	existingFile := filepath.Join(examplesDir, "testservice-policy-guide.md")
	existingContent := "# Policy Guide\n\nExisting content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate without force
	resources := []string{"resource1"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	// Verify file was not overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "Existing content") {
		t.Error("Existing policy guide was overwritten without force flag")
	}
}

func TestGeneratePolicyGuide_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	examplesDir := filepath.Join(tmpDir, "examples", "policies")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		t.Fatalf("Failed to create examples dir: %v", err)
	}

	// Create existing guide file
	existingFile := filepath.Join(examplesDir, "testservice-policy-guide.md")
	existingContent := "# Policy Guide\n\nOld content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Generate with force
	resources := []string{"resource1"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources, true)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "Old content") {
		t.Error("File was not overwritten with force flag")
	}
}

func TestGeneratePolicyGuide_DirectoryCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Directory doesn't exist yet
	resources := []string{"resource1"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	examplesDir := filepath.Join(tmpDir, "examples", "policies")
	if info, err := os.Stat(examplesDir); err != nil || !info.IsDir() {
		t.Error("Examples directory was not created")
	}

	guideFile := filepath.Join(examplesDir, "testservice-policy-guide.md")
	if !fileExists(guideFile) {
		t.Error("Policy guide file was not created in new directory")
	}
}

func TestGeneratePolicyGuide_MultipleResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	resources := []string{"resource1", "resource2", "resource3", "resource4"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	guideFile := filepath.Join(tmpDir, "examples", "policies", "testservice-policy-guide.md")
	content, err := os.ReadFile(guideFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	for _, res := range resources {
		if !strings.Contains(contentStr, res) {
			t.Errorf("Policy guide missing resource %q", res)
		}
	}
}

func TestGeneratePolicyGuide_MarkdownValidity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	resources := []string{"resource1"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	guideFile := filepath.Join(tmpDir, "examples", "policies", "testservice-policy-guide.md")
	content, err := os.ReadFile(guideFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Basic markdown validity checks
	if !strings.Contains(contentStr, "#") {
		t.Error("Generated guide missing markdown headers")
	}
	if !strings.Contains(contentStr, "```") {
		t.Error("Generated guide missing code blocks")
	}
}

