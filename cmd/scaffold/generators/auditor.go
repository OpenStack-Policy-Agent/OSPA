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
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/{{.ServiceName}}/v2/{{.ResourcePackage}}"
)

// {{.ResourceTitle}}Auditor audits {{.ServiceName}} resources of type {{.ResourceName}}
type {{.ResourceTitle}}Auditor struct{}

// ResourceType returns the resource type this auditor handles
func (a *{{.ResourceTitle}}Auditor) ResourceType() string {
	return "{{.ResourceName}}"
}

// Check evaluates a resource against a policy rule
func (a *{{.ResourceTitle}}Auditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	res, ok := resource.({{.ResourcePackage}}.{{.ResourceTitle}})
	if !ok {
		return nil, fmt.Errorf("expected {{.ResourcePackage}}.{{.ResourceTitle}}, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   res.ID,
		ResourceName: res.Name, // Adjust based on resource structure
		ProjectID:    res.TenantID, // Adjust based on resource structure
		Compliant:    true,
		Rule:         rule,
		Status:       res.Status, // If available
		UpdatedAt:    res.UpdatedAt, // If available
	}

	check := rule.Check

	// Check status
	if check.Status != "" {
		if res.Status != check.Status {
			return result, nil // Status doesn't match, but that's not a violation
		}
	}

	// Check age
	if check.AgeGT != "" {
		ageThreshold, err := check.ParseAgeGT()
		if err != nil {
			return nil, fmt.Errorf("failed to parse age_gt: %w", err)
		}

		evalTime := res.UpdatedAt
		if evalTime.IsZero() {
			evalTime = res.CreatedAt
		}
		if !evalTime.IsZero() {
			age := time.Now().Sub(evalTime)
			if age >= ageThreshold {
				result.Compliant = false
				result.Observation = fmt.Sprintf("Resource is %s old (>= %s)", age.Round(time.Hour*24), ageThreshold)
			}
		}
	}

	// Check exemptions
	if check.ExemptNames != nil {
		for _, exemptName := range check.ExemptNames {
			if res.Name == exemptName {
				return result, nil // Exempt, so compliant
			}
		}
	}

	// Add more check conditions as needed based on resource type

	return result, nil
}

// Fix applies remediation to a resource based on the rule action
func (a *{{.ResourceTitle}}Auditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	res, ok := resource.({{.ResourcePackage}}.{{.ResourceTitle}})
	if !ok {
		return fmt.Errorf("expected {{.ResourcePackage}}.{{.ResourceTitle}}, got %T", resource)
	}

	serviceClient, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	switch rule.Action {
	case "delete":
		return {{.ResourcePackage}}.Delete(serviceClient, res.ID).ExtractErr()
	case "tag":
		// Implement tagging logic based on resource type
		tagName := rule.TagName
		if tagName == "" {
			tagName = rule.ActionTagName
		}
		if tagName == "" {
			return fmt.Errorf("tag_name or action_tag_name is required for tag action")
		}
		// TODO: Implement tagging for {{.ResourceName}}
		return fmt.Errorf("tag action not yet implemented for {{.ResourceName}}")
	case "log":
		return nil
	default:
		return fmt.Errorf("unsupported action %q", rule.Action)
	}
}
`

	funcMap := template.FuncMap{
		"Title": strings.Title,
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
			ResourcePackage string
		}{
			ServiceName:     serviceName,
			DisplayName:     displayName,
			ResourceName:    resource,
			ResourceTitle:   strings.Title(resource),
			ResourcePackage: serviceName, // May need adjustment
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

