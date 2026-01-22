package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateAuditorFiles generates auditor implementation files for each resource.
func GenerateAuditorFiles(baseDir, serviceName, displayName string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateAuditorFilesWithSpecs(baseDir, serviceName, displayName, specs)
}

func generateAuditorFilesWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec) error {
	auditDir := filepath.Join(baseDir, "pkg", "audit", serviceName)

	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return fmt.Errorf("creating audit directory: %w", err)
	}

	tmpl := `package {{.ServiceName}}

import (
	"context"
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	// TODO: Import the gophercloud resource type for {{.ResourceName}}.
	// Example: "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

// {{.ResourceTitle}}Auditor audits {{.ServiceName}}/{{.ResourceName}} resources.
//
// Allowed checks: {{JoinOrNone .Checks}}
// Allowed actions: {{JoinOrNone .Actions}}
//
// TODO: Cast 'resource' to the correct gophercloud type and implement checks.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/{{.ServiceName}}
type {{.ResourceTitle}}Auditor struct{}

func (a *{{.ResourceTitle}}Auditor) ResourceType() string {
	return "{{.ResourceName}}"
}

func (a *{{.ResourceTitle}}Auditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	// TODO: Cast resource to the correct type.
	// Example: r := resource.(servers.Server)
	//
	// Then populate the result:
	//   result.ResourceID = r.ID
	//   result.ResourceName = r.Name
	//   result.ProjectID = r.TenantID
	//   result.Status = r.Status
	//   result.UpdatedAt = r.Updated
	//
	// Implement checks based on rule.Check fields:
	//   - Status: compare r.Status with rule.Check.Status
	//   - AgeGT: compare time.Since(r.Updated) with rule.Check.AgeGT
	//   - Unused: implement resource-specific unused detection
	//   - ExemptNames: skip if r.Name matches any exempt pattern

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   "unknown",
		ResourceName: "unknown",
		ProjectID:    "",
		Compliant:    true,
		Rule:         rule,
		Status:       "",
	}

	_ = resource
	return result, nil
}

func (a *{{.ResourceTitle}}Auditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx
	_ = client
	_ = resource

	// TODO: Implement remediation actions.
	// Cast client to *gophercloud.ServiceClient.
	// Allowed actions: {{JoinOrNone .Actions}}
	//
	// Example for delete:
	//   c := client.(*gophercloud.ServiceClient)
	//   r := resource.(servers.Server)
	//   return servers.Delete(c, r.ID).ExtractErr()

	switch rule.Action {
	case "log":
		return nil
	default:
		return fmt.Errorf("%s/%s: action %q not implemented", "{{.ServiceName}}", "{{.ResourceName}}", rule.Action)
	}
}
`

	funcMap := template.FuncMap{
		"Pascal":     ToPascal,
		"JoinOrNone": JoinOrNone,
	}

	for _, resource := range resources {
		filePath := filepath.Join(auditDir, resource.Name+".go")

		data := struct {
			ServiceName   string
			DisplayName   string
			ResourceName  string
			ResourceTitle string
			ResourceDesc  string
			Checks        []string
			Actions       []string
		}{
			ServiceName:   serviceName,
			DisplayName:   displayName,
			ResourceName:  resource.Name,
			ResourceTitle: ToPascal(resource.Name),
			ResourceDesc:  resource.Description,
			Checks:        append([]string{}, resource.Checks...),
			Actions:       append([]string{}, resource.Actions...),
		}

		t, err := template.New("auditor").Funcs(funcMap).Parse(tmpl)
		if err != nil {
			return fmt.Errorf("parsing template for %s: %w", resource.Name, err)
		}

		if err := writeFile(filePath, t, data); err != nil {
			return fmt.Errorf("writing auditor file for %s: %w", resource.Name, err)
		}
	}

	return nil
}
