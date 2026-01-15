package generators

import (
	"fmt"
	"path/filepath"
	"text/template"
)

// GenerateUnitTests generates unit test files for each resource auditor
func GenerateUnitTests(baseDir, serviceName, displayName string, resources []string, force bool) error {
	auditDir := filepath.Join(baseDir, "pkg", "audit", serviceName)

	tmpl := `package {{.ServiceName}}

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

func Test{{.ResourceTitle}}Auditor_ResourceType(t *testing.T) {
	auditor := &{{.ResourceTitle}}Auditor{}
	if got := auditor.ResourceType(); got != "{{.ResourceName}}" {
		t.Errorf("ResourceType() = %q, want %q", got, "{{.ResourceName}}")
	}
}

func Test{{.ResourceTitle}}Auditor_Check(t *testing.T) {
	auditor := &{{.ResourceTitle}}Auditor{}

	// TODO(OSPA): Replace this placeholder resource with the real SDK type used by the discoverer.
	resource := map[string]interface{}{"id": "test-id", "name": "test-resource"}

	rule := &policy.Rule{
		Name:     "test-rule",
		Service:  "{{.ServiceName}}",
		Resource: "{{.ResourceName}}",
		Check: policy.CheckConditions{
			Status: "active",
		},
		Action: "log",
	}

	ctx := context.Background()
	result, err := auditor.Check(ctx, resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result == nil {
		t.Fatal("Check() returned nil result")
	}

	if result.RuleID != rule.Name {
		t.Errorf("Result.RuleID = %q, want %q", result.RuleID, rule.Name)
	}

	// TODO(OSPA): Add assertions for ResourceID/ProjectID/Compliant/Observation once real extraction is implemented.
}

func Test{{.ResourceTitle}}Auditor_Check_AgeGT(t *testing.T) {
	t.Skip("placeholder auditor does not implement age-based checks yet")
}

func Test{{.ResourceTitle}}Auditor_Fix(t *testing.T) {
	// TODO: Implement integration test with mock client
	// This requires setting up a mock gophercloud client
	t.Skip("Fix() test requires mock client setup")
}
`

	funcMap := template.FuncMap{
		"Pascal": ToPascal,
	}

	for _, resource := range resources {
		filePath := filepath.Join(auditDir, resource+"_test.go")
		
		if !force && fileExists(filePath) {
			fmt.Printf("Warning: %s already exists, skipping (use --force to overwrite)\n", filePath)
			continue
		}

		data := struct {
			ServiceName     string
			DisplayName     string
			ResourceName    string
			ResourceTitle   string
		}{
			ServiceName:     serviceName,
			DisplayName:     displayName,
			ResourceName:    resource,
			ResourceTitle:   ToPascal(resource),
		}

		t, err := template.New("unittest").Funcs(funcMap).Parse(tmpl)
		if err != nil {
			return fmt.Errorf("parsing unit test template for %s: %w", resource, err)
		}

		if err := writeFile(filePath, t, data); err != nil {
			return fmt.Errorf("writing unit test file for %s: %w", resource, err)
		}
	}

	return nil
}

