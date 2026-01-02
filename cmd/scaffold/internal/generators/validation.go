package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateValidationFile generates or updates a validation file for a service
func GenerateValidationFile(baseDir, serviceName, displayName string, resources []string, force bool) error {
	validationDir := filepath.Join(baseDir, "pkg", "policy", "validation")
	validationFile := filepath.Join(validationDir, fmt.Sprintf("%s.go", serviceName))

	// Check if file exists
	exists := fileExists(validationFile)
	
	// If file exists and we're not forcing, we need to update it instead of overwriting
	if exists && !force {
		return updateValidationFile(validationFile, serviceName, displayName, resources)
	}

	// Generate new validation file
	tmpl := `package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// {{.DisplayName}}Validator validates {{.DisplayName}} service policies
//
// TODO(OSPA): Tighten validation rules for {{.ServiceName}} over time:
// - Require at least one check condition per rule
// - Validate supported check fields per resource
// - Validate allowed enum values (status/protocol/ethertype/etc.)
type {{.DisplayName}}Validator struct{}

func init() {
	policy.RegisterValidator(&{{.DisplayName}}Validator{})
}

func (v *{{.DisplayName}}Validator) ServiceName() string {
	return "{{.ServiceName}}"
}

func (v *{{.DisplayName}}Validator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
{{range .Resources}}
	case "{{.}}":
		// Placeholder validation: accept any checks for now.
		// TODO(OSPA): Add real validation for {{$.ServiceName}}/{{.}}.
		_ = check

{{end}}
	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for {{.ServiceName}} service", ruleName, resourceType)
	}

	return nil
}
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

	t, err := template.New("validation").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err := writeFile(validationFile, t, data); err != nil {
		return fmt.Errorf("writing validation file: %w", err)
	}

	// No need to update pkg/policy/validator.go anymore.
	//
	// Validators are registered via init() in pkg/policy/validation (single package),
	// and the application entrypoints should import that package once (blank import)
	// to enable resource-specific policy validation.
	return nil
}

// updateValidationFile updates an existing validation file with new resources
func updateValidationFile(filePath, serviceName, displayName string, resources []string) error {
	// Read existing file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading existing validation file: %w", err)
	}

	contentStr := string(content)

	// Check which resources are already in the switch statement
	existingResources := make(map[string]bool)
	for _, resource := range resources {
		if strings.Contains(contentStr, fmt.Sprintf(`case "%s":`, resource)) {
			existingResources[resource] = true
		}
	}

	// Find resources that need to be added
	newResources := []string{}
	for _, resource := range resources {
		if !existingResources[resource] {
			newResources = append(newResources, resource)
		}
	}

	if len(newResources) == 0 {
		// All resources already exist, nothing to update
		return nil
	}

	// Find the switch statement and add new cases before the default case
	defaultCaseIndex := strings.Index(contentStr, "\n\tdefault:")
	if defaultCaseIndex == -1 {
		// No default case found, append before the closing brace
		closingBraceIndex := strings.LastIndex(contentStr, "\treturn nil\n}")
		if closingBraceIndex == -1 {
			return fmt.Errorf("could not find insertion point in validation file")
		}

		// Build new cases
		newCases := "\n"
		for _, resource := range newResources {
			newCases += fmt.Sprintf(`	case "%s":
		// TODO(OSPA): Add validation rules for %s resource.
		_ = check

`, resource, resource)
		}

		newContent := contentStr[:closingBraceIndex] + newCases + contentStr[closingBraceIndex:]
		return os.WriteFile(filePath, []byte(newContent), 0644)
	}

	// Insert before default case
	newCases := "\n"
	for _, resource := range newResources {
		newCases += fmt.Sprintf(`	case "%s":
		// TODO(OSPA): Add validation rules for %s resource.
		_ = check

`, resource, resource)
	}

	newContent := contentStr[:defaultCaseIndex] + newCases + contentStr[defaultCaseIndex:]
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

