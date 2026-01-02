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
type {{.DisplayName}}Validator struct{}

func init() {
	Register(&{{.DisplayName}}Validator{})
}

func (v *{{.DisplayName}}Validator) ServiceName() string {
	return "{{.ServiceName}}"
}

func (v *{{.DisplayName}}Validator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
	switch resourceType {
{{range .Resources}}
	case "{{.}}":
		// TODO: Add validation rules for {{.}} resource
		// Example validations:
		// - Check required fields (e.g., status, age_gt, unused)
		// - Validate field values (e.g., status must be one of allowed values)
		// - Ensure at least one check condition is specified
		// Example:
		// if check.Status == "" && check.AgeGT == "" {
		//     return fmt.Errorf("rule %q: check must specify at least one of status or age_gt", ruleName)
		// }

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

	// Update validator.go to import the new validation package
	return updateValidatorImports(baseDir, serviceName)
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
		// TODO: Add validation rules for %s resource
		// Example validations:
		// - Check required fields (e.g., status, age_gt, unused)
		// - Validate field values (e.g., status must be one of allowed values)
		// - Ensure at least one check condition is specified
		// Example:
		// if check.Status == "" && check.AgeGT == "" {
		//     return fmt.Errorf("rule %%q: check must specify at least one of status or age_gt", ruleName)
		// }

`, resource, resource)
		}

		newContent := contentStr[:closingBraceIndex] + newCases + contentStr[closingBraceIndex:]
		return os.WriteFile(filePath, []byte(newContent), 0644)
	}

	// Insert before default case
	newCases := "\n"
	for _, resource := range newResources {
		newCases += fmt.Sprintf(`	case "%s":
		// TODO: Add validation rules for %s resource
		// Example validations:
		// - Check required fields (e.g., status, age_gt, unused)
		// - Validate field values (e.g., status must be one of allowed values)
		// - Ensure at least one check condition is specified
		// Example:
		// if check.Status == "" && check.AgeGT == "" {
		//     return fmt.Errorf("rule %%q: check must specify at least one of status or age_gt", ruleName)
		// }

`, resource, resource)
	}

	newContent := contentStr[:defaultCaseIndex] + newCases + contentStr[defaultCaseIndex:]
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// updateValidatorImports updates validator.go to import the new validation package
func updateValidatorImports(baseDir, serviceName string) error {
	validatorFile := filepath.Join(baseDir, "pkg", "policy", "validator.go")

	content, err := os.ReadFile(validatorFile)
	if err != nil {
		return fmt.Errorf("reading validator.go: %w", err)
	}

	contentStr := string(content)

	// Check if import already exists
	importPath := fmt.Sprintf(`_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation/%s"`, serviceName)
	if strings.Contains(contentStr, importPath) {
		// Already imported
		return nil
	}

	// Find the last validation import
	lastImportIndex := strings.LastIndex(contentStr, `_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation/`)
	if lastImportIndex == -1 {
		return fmt.Errorf("could not find validation imports in validator.go")
	}

	// Find the end of that import line
	lineEnd := strings.Index(contentStr[lastImportIndex:], "\n")
	if lineEnd == -1 {
		return fmt.Errorf("could not find end of import line")
	}

	// Insert new import after the last validation import
	insertPos := lastImportIndex + lineEnd
	newImport := fmt.Sprintf("\t_ \"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy/validation/%s\"\n", serviceName)
	newContent := contentStr[:insertPos] + newImport + contentStr[insertPos:]

	return os.WriteFile(validatorFile, []byte(newContent), 0644)
}

