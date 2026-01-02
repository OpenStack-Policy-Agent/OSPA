package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GeneratePolicyGuide generates the policy guide markdown file
func GeneratePolicyGuide(baseDir, serviceName, displayName, serviceType string, resources []string, force bool) error {
	// Create examples/policies directory if it doesn't exist
	examplesDir := filepath.Join(baseDir, "examples", "policies")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		return fmt.Errorf("creating examples/policies directory: %w", err)
	}

	guideFile := filepath.Join(examplesDir, serviceName+"-policy-guide.md")
	
	if !force && fileExists(guideFile) {
		fmt.Printf("Warning: %s already exists, skipping (use --force to overwrite)\n", guideFile)
		return nil
	}

	tmpl := `# Policy Guide: {{.DisplayName}} ({{.ServiceName}})

This guide explains how to write policies for {{.DisplayName}} resources in OSPA.

## Service Overview

**Service Name:** ` + "`{{.ServiceName}}`" + `
**Display Name:** {{.DisplayName}}
**OpenStack Service Type:** {{.ServiceType}}

## Supported Resources

{{range .Resources}}
### {{. | Title}}

**Resource Type:** ` + "`{{.}}`" + `

{{end}}

## Policy Structure

All policies for {{.DisplayName}} follow this structure:

\`\`\`yaml
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
\`\`\`

## Check Conditions

### Common Check Conditions

The following check conditions are available for most resources:

#### Status Check

Check resources by their status:

\`\`\`yaml
check:
  status: active|inactive|available|unavailable|DOWN|UP
\`\`\`

**Example:**
\`\`\`yaml
- name: find-inactive-resources
  description: Find inactive {{.ServiceName}} resources
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    status: inactive
  action: log
\`\`\`

#### Age Check

Find resources older than a specified age:

\`\`\`yaml
check:
  age_gt: 30d  # Options: 7d, 30d, 90d, 1h, 24h, etc.
\`\`\`

**Supported units:**
- ` + "`d`" + ` or ` + "`day`" + ` or ` + "`days`" + ` - Days
- ` + "`h`" + ` or ` + "`hour`" + ` or ` + "`hours`" + ` - Hours
- ` + "`m`" + ` or ` + "`min`" + ` or ` + "`minute`" + ` or ` + "`minutes`" + ` - Minutes

**Example:**
\`\`\`yaml
- name: find-old-resources
  description: Find resources older than 30 days
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    age_gt: 30d
  action: log
\`\`\`

#### Unused Check

Find resources that are not being used:

\`\`\`yaml
check:
  unused: true
\`\`\`

**Example:**
\`\`\`yaml
- name: find-unused-resources
  description: Find unused {{.ServiceName}} resources
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    unused: true
  action: log
\`\`\`

#### Exemptions

Exclude specific resources from checks:

\`\`\`yaml
check:
  status: active
  exempt_names:
    - default
    - system-resource
\`\`\`

**Example:**
\`\`\`yaml
- name: find-active-except-default
  description: Find active resources except default ones
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    status: active
    exempt_names:
      - default
  action: log
\`\`\`

## Actions

### Log Action

Log violations without taking any action:

\`\`\`yaml
action: log
\`\`\`

**Example:**
\`\`\`yaml
- name: audit-resources
  description: Audit {{.ServiceName}} resources
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    status: inactive
  action: log
\`\`\`

### Delete Action

Delete non-compliant resources (use with caution):

\`\`\`yaml
action: delete
\`\`\`

**Example:**
\`\`\`yaml
- name: cleanup-old-resources
  description: Delete resources older than 90 days
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    age_gt: 90d
  action: delete
\`\`\`

**Note:** The ` + "`--apply`" + ` flag must be set when running the agent for delete actions to take effect.

### Tag Action

Tag non-compliant resources with metadata:

\`\`\`yaml
action: tag
tag_name: audit-tag-name
action_tag_name: "Display Name for Tag"
\`\`\`

**Example:**
\`\`\`yaml
- name: tag-old-resources
  description: Tag resources older than 30 days
  service: {{.ServiceName}}
  resource: <resource_type>
  check:
    age_gt: 30d
  action: tag
  tag_name: audit-old-resource
  action_tag_name: "Old Resource"
\`\`\`

## Resource-Specific Examples

{{range .Resources}}
### {{. | Title}} Examples

#### Example 1: Find Inactive {{. | Title}} Resources

\`\`\`yaml
- name: find-inactive-{{.}}
  description: Find inactive {{.}} resources
  service: {{$.ServiceName}}
  resource: {{.}}
  check:
    status: inactive
  action: log
\`\`\`

#### Example 2: Find Old {{. | Title}} Resources

\`\`\`yaml
- name: find-old-{{.}}
  description: Find {{.}} resources older than 30 days
  service: {{$.ServiceName}}
  resource: {{.}}
  check:
    age_gt: 30d
  action: log
\`\`\`

#### Example 3: Cleanup Unused {{. | Title}} Resources

\`\`\`yaml
- name: cleanup-unused-{{.}}
  description: Delete unused {{.}} resources
  service: {{$.ServiceName}}
  resource: {{.}}
  check:
    unused: true
    exempt_names:
      - default
  action: delete
\`\`\`

#### Example 4: Tag Old {{. | Title}} Resources

\`\`\`yaml
- name: tag-old-{{.}}
  description: Tag {{.}} resources older than 7 days
  service: {{$.ServiceName}}
  resource: {{.}}
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-{{.}}
  action_tag_name: "Old {{. | Title}}"
\`\`\`

{{end}}

## Complete Policy Example

Here's a complete policy file example for {{.DisplayName}}:

\`\`\`yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - {{.ServiceName}}:{{range .Resources}}
    - name: audit-{{.}}
      description: Audit {{.}} resources
      service: {{$.ServiceName}}
      resource: {{.}}
      check:
        status: active
      action: log
    - name: cleanup-old-{{.}}
      description: Find {{.}} resources older than 90 days
      service: {{$.ServiceName}}
      resource: {{.}}
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log{{end}}
\`\`\`

## OpenStack Documentation References

For more information about {{.DisplayName}} resources and their properties:

- **OpenStack {{.DisplayName}} API Documentation:** https://docs.openstack.org/api-ref/{{.ServiceName}}/
- **{{.DisplayName}} Service Guide:** https://docs.openstack.org/{{.ServiceName}}/latest/

## Testing Your Policy

1. **Validate the policy:**
   \`\`\`bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out /dev/null
   \`\`\`

2. **Run in audit mode (safe):**
   \`\`\`bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.jsonl
   \`\`\`

3. **Apply remediations (use with caution):**
   \`\`\`bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.jsonl --apply
   \`\`\`

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
- Ensure ` + "`--apply`" + ` flag is set for delete/tag actions
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
		Resources   []string
	}{
		ServiceName: serviceName,
		DisplayName: displayName,
		ServiceType: serviceType,
		Resources:   resources,
	}

	funcMap := template.FuncMap{
		"Title": strings.Title,
	}

	t, err := template.New("policyguide").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing policy guide template: %w", err)
	}

	return writeFile(guideFile, t, data)
}

