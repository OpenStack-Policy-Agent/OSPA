package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateValidationFile generates or updates a validation file for a service.
// If the file already has a proper validation structure (switch/default),
// only new resource cases are appended, preserving hand-tuned checks.
// Otherwise, the file is generated from scratch.
func GenerateValidationFile(baseDir, serviceName, displayName string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}

	validationFile := filepath.Join(baseDir, "pkg", "policy", "validation", fmt.Sprintf("%s.go", serviceName))

	if existing, err := os.ReadFile(validationFile); err == nil {
		content := string(existing)
		if strings.Contains(content, "switch resourceType") && strings.Contains(content, "\tdefault:") {
			return appendValidationCases(validationFile, content, specs)
		}
	}

	return generateValidationFileWithSpecs(baseDir, serviceName, displayName, specs)
}

// appendValidationCases inserts new case blocks into an existing validation
// file's switch statement, right before the default: label.
func appendValidationCases(filePath, content string, resources []ResourceSpec) error {
	var toAdd []ResourceSpec
	for _, r := range resources {
		caseLabel := fmt.Sprintf(`case "%s":`, r.Name)
		if strings.Contains(content, caseLabel) {
			fmt.Printf("Info: Validation case for %q already exists, skipping\n", r.Name)
		} else {
			toAdd = append(toAdd, r)
		}
	}

	if len(toAdd) == 0 {
		fmt.Printf("Info: All requested resources already have validation cases\n")
		return nil
	}

	var buf strings.Builder
	for _, r := range toAdd {
		fmt.Fprintf(&buf, "\tcase %q:\n\t\tif err := validateAllowedChecks(check, %s); err != nil {\n\t\t\treturn fmt.Errorf(\"rule %%q: %%w\", ruleName, err)\n\t\t}\n\n", r.Name, formatChecksSlice(r.Checks))
	}

	idx := strings.Index(content, "\n\tdefault:")
	if idx == -1 {
		return fmt.Errorf("could not find default case in %s", filePath)
	}

	final := content[:idx+1] + buf.String() + content[idx+1:]
	return os.WriteFile(filePath, []byte(final), 0644)
}

// formatChecksSlice renders a Go []string literal, e.g. []string{"status", "age_gt"}.
func formatChecksSlice(checks []string) string {
	quoted := make([]string, len(checks))
	for i, c := range checks {
		quoted[i] = fmt.Sprintf("%q", c)
	}
	return "[]string{" + strings.Join(quoted, ", ") + "}"
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
		if err := validateAllowedChecks(check, {{ChecksSlice .Checks}}); err != nil {
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

	funcMap := template.FuncMap{
		"ChecksSlice": formatChecksSlice,
	}

	t, err := template.New("validation").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	return writeFile(validationFile, t, data)
}
