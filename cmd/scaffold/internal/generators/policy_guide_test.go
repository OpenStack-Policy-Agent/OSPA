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
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1", "resource2"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	guideFile := filepath.Join(tmpDir, "examples", "policies", "testservice-policy-guide.md")
	if !fileExists(guideFile) {
		t.Fatal("Policy guide was not created")
	}

	content, err := os.ReadFile(guideFile)
	if err != nil {
		t.Fatalf("Failed to read policy guide: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "# Policy Guide: TestService") {
		t.Error("Policy guide missing title")
	}
	if !strings.Contains(contentStr, "testservice") {
		t.Error("Policy guide missing service name")
	}

	for _, res := range resources {
		if !strings.Contains(contentStr, res) {
			t.Errorf("Policy guide missing resource %q", res)
		}
	}
}

func TestGeneratePolicyGuide_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	examplesDir := filepath.Join(tmpDir, "examples", "policies")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		t.Fatalf("Failed to create examples dir: %v", err)
	}

	existingFile := filepath.Join(examplesDir, "testservice-policy-guide.md")
	existingContent := "# Old content\n"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	resources := []string{"resource1"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if strings.Contains(string(content), "Old content") {
		t.Error("File was not overwritten")
	}
}

func TestGeneratePolicyGuide_ContainsExamples(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_policy_guide_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1"}
	err = GeneratePolicyGuide(tmpDir, "testservice", "TestService", "test", resources)
	if err != nil {
		t.Fatalf("GeneratePolicyGuide() = %v, want nil", err)
	}

	guideFile := filepath.Join(tmpDir, "examples", "policies", "testservice-policy-guide.md")
	content, err := os.ReadFile(guideFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "action: log") {
		t.Error("Policy guide missing log action example")
	}
	if !strings.Contains(contentStr, "action: delete") {
		t.Error("Policy guide missing delete action example")
	}
}
