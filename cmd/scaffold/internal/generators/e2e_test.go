package generators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateE2ETest_CreatesServiceDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"resource1", "resource2"}
	err = GenerateE2ETest(tmpDir, "testservice", "TestService", resources)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	// Verify service directory was created
	serviceDir := filepath.Join(tmpDir, "e2e", "testservice")
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		t.Fatal("Service directory was not created")
	}
}

func TestGenerateE2ETest_CreatesResourceCreator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"network", "port"}
	err = GenerateE2ETest(tmpDir, "neutron", "Neutron", resources)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	// Verify resource_creator.go was created
	creatorPath := filepath.Join(tmpDir, "e2e", "neutron", "resource_creator.go")
	if !fileExists(creatorPath) {
		t.Fatal("resource_creator.go was not created")
	}

	content, err := os.ReadFile(creatorPath)
	if err != nil {
		t.Fatalf("Failed to read resource_creator.go: %v", err)
	}

	contentStr := string(content)

	// Verify it has the right package
	if !strings.Contains(contentStr, "package neutron") {
		t.Error("resource_creator.go missing package declaration")
	}

	// Verify it has build tag
	if !strings.Contains(contentStr, "//go:build e2e") {
		t.Error("resource_creator.go missing build tag")
	}

	// Verify it has Create functions for each resource
	if !strings.Contains(contentStr, "CreateNetwork") {
		t.Error("resource_creator.go missing CreateNetwork function")
	}
	if !strings.Contains(contentStr, "CreatePort") {
		t.Error("resource_creator.go missing CreatePort function")
	}

	// Verify it has instructions
	if !strings.Contains(contentStr, "RESOURCE CREATOR") {
		t.Error("resource_creator.go missing header instructions")
	}
	if !strings.Contains(contentStr, "DEPENDENCY GRAPH") {
		t.Error("resource_creator.go missing dependency graph")
	}
}

func TestGenerateE2ETest_CreatesResourceTestFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"network", "port"}
	err = GenerateE2ETest(tmpDir, "neutron", "Neutron", resources)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	// Verify individual test files were created
	for _, res := range resources {
		testPath := filepath.Join(tmpDir, "e2e", "neutron", res+"_test.go")
		if !fileExists(testPath) {
			t.Errorf("%s_test.go was not created", res)
			continue
		}

		content, err := os.ReadFile(testPath)
		if err != nil {
			t.Errorf("Failed to read %s_test.go: %v", res, err)
			continue
		}

		contentStr := string(content)

		// Verify package
		if !strings.Contains(contentStr, "package neutron") {
			t.Errorf("%s_test.go missing package declaration", res)
		}

		// Verify test functions
		pascalRes := ToPascal(res)
		if !strings.Contains(contentStr, "TestNeutron_"+pascalRes+"_StatusCheck") {
			t.Errorf("%s_test.go missing StatusCheck test", res)
		}
		if !strings.Contains(contentStr, "TestNeutron_"+pascalRes+"_UnusedCheck") {
			t.Errorf("%s_test.go missing UnusedCheck test", res)
		}
		if !strings.Contains(contentStr, "TestNeutron_"+pascalRes+"_ExemptNames") {
			t.Errorf("%s_test.go missing ExemptNames test", res)
		}

		// Verify it references the resource creator
		if !strings.Contains(contentStr, "Create"+pascalRes) {
			t.Errorf("%s_test.go missing Create%s call", res, pascalRes)
		}

		// Verify instructions
		if !strings.Contains(contentStr, "BEFORE WRITING TESTS") {
			t.Errorf("%s_test.go missing instructions", res)
		}
		if !strings.Contains(contentStr, "TEST COVERAGE CHECKLIST") {
			t.Errorf("%s_test.go missing coverage checklist", res)
		}
	}
}

func TestGenerateE2ETest_GoSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"network", "port"}
	err = GenerateE2ETest(tmpDir, "neutron", "Neutron", resources)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	serviceDir := filepath.Join(tmpDir, "e2e", "neutron")

	// Check all generated files have valid Go syntax
	files := []string{"resource_creator.go", "network_test.go", "port_test.go"}
	for _, file := range files {
		filePath := filepath.Join(serviceDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
			continue
		}

		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
		if err != nil {
			t.Errorf("%s has invalid Go syntax: %v", file, err)
		}
	}
}

