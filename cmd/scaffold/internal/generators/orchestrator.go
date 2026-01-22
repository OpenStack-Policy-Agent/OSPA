package generators

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// GenerateService orchestrates the generation of all files for a service.
// Existing resources are preserved and new resources are added.
func GenerateService(serviceName, displayName, serviceType string, resources []string) error {
	baseDir := "."

	// Get existing resources from service file (if it exists)
	existingResources := getExistingResources(baseDir, serviceName)

	// Merge: all resources = existing + new (deduplicated)
	allResources := mergeResources(existingResources, resources)

	// Only generate auditor/test files for resources that don't already have files
	newResources := findNewResources(baseDir, serviceName, resources)

	// Generate service file with ALL resources
	if err := GenerateServiceFile(baseDir, serviceName, displayName, serviceType, allResources); err != nil {
		return fmt.Errorf("generating service file: %w", err)
	}

	// Generate discovery file with ALL resources
	if err := GenerateDiscoveryFile(baseDir, serviceName, displayName, allResources); err != nil {
		return fmt.Errorf("generating discovery file: %w", err)
	}

	// Generate auditor files only for NEW resources (skip existing)
	if len(newResources) > 0 {
		if err := GenerateAuditorFiles(baseDir, serviceName, displayName, newResources); err != nil {
			return fmt.Errorf("generating auditor files: %w", err)
		}
	}

	// Generate auth client method (idempotent - skips if exists)
	if err := GenerateAuthMethod(baseDir, serviceName, displayName, serviceType); err != nil {
		return fmt.Errorf("generating auth method: %w", err)
	}

	// Generate unit test files only for NEW resources (skip existing)
	if len(newResources) > 0 {
		if err := GenerateUnitTests(baseDir, serviceName, displayName, newResources); err != nil {
			return fmt.Errorf("generating unit tests: %w", err)
		}
	}

	// Generate e2e test file with ALL resources
	if err := GenerateE2ETest(baseDir, serviceName, displayName, allResources); err != nil {
		return fmt.Errorf("generating e2e test: %w", err)
	}

	// Generate policy guide with ALL resources
	if err := GeneratePolicyGuide(baseDir, serviceName, displayName, serviceType, allResources); err != nil {
		return fmt.Errorf("generating policy guide: %w", err)
	}

	// Generate validation file with ALL resources
	if err := GenerateValidationFile(baseDir, serviceName, displayName, allResources); err != nil {
		return fmt.Errorf("generating validation file: %w", err)
	}

	return nil
}

// getExistingResources extracts registered resources from an existing service file.
func getExistingResources(baseDir, serviceName string) []string {
	serviceFile := filepath.Join(baseDir, "pkg", "services", "services", serviceName+".go")

	src, err := os.ReadFile(serviceFile)
	if err != nil {
		return nil // File doesn't exist
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, serviceFile, src, parser.ParseComments)
	if err != nil {
		return nil
	}

	var resources []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for RegisterResource calls
		isRegisterResource := false
		switch fn := call.Fun.(type) {
		case *ast.Ident:
			isRegisterResource = fn.Name == "RegisterResource"
		case *ast.SelectorExpr:
			isRegisterResource = fn.Sel != nil && fn.Sel.Name == "RegisterResource"
		}
		if !isRegisterResource {
			return true
		}

		// Extract resource name from second argument
		if len(call.Args) >= 2 {
			if lit, ok := call.Args[1].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				resource := strings.Trim(lit.Value, `"`)
				resources = append(resources, resource)
			}
		}
		return true
	})

	return resources
}

// mergeResources merges existing and new resources, removing duplicates.
func mergeResources(existing, new []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, r := range existing {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}
	for _, r := range new {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}

	return result
}

// findNewResources returns resources that don't have existing auditor files.
func findNewResources(baseDir, serviceName string, resources []string) []string {
	auditDir := filepath.Join(baseDir, "pkg", "audit", serviceName)

	var newResources []string
	for _, res := range resources {
		auditorFile := filepath.Join(auditDir, res+".go")
		if _, err := os.Stat(auditorFile); os.IsNotExist(err) {
			newResources = append(newResources, res)
		}
	}

	return newResources
}
