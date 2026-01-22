package generators

import (
	"fmt"
	"path/filepath"
	"text/template"
)

// GenerateUnitTests generates unit test files for each resource auditor.
func GenerateUnitTests(baseDir, serviceName, displayName string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateUnitTestsWithSpecs(baseDir, serviceName, displayName, specs)
}

func generateUnitTestsWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec) error {
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

	// TODO: Replace with the real gophercloud type once the auditor is implemented.
	resource := map[string]interface{}{"id": "test-id", "name": "test-resource"}

	rule := &policy.Rule{
		Name:     "test-rule",
		Service:  "{{.ServiceName}}",
		Resource: "{{.ResourceName}}",
		Check:    policy.CheckConditions{Status: "active"},
		Action:   "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result == nil {
		t.Fatal("Check() returned nil result")
	}
	if result.RuleID != rule.Name {
		t.Errorf("Result.RuleID = %q, want %q", result.RuleID, rule.Name)
	}
}

func Test{{.ResourceTitle}}Auditor_Fix(t *testing.T) {
	t.Skip("Fix() requires a mock gophercloud client")
}
`

	funcMap := template.FuncMap{
		"Pascal":     ToPascal,
		"JoinOrNone": JoinOrNone,
	}

	for _, resource := range resources {
		filePath := filepath.Join(auditDir, resource.Name+"_test.go")

		data := struct {
			ServiceName   string
			DisplayName   string
			ResourceName  string
			ResourceTitle string
			Checks        []string
			Actions       []string
		}{
			ServiceName:   serviceName,
			DisplayName:   displayName,
			ResourceName:  resource.Name,
			ResourceTitle: ToPascal(resource.Name),
			Checks:        append([]string{}, resource.Checks...),
			Actions:       append([]string{}, resource.Actions...),
		}

		t, err := template.New("unittest").Funcs(funcMap).Parse(tmpl)
		if err != nil {
			return fmt.Errorf("parsing unit test template for %s: %w", resource.Name, err)
		}

		if err := writeFile(filePath, t, data); err != nil {
			return fmt.Errorf("writing unit test file for %s: %w", resource.Name, err)
		}
	}

	return nil
}
