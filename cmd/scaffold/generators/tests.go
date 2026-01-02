package generators

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateUnitTests generates unit test files for each resource auditor
func GenerateUnitTests(baseDir, serviceName, displayName string, resources []string, force bool) error {
	auditDir := filepath.Join(baseDir, "pkg", "audit", serviceName)

	tmpl := `package {{.ServiceName}}

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/{{.ServiceName}}/v2/{{.ResourcePackage}}"
)

func Test{{.ResourceTitle}}Auditor_ResourceType(t *testing.T) {
	auditor := &{{.ResourceTitle}}Auditor{}
	if got := auditor.ResourceType(); got != "{{.ResourceName}}" {
		t.Errorf("ResourceType() = %q, want %q", got, "{{.ResourceName}}")
	}
}

func Test{{.ResourceTitle}}Auditor_Check(t *testing.T) {
	auditor := &{{.ResourceTitle}}Auditor{}

	// Create a mock resource
	// TODO: Replace with actual resource struct from OpenStack client
	resource := {{.ResourcePackage}}.{{.ResourceTitle}}{
		ID:   "test-id",
		Name: "test-resource",
		// Add required fields based on actual struct
	}

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

	if result.ResourceID != resource.ID {
		t.Errorf("Result.ResourceID = %q, want %q", result.ResourceID, resource.ID)
	}
}

func Test{{.ResourceTitle}}Auditor_Check_AgeGT(t *testing.T) {
	auditor := &{{.ResourceTitle}}Auditor{}

	resource := {{.ResourcePackage}}.{{.ResourceTitle}}{
		ID:        "test-id",
		Name:      "test-resource",
		UpdatedAt: time.Now().Add(-10 * 24 * time.Hour), // 10 days ago
		// Add required fields
	}

	rule := &policy.Rule{
		Name:     "test-rule",
		Service:  "{{.ServiceName}}",
		Resource: "{{.ResourceName}}",
		Check: policy.CheckConditions{
			AgeGT: "7d", // Older than 7 days
		},
		Action: "log",
	}

	ctx := context.Background()
	result, err := auditor.Check(ctx, resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// Resource is 10 days old, should be non-compliant
	if result.Compliant {
		t.Error("Check() returned Compliant=true, want false for old resource")
	}
}

func Test{{.ResourceTitle}}Auditor_Fix(t *testing.T) {
	// TODO: Implement integration test with mock client
	// This requires setting up a mock gophercloud client
	t.Skip("Fix() test requires mock client setup")
}
`

	funcMap := template.FuncMap{
		"Title": strings.Title,
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
			ResourcePackage string
		}{
			ServiceName:     serviceName,
			DisplayName:     displayName,
			ResourceName:    resource,
			ResourceTitle:   strings.Title(resource),
			ResourcePackage: serviceName,
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

