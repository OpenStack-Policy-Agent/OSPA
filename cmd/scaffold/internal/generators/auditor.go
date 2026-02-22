package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// checksQuoted formats a string slice as Go quoted literals for a template,
// e.g. ["status", "age_gt"] => `"status", "age_gt"`.
func checksQuoted(checks []string) string {
	quoted := make([]string, len(checks))
	for i, c := range checks {
		quoted[i] = fmt.Sprintf("%q", c)
	}
	return strings.Join(quoted, ", ")
}

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
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/common"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	// TODO: Import the gophercloud resource type for {{.ResourceName}}.
	// Example: "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

// TODO: Replace this placeholder adapter with the real gophercloud type fields.
// Example for servers.Server:
//
//	type {{.ResourceName}}Adapter struct{ r servers.Server }
//	func (a {{.ResourceName}}Adapter) GetID() string           { return a.r.ID }
//	func (a {{.ResourceName}}Adapter) GetName() string         { return a.r.Name }
//	func (a {{.ResourceName}}Adapter) GetProjectID() string    { return a.r.TenantID }
//	func (a {{.ResourceName}}Adapter) GetStatus() string       { return a.r.Status }
//	func (a {{.ResourceName}}Adapter) GetCreatedAt() time.Time { return a.r.Created }
//	func (a {{.ResourceName}}Adapter) GetUpdatedAt() time.Time { return a.r.Updated }
type {{.ResourceName}}Adapter struct{ r interface{} }

func (a {{.ResourceName}}Adapter) GetID() string           { return "unknown" }
func (a {{.ResourceName}}Adapter) GetName() string         { return "unknown" }
func (a {{.ResourceName}}Adapter) GetProjectID() string    { return "" }
func (a {{.ResourceName}}Adapter) GetStatus() string       { return "" }
func (a {{.ResourceName}}Adapter) GetCreatedAt() time.Time { return time.Time{} }
func (a {{.ResourceName}}Adapter) GetUpdatedAt() time.Time { return time.Time{} }

// {{.ResourceTitle}}Auditor audits {{.ServiceName}}/{{.ResourceName}} resources.
//
// Allowed checks: {{JoinOrNone .Checks}}
// Allowed actions: {{JoinOrNone .Actions}}
//
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/{{.ServiceName}}
type {{.ResourceTitle}}Auditor struct{}

func (a *{{.ResourceTitle}}Auditor) ResourceType() string {
	return "{{.ResourceName}}"
}

func (a *{{.ResourceTitle}}Auditor) ImplementedChecks() []string {
	return []string{ {{ChecksQuoted .Checks}} }
}

func (a *{{.ResourceTitle}}Auditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	// TODO: Replace with the correct gophercloud type assertion.
	// Example: r, ok := resource.(servers.Server)
	adapter := {{.ResourceName}}Adapter{r: resource}

	result := common.BuildBaseResult(adapter, rule)

	exempt, err := common.RunCommonChecks(adapter, rule, result)
	if exempt || err != nil {
		return result, err
	}

	// TODO: Implement resource-specific unused detection.
	if rule.Check.Unused {
		result.Observation = "unused check not yet implemented"
	}

	// TODO: Implement domain-specific checks for {{.ResourceName}}.

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
		"Pascal":       ToPascal,
		"JoinOrNone":   JoinOrNone,
		"ChecksQuoted": checksQuoted,
	}

	for _, resource := range resources {
		filePath := filepath.Join(auditDir, resource.Name+".go")

		// Check if file exists and has a real implementation
		if hasAuditorImplementation(filePath, resource.Name) {
			fmt.Printf("Info: %sAuditor already has an implementation, skipping\n", ToPascal(resource.Name))
			continue
		}

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

// hasAuditorImplementation checks if an auditor file already exists.
// Returns true if the file exists (regardless of whether it's a placeholder or real).
func hasAuditorImplementation(filePath, resourceName string) bool {
	_, err := os.Stat(filePath)
	return err == nil // File exists
}
