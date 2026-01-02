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

// ServiceAnalysis contains information about existing service implementation
type ServiceAnalysis struct {
	ServiceExists     bool
	DiscoveryExists   bool
	ValidationExists  bool
	E2ETestExists     bool
	AuthMethodExists  bool
	ExistingResources []string
	MissingResources  []string
	MissingFiles      []string
}

// AnalyzeService analyzes an existing service to determine what's implemented and what's missing
func AnalyzeService(baseDir, serviceName string, requestedResources []string) (*ServiceAnalysis, error) {
	analysis := &ServiceAnalysis{
		ExistingResources: []string{},
		MissingResources:  []string{},
		MissingFiles:      []string{},
	}

	// Check service file
	serviceFile := filepath.Join(baseDir, "pkg", "services", "services", serviceName+".go")
	analysis.ServiceExists = fileExists(serviceFile)
	if analysis.ServiceExists {
		// Parse existing resources from service file
		resources, err := extractResourcesFromServiceFile(serviceFile)
		if err != nil {
			return nil, fmt.Errorf("analyzing service file: %w", err)
		}
		analysis.ExistingResources = resources
	} else {
		analysis.MissingFiles = append(analysis.MissingFiles, serviceFile)
	}

	// Check discovery file
	discoveryFile := filepath.Join(baseDir, "pkg", "discovery", "services", serviceName+".go")
	analysis.DiscoveryExists = fileExists(discoveryFile)

	// Check validation file
	validationFile := filepath.Join(baseDir, "pkg", "policy", "validation", serviceName+".go")
	analysis.ValidationExists = fileExists(validationFile)

	// Check e2e test file
	e2eTestFile := filepath.Join(baseDir, "e2e", serviceName+"_test.go")
	analysis.E2ETestExists = fileExists(e2eTestFile)

	// Check auth method (read auth.go and check for method)
	authFile := filepath.Join(baseDir, "pkg", "auth", "auth.go")
	authMethodName := fmt.Sprintf("Get%sClient", getDisplayName(serviceName))
	analysis.AuthMethodExists = checkAuthMethodExists(authFile, authMethodName)

	// Determine missing resources
	existingSet := make(map[string]bool)
	for _, r := range analysis.ExistingResources {
		existingSet[r] = true
	}

	for _, reqRes := range requestedResources {
		if !existingSet[reqRes] {
			analysis.MissingResources = append(analysis.MissingResources, reqRes)
		}
	}

	// Check for missing auditor files
	auditDir := filepath.Join(baseDir, "pkg", "audit", serviceName)
	for _, resource := range requestedResources {
		auditorFile := filepath.Join(auditDir, resource+".go")
		if !fileExists(auditorFile) {
			analysis.MissingFiles = append(analysis.MissingFiles, auditorFile)
		}
		// Check test file
		testFile := filepath.Join(auditDir, resource+"_test.go")
		if !fileExists(testFile) {
			analysis.MissingFiles = append(analysis.MissingFiles, testFile)
		}
	}

	return analysis, nil
}

// extractResourcesFromServiceFile parses a service file to extract registered resources
func extractResourcesFromServiceFile(filePath string) ([]string, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	resources := []string{}

	// Look for RegisterResource calls in init() function
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			// Check if this is a RegisterResource call (either RegisterResource(...) or pkg.RegisterResource(...))
			isRegisterResource := false
			switch fn := x.Fun.(type) {
			case *ast.Ident:
				isRegisterResource = fn.Name == "RegisterResource"
			case *ast.SelectorExpr:
				isRegisterResource = fn.Sel != nil && fn.Sel.Name == "RegisterResource"
			}
			if !isRegisterResource {
				break
			}

			// Extract resource name from arguments: RegisterResource("<service>", "<resource>")
					if len(x.Args) >= 2 {
				if lit, ok := x.Args[1].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							resource := strings.Trim(lit.Value, `"`)
							resources = append(resources, resource)
				}
			}
		}
		return true
	})

	return resources, nil
}

// checkAuthMethodExists checks if an auth method exists in auth.go
func checkAuthMethodExists(authFile, methodName string) bool {
	if !fileExists(authFile) {
		return false
	}

	src, err := os.ReadFile(authFile)
	if err != nil {
		return false
	}

	// Simple string check for method signature
	content := string(src)
	return strings.Contains(content, fmt.Sprintf("func (s *Session) %s(", methodName))
}

// getDisplayName converts service name to display name (e.g., "glance" -> "Glance")
func getDisplayName(serviceName string) string {
	if len(serviceName) == 0 {
		return ""
	}
	return strings.ToUpper(serviceName[:1]) + serviceName[1:]
}