func TestGenerateE2ETest_ClientMethodMapping(t *testing.T) {
	testCases := []struct {
		service    string
		wantClient string
	}{
		{"nova", "GetComputeClient"},
		{"neutron", "GetNetworkClient"},
		{"cinder", "GetBlockStorageClient"},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			resources := []string{"resource1"}
			err = GenerateE2ETest(tmpDir, tc.service, ToPascal(tc.service), resources)
			if err != nil {
				t.Fatalf("GenerateE2ETest(%s) = %v, want nil", tc.service, err)
			}

			testPath := filepath.Join(tmpDir, "e2e", tc.service, "resource1_test.go")
			content, err := os.ReadFile(testPath)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			if !strings.Contains(string(content), tc.wantClient) {
				t.Errorf("Expected %s in generated file for service %s", tc.wantClient, tc.service)
			}
		})
	}
}

func TestGenerateE2ETest_MultipleResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_gen_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	resources := []string{"network", "subnet", "port", "security_group"}
	err = GenerateE2ETest(tmpDir, "neutron", "Neutron", resources)
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	serviceDir := filepath.Join(tmpDir, "e2e", "neutron")

	// Verify all resource test files exist
	for _, res := range resources {
		testPath := filepath.Join(serviceDir, res+"_test.go")
		if !fileExists(testPath) {
			t.Errorf("%s_test.go was not created", res)
		}
	}

	// Verify resource_creator has all resources
	creatorPath := filepath.Join(serviceDir, "resource_creator.go")
	content, err := os.ReadFile(creatorPath)
	if err != nil {
		t.Fatalf("Failed to read resource_creator.go: %v", err)
	}

	contentStr := string(content)
	for _, res := range resources {
		funcName := "Create" + ToPascal(res)
		if !strings.Contains(contentStr, funcName) {
			t.Errorf("resource_creator.go missing %s function", funcName)
		}
	}
}

func TestGenerateE2ETest_AddsEngineClientMethod(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_engine_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create e2e directory and a minimal engine.go
	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	engineContent := `package e2e

import (
	"testing"
	"github.com/gophercloud/gophercloud/v2"
)

type TestEngine struct{}

// GetNetworkClient returns a gophercloud client for the networking (Neutron) service.
func (e *TestEngine) GetNetworkClient(t *testing.T) *gophercloud.ServiceClient {
	return nil
}

// LoadPolicy loads a policy file.
func (e *TestEngine) LoadPolicy(t *testing.T, path string) {}
`
	if err := os.WriteFile(filepath.Join(e2eDir, "engine.go"), []byte(engineContent), 0644); err != nil {
		t.Fatalf("Failed to write engine.go: %v", err)
	}

	// Generate for glance service
	err = GenerateE2ETest(tmpDir, "glance", "Glance", []string{"image"})
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	// Read engine.go and verify the new method was added
	content, err := os.ReadFile(filepath.Join(e2eDir, "engine.go"))
	if err != nil {
		t.Fatalf("Failed to read engine.go: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "GetImageClient") {
		t.Error("engine.go missing GetImageClient method")
	}
	if !strings.Contains(contentStr, "GetGlanceClient") {
		t.Error("engine.go missing GetGlanceClient auth call")
	}
}

func TestGenerateE2ETest_SkipsExistingClientMethod(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_e2e_skip_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create e2e directory with engine.go that already has the method
	e2eDir := filepath.Join(tmpDir, "e2e")
	if err := os.MkdirAll(e2eDir, 0755); err != nil {
		t.Fatalf("Failed to create e2e dir: %v", err)
	}

	originalContent := `package e2e

import (
	"testing"
	"github.com/gophercloud/gophercloud/v2"
)

type TestEngine struct{}

// GetNetworkClient returns a gophercloud client for the networking (Neutron) service.
func (e *TestEngine) GetNetworkClient(t *testing.T) *gophercloud.ServiceClient {
	return nil
}
`
	if err := os.WriteFile(filepath.Join(e2eDir, "engine.go"), []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write engine.go: %v", err)
	}

	// Generate for neutron (which maps to GetNetworkClient, already exists)
	err = GenerateE2ETest(tmpDir, "neutron", "Neutron", []string{"network"})
	if err != nil {
		t.Fatalf("GenerateE2ETest() = %v, want nil", err)
	}

	// Read engine.go and verify no duplicate was added
	content, err := os.ReadFile(filepath.Join(e2eDir, "engine.go"))
	if err != nil {
		t.Fatalf("Failed to read engine.go: %v", err)
	}

	// Count occurrences of GetNetworkClient
	contentStr := string(content)
	count := strings.Count(contentStr, "func (e *TestEngine) GetNetworkClient")
	if count != 1 {
		t.Errorf("GetNetworkClient appears %d times, want 1", count)
	}
}
