package generators

import (
	"fmt"
	"path/filepath"
	"text/template"
)

// GenerateValidationFile generates a validation file for a service.
func GenerateValidationFile(baseDir, serviceName, displayName string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateValidationFileWithSpecs(baseDir, serviceName, displayName, specs)
}

func generateValidationFileWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec) error {
	validationDir := filepath.Join(baseDir, "pkg", "policy", "validation")
	validationFile := filepath.Join(validationDir, fmt.Sprintf("%s.go", serviceName))

	tmpl := `package validation

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// {{.DisplayName}}Validator validates {{.DisplayName}} service policies.
//
// TODO: Review and update the allowed checks for each resource type.
// The checks listed below are defaults from the registry. You may need to:
//   - Add resource-specific checks (e.g., direction, protocol for security rules)
//   - Remove checks that don't apply to this resource
//   - Update the CheckConditions struct in pkg/policy/types.go if adding new check types
//
// See pkg/policy/validation/helpers.go for the validateAllowedChecks implementation.
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
	case "{{.Name}}":
		// TODO: Update allowed checks to match the auditor implementation in pkg/audit/{{$.ServiceName}}/{{.Name}}.go
		if err := validateAllowedChecks(check, []string{ {{range $i, $c := .Checks}}{{if $i}}, {{end}}"{{$c}}"{{end}} }); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

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
		Resources   []ResourceSpec
	}{
		ServiceName: serviceName,
		DisplayName: displayName,
		Resources:   resources,
	}

	t, err := template.New("validation").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	return writeFile(validationFile, t, data)
}
