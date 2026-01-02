package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateE2ETest generates the e2e test file
func GenerateE2ETest(baseDir, serviceName, displayName string, resources []string, force bool) error {
	e2eFile := filepath.Join(baseDir, "e2e", serviceName+"_test.go")
	
	// If file exists and not forcing, try to update it instead
	if !force && fileExists(e2eFile) {
		// Check which resources already have tests
		content, err := os.ReadFile(e2eFile)
		if err != nil {
			return fmt.Errorf("file %s already exists and could not be read (use --force to overwrite): %w", e2eFile, err)
		}
		
		contentStr := string(content)
		existingResources := make(map[string]bool)
		for _, res := range resources {
			testName := "Test" + displayName + "_" + strings.Title(res) + "Audit"
			if strings.Contains(contentStr, "func "+testName) {
				existingResources[res] = true
			}
		}
		
		// Find new resources
		newResources := []string{}
		for _, r := range resources {
			if !existingResources[r] {
				newResources = append(newResources, r)
			}
		}
		
		if len(newResources) == 0 {
			// All resources already have tests
			return nil
		}
		
		// Append new test functions before the closing of the file
		testCode := generateE2ETestCode(serviceName, displayName, newResources)
		
		// Find the last closing brace (end of package)
		lastBrace := strings.LastIndex(contentStr, "}")
		if lastBrace == -1 {
			return fmt.Errorf("could not find end of file")
		}
		
		// Insert before the last closing brace
		newContent := contentStr[:lastBrace] + "\n" + testCode + "\n" + contentStr[lastBrace:]
		
		return os.WriteFile(e2eFile, []byte(newContent), 0644)
	}

	tmpl := `//go:build e2e

package e2e

import (
	"testing"
)

{{range .Resources}}
// Test{{$.DisplayName}}_{{. | Title}}Audit tests {{$.ServiceName}} {{.}} auditing
func Test{{$.DisplayName}}_{{. | Title}}Audit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{$.ServiceName}}:
    - name: test-{{.}}-check
      description: Test {{.}} check
      service: {{$.ServiceName}}
      resource: {{.}}
      check:
        status: active
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for {{$.ServiceName}}/{{.}} results
	{{.}}Results := results.FilterByService("{{$.ServiceName}}").FilterByResourceType("{{.}}")

	{{.}}Results.LogSummary(t)

	// Basic assertions
	if {{.}}Results.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during {{.}} audit", {{.}}Results.Errors)
	}
}

{{end}}
`

	data := struct {
		ServiceName string
		DisplayName string
		Resources   []string
	}{
		ServiceName: serviceName,
		DisplayName: displayName,
		Resources:   resources,
	}

	funcMap := template.FuncMap{
		"Title": strings.Title,
	}

	t, err := template.New("e2etest").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(e2eFile, t, data)
}

// generateE2ETestCode generates test function code for resources
func generateE2ETestCode(serviceName, displayName string, resources []string) string {
	code := ""
	
	for _, resource := range resources {
		titleRes := strings.Title(resource)
		code += fmt.Sprintf(`// Test%s_%sAudit tests %s %s auditing
func Test%s_%sAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `+"`"+`version: v1
defaults:
  workers: 2
policies:
  - %s:
    - name: test-%s-check
      description: Test %s check
      service: %s
      resource: %s
      check:
        status: active
      action: log`+"`"+`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for %s/%s results
	%sResults := results.FilterByService("%s").FilterByResourceType("%s")

	%sResults.LogSummary(t)

	// Basic assertions
	if %sResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during %s audit", %sResults.Errors)
	}
}

`, 
			displayName, titleRes, serviceName, resource,
			displayName, titleRes,
			serviceName, resource, resource, serviceName, resource,
			serviceName, resource, resource, serviceName, resource,
			resource, resource, resource)
	}
	
	return code
}

