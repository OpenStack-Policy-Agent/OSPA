package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateAuditorFiles generates auditor implementation files for each resource
func GenerateAuditorFiles(baseDir, serviceName, displayName string, resources []string, force bool) error {
	auditDir := filepath.Join(baseDir, "pkg", "audit", serviceName)
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return fmt.Errorf("creating audit directory: %w", err)
	}

	tmpl := `package {{.ServiceName}}

import (
	"context"
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// {{.ResourceTitle}}Auditor audits {{.ServiceName}} resources of type {{.ResourceName}}
//
// TODO(OSPA): Replace placeholder logic with real field extraction + rule evaluation for {{.ServiceName}}/{{.ResourceName}}.
type {{.ResourceTitle}}Auditor struct{}

// ResourceType returns the resource type this auditor handles
func (a *{{.ResourceTitle}}Auditor) ResourceType() string {
	return "{{.ResourceName}}"
}

// Check evaluates a resource against a policy rule
func (a *{{.ResourceTitle}}Auditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx
	_ = resource

	// TODO(OSPA): Parse 'resource' into the correct OpenStack SDK type for {{.ServiceName}}/{{.ResourceName}}.
	// Populate ResourceID/ResourceName/ProjectID/Status/UpdatedAt, and implement checks (status, age_gt, unused, etc.).
	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   "unknown",
		ResourceName: "unknown",
		ProjectID:    "",
		Compliant:    true,
		Rule:         rule,
		Status:       "",
	}

	return result, nil
}

// Fix applies remediation to a resource based on the rule action
func (a *{{.ResourceTitle}}Auditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx
	_ = client
	_ = resource

	// TODO(OSPA): Implement remediation actions using the correct OpenStack client calls:
	// - delete: delete the resource
	// - tag: apply policy tag/metadata
	// - log: no-op (already supported)
	switch rule.Action {
	case "log":
		return nil
	default:
		return fmt.Errorf("%s.%s fix action %q not implemented", "{{.ServiceName}}", "{{.ResourceName}}", rule.Action)
	}
}
`

	funcMap := template.FuncMap{
		"Title":  strings.Title,
		"Pascal": ToPascal,
	}

	for _, resource := range resources {
		filePath := filepath.Join(auditDir, resource+".go")
		
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

		t, err := template.New("auditor").Funcs(funcMap).Parse(tmpl)
		if err != nil {
			return fmt.Errorf("parsing template for %s: %w", resource, err)
		}

		if err := writeFile(filePath, t, data); err != nil {
			return fmt.Errorf("writing auditor file for %s: %w", resource, err)
		}
	}

	return nil
}

