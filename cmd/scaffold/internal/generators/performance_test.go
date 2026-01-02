package generators

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPerformance_GenerateLargeService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "test_performance_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate service with many resources (10+)
	largeResourceList := []string{
		"resource1", "resource2", "resource3", "resource4", "resource5",
		"resource6", "resource7", "resource8", "resource9", "resource10",
		"resource11", "resource12", "resource13", "resource14", "resource15",
	}

	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer os.Chdir(oldDir)
		os.Chdir(tmpDir)
	}
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	start := time.Now()
	err = GenerateService("testservice", "TestService", "test", largeResourceList, false)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Should complete in reasonable time (< 5 seconds)
	if duration > 5*time.Second {
		t.Errorf("GenerateService with %d resources took %v, want < 5s", len(largeResourceList), duration)
	}

	t.Logf("Generated service with %d resources in %v", len(largeResourceList), duration)
}

func TestPerformance_AnalyzeLargeService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "test_performance_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create service file with many resources
	serviceDir := filepath.Join(tmpDir, "pkg/services/services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	content := "package services\n\nfunc init() {\n"
	for i := 1; i <= 20; i++ {
		content += `	RegisterResource("testservice", "resource` + string(rune('0'+i%10)) + `")` + "\n"
	}
	content += "}\n"

	if err := os.WriteFile(serviceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Analyze with many resources
	largeResourceList := make([]string, 20)
	for i := range largeResourceList {
		largeResourceList[i] = "resource" + string(rune('0'+i%10))
	}

	start := time.Now()
	_, err = AnalyzeService(tmpDir, "testservice", largeResourceList)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("AnalyzeService() = %v, want nil", err)
	}

	// Should complete quickly (< 1 second)
	if duration > 1*time.Second {
		t.Errorf("AnalyzeService with %d resources took %v, want < 1s", len(largeResourceList), duration)
	}

	t.Logf("Analyzed service with %d resources in %v", len(largeResourceList), duration)
}

func TestPerformance_UpdateLargeService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "test_performance_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create service file with many existing resources
	serviceDir := filepath.Join(tmpDir, "pkg/services/services")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	serviceFile := filepath.Join(serviceDir, "testservice.go")
	content := `package services

func init() {
`
	for i := 1; i <= 15; i++ {
		content += `	RegisterResource("testservice", "resource` + string(rune('0'+i%10)) + `")` + "\n"
	}
	content += `}

func (s *TestServiceService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
`
	for i := 1; i <= 15; i++ {
		content += `	case "resource` + string(rune('0'+i%10)) + `":
		return nil, nil
`
	}
	content += `	default:
		return nil, nil
	}
}
`

	if err := os.WriteFile(serviceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Update with new resources
	newResources := []string{"newresource1", "newresource2", "newresource3"}

	start := time.Now()
	err = UpdateServiceFile(tmpDir, "testservice", "TestService", newResources)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("UpdateServiceFile() = %v, want nil", err)
	}

	// Should complete quickly (< 1 second)
	if duration > 1*time.Second {
		t.Errorf("UpdateServiceFile took %v, want < 1s", duration)
	}

	t.Logf("Updated service with %d new resources in %v", len(newResources), duration)
}

func TestPerformance_FileIOCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "test_performance_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to tmpDir for GenerateService
	oldDir, err := os.Getwd()
	if err == nil {
		defer os.Chdir(oldDir)
		os.Chdir(tmpDir)
	}
	if err := setupRepoPrereqs(tmpDir); err != nil {
		t.Fatalf("setupRepoPrereqs() = %v", err)
	}

	// Generate service with multiple resources
	resources := []string{"resource1", "resource2", "resource3", "resource4", "resource5"}
	err = GenerateService("testservice", "TestService", "test", resources, false)
	if err != nil {
		t.Fatalf("GenerateService() = %v, want nil", err)
	}

	// Count generated files
	fileCount := 0
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	// Should generate reasonable number of files
	// For 5 resources: service, discovery, 5 auditors, 5 tests, validation, e2e, guide = ~15 files
	expectedMin := 10
	expectedMax := 20

	if fileCount < expectedMin || fileCount > expectedMax {
		t.Logf("Generated %d files for %d resources (expected %d-%d)", fileCount, len(resources), expectedMin, expectedMax)
	}
}

