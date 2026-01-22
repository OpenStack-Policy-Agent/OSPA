package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// GeneratePolicyGuide generates the policy guide markdown file.
func GeneratePolicyGuide(baseDir, serviceName, displayName, serviceType string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generatePolicyGuideWithSpecs(baseDir, serviceName, displayName, serviceType, specs)
}

func generatePolicyGuideWithSpecs(baseDir, serviceName, displayName, serviceType string, resources []ResourceSpec) error {
	examplesDir := filepath.Join(baseDir, "examples", "policies")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		return fmt.Errorf("creating examples/policies directory: %w", err)
	}

	guideFile := filepath.Join(examplesDir, serviceName+"-policy-guide.md")

	tmpl := `# Policy Guide: {{.DisplayName}} ({{.ServiceName}})

This guide explains how to write policies for {{.DisplayName}} resources in OSPA.

## Service Overview

**Service Name:** ` + "`{{.ServiceName}}`" + `
**Display Name:** {{.DisplayName}}
**OpenStack Service Type:** {{.ServiceType}}

## Supported Resources

{{range .Resources}}
### {{.Name | Title}}

**Resource Type:** ` + "`{{.Name}}`" + `

**Allowed Actions:** {{JoinOrNone .Actions}}
**Allowed Checks:** {{JoinOrNone .Checks}}

{{end}}

## Policy Structure

All policies for {{.DisplayName}} follow this structure:

{{printf "%c%c%c" 96 96 96}}yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - {{.ServiceName}}:
    - name: rule-name
      description: Rule description
      service: {{.ServiceName}}
      resource: <resource_type>
      check:
        # Check conditions (see below)
      action: log|delete|tag
{{printf "%c%c%c" 96 96 96}}

## Check Conditions

### Common Check Conditions

The following check conditions are available for most resources:

#### Status Check

Check resources by their status:

{{printf "%c%c%c" 96 96 96}}yaml
check:
  status: active|inactive|available|unavailable|DOWN|UP
{{printf "%c%c%c" 96 96 96}}

**Example:**
{{printf "%c%c%c" 96 96 96}}yaml
- name: find-inactive-resources
  description: Find inactive {{.ServiceName}} resources
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    status: inactive
  action: log
{{printf "%c%c%c" 96 96 96}}

#### Age Check

Find resources older than a specified age:

{{printf "%c%c%c" 96 96 96}}yaml
check:
  age_gt: 30d  # Options: 7d, 30d, 90d, 1h, 24h, etc.
{{printf "%c%c%c" 96 96 96}}

**Supported units:**
- ` + "`d`" + ` or ` + "`day`" + ` or ` + "`days`" + ` - Days
- ` + "`h`" + ` or ` + "`hour`" + ` or ` + "`hours`" + ` - Hours
- ` + "`m`" + ` or ` + "`min`" + ` or ` + "`minute`" + ` or ` + "`minutes`" + ` - Minutes

**Example:**
{{printf "%c%c%c" 96 96 96}}yaml
- name: find-old-resources
  description: Find resources older than 30 days
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    age_gt: 30d
  action: log
{{printf "%c%c%c" 96 96 96}}

#### Unused Check

Find resources that are not being used:

{{printf "%c%c%c" 96 96 96}}yaml
check:
  unused: true
{{printf "%c%c%c" 96 96 96}}

**Example:**
{{printf "%c%c%c" 96 96 96}}yaml
- name: find-unused-resources
  description: Find unused {{.ServiceName}} resources
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    unused: true
  action: log
{{printf "%c%c%c" 96 96 96}}

#### Exemptions

Exclude specific resources from checks:

{{printf "%c%c%c" 96 96 96}}yaml
check:
  status: active
  exempt_names:
    - default
    - system-resource
{{printf "%c%c%c" 96 96 96}}

**Example:**
{{printf "%c%c%c" 96 96 96}}yaml
- name: find-active-except-default
  description: Find active resources except default ones
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    status: active
    exempt_names:
      - default
  action: log
{{printf "%c%c%c" 96 96 96}}

## Actions

### Log Action

Log violations without taking any action:

{{printf "%c%c%c" 96 96 96}}yaml
action: log
{{printf "%c%c%c" 96 96 96}}

**Example:**
{{printf "%c%c%c" 96 96 96}}yaml
- name: audit-resources
  description: Audit {{.ServiceName}} resources
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    status: inactive
  action: log
{{printf "%c%c%c" 96 96 96}}

### Delete Action

Delete non-compliant resources (use with caution):

{{printf "%c%c%c" 96 96 96}}yaml
action: delete
{{printf "%c%c%c" 96 96 96}}

**Example:**
{{printf "%c%c%c" 96 96 96}}yaml
- name: cleanup-old-resources
  description: Delete resources older than 90 days
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    age_gt: 90d
  action: delete
{{printf "%c%c%c" 96 96 96}}

**Note:** The ` + "`--fix`" + ` flag must be set when running the agent for delete actions to take effect.

### Tag Action

Tag non-compliant resources with metadata:

{{printf "%c%c%c" 96 96 96}}yaml
action: tag
tag_name: audit-tag-name
action_tag_name: "Display Name for Tag"
{{printf "%c%c%c" 96 96 96}}

**Example:**
{{printf "%c%c%c" 96 96 96}}yaml
- name: tag-old-resources
  description: Tag resources older than 30 days
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    age_gt: 30d
  action: tag
  tag_name: audit-old-resource
  action_tag_name: "Old Resource"
{{printf "%c%c%c" 96 96 96}}

## Resource-Specific Examples

{{range .Resources}}
### {{.Name | Title}} Examples

#### Example 1: Find Inactive {{.Name | Title}} Resources

{{printf "%c%c%c" 96 96 96}}yaml
- name: find-inactive-{{.Name}}
  description: Find inactive {{.Name}} resources
  service: {{$.ServiceName}}
  resource: {{.Name}}
  check:
    status: inactive
  action: log
{{printf "%c%c%c" 96 96 96}}

#### Example 2: Find Old {{.Name | Title}} Resources

{{printf "%c%c%c" 96 96 96}}yaml
- name: find-old-{{.Name}}
  description: Find {{.Name}} resources older than 30 days
  service: {{$.ServiceName}}
  resource: {{.Name}}
  check:
    age_gt: 30d
  action: log
{{printf "%c%c%c" 96 96 96}}

#### Example 3: Cleanup Unused {{.Name | Title}} Resources

{{printf "%c%c%c" 96 96 96}}yaml
- name: cleanup-unused-{{.Name}}
  description: Delete unused {{.Name}} resources
  service: {{$.ServiceName}}
  resource: {{.Name}}
  check:
    unused: true
    exempt_names:
      - default
  action: delete
{{printf "%c%c%c" 96 96 96}}

#### Example 4: Tag Old {{.Name | Title}} Resources

{{printf "%c%c%c" 96 96 96}}yaml
- name: tag-old-{{.Name}}
  description: Tag {{.Name}} resources older than 7 days
  service: {{$.ServiceName}}
  resource: {{.Name}}
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-{{.Name}}
  action_tag_name: "Old {{.Name | Title}}"
{{printf "%c%c%c" 96 96 96}}

{{end}}

## Complete Policy Example

Here's a complete policy file example for {{.DisplayName}}:

{{printf "%c%c%c" 96 96 96}}yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - {{.ServiceName}}:{{range .Resources}}
    - name: audit-{{.Name}}
      description: Audit {{.Name}} resources
      service: {{$.ServiceName}}
      resource: {{.Name}}
      check:
        status: active
      action: log
    - name: cleanup-old-{{.Name}}
      description: Find {{.Name}} resources older than 90 days
      service: {{$.ServiceName}}
      resource: {{.Name}}
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log{{end}}
{{printf "%c%c%c" 96 96 96}}

## OpenStack Documentation References

For more information about {{.DisplayName}} resources and their properties:

- **OpenStack {{.DisplayName}} API Documentation:** https://docs.openstack.org/api-ref/{{.ServiceName}}/
- **{{.DisplayName}} Service Guide:** https://docs.openstack.org/{{.ServiceName}}/latest/

## Testing Your Policy

1. **Validate the policy:**
   {{printf "%c%c%c" 96 96 96}}bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out /dev/null
   {{printf "%c%c%c" 96 96 96}}

2. **Run in audit mode (safe):**
   {{printf "%c%c%c" 96 96 96}}bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.jsonl
   {{printf "%c%c%c" 96 96 96}}

3. **Apply remediations (use with caution):**
   {{printf "%c%c%c" 96 96 96}}bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.jsonl --fix
   {{printf "%c%c%c" 96 96 96}}

## Notes

- All check conditions are optional, but at least one should be specified
- Multiple check conditions are combined with AND logic (all must match)
- The ` + "`exempt_names`" + ` list allows you to exclude specific resources by name
- Age checks use the resource's ` + "`UpdatedAt`" + ` timestamp, falling back to ` + "`CreatedAt`" + ` if not available
- Status values are case-sensitive and should match OpenStack API responses exactly

## Troubleshooting

**Policy validation fails:**
- Ensure service name matches exactly: ` + "`{{.ServiceName}}`" + `
- Verify resource type is supported: {{range $i, $r := .Resources}}{{if $i}}, {{end}}` + "`{{$r}}`" + `{{end}}
- Check YAML syntax is correct

**No resources found:**
- Verify resources exist in your OpenStack project
- Use ` + "`--all-tenants`" + ` flag if resources are in other projects (requires admin)
- Check OpenStack API endpoints are accessible

**Actions not working:**
- Ensure ` + "`--fix`" + ` flag is set for delete/tag actions
- Verify you have permissions to modify resources
- Check action-specific requirements (e.g., ` + "`tag_name`" + ` for tag action)

## See Also

- [OSPA Development Guide](../../docs/DEVELOPMENT.md)
- [OSPA Architecture Guide](../../docs/ARCHITECTURE.md)
- [Example Policies](../policies.yaml)
`

	data := struct {
		ServiceName string
		DisplayName string
		ServiceType string
		Resources   []ResourceSpec
	}{
		ServiceName: serviceName,
		DisplayName: displayName,
		ServiceType: serviceType,
		Resources:   resources,
	}

	funcMap := template.FuncMap{
		"Title":      ToPascal,
		"JoinOrNone": JoinOrNone,
	}

	t, err := template.New("policyguide").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing policy guide template: %w", err)
	}

	return writeFile(guideFile, t, data)
}
