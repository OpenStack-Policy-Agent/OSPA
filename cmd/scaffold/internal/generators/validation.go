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
	"strings"

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
		if err := validateAllowedChecks(check, []string{ {{range $i, $c := .Checks}}{{if $i}}, {{end}}"{{$c}}"{{end}} }); err != nil {
			return fmt.Errorf("rule %q: %w", ruleName, err)
		}

{{end}}
	default:
		return fmt.Errorf("rule %q: unsupported resource type %q for {{.ServiceName}} service", ruleName, resourceType)
	}

	return nil
}

func validateAllowedChecks(check *policy.CheckConditions, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = true
	}

	if !hasAnyCheck(check, allowedSet) {
		return fmt.Errorf("check must specify at least one of: %s", strings.Join(allowed, ", "))
	}

	disallowed := findDisallowedChecks(check, allowedSet)
	if len(disallowed) > 0 {
		return fmt.Errorf("check specifies unsupported fields: %s", strings.Join(disallowed, ", "))
	}

	return nil
}

func hasAnyCheck(check *policy.CheckConditions, allowed map[string]bool) bool {
	for name := range allowed {
		if isCheckSet(check, name) {
			return true
		}
	}
	return false
}

func findDisallowedChecks(check *policy.CheckConditions, allowed map[string]bool) []string {
	var disallowed []string
	for _, name := range []string{
		"direction",
		"ethertype",
		"protocol",
		"port",
		"remote_ip_prefix",
		"status",
		"age_gt",
		"unused",
		"exempt_names",
		"exempt_metadata",
		"image_name",
	} {
		if isCheckSet(check, name) && !allowed[name] {
			disallowed = append(disallowed, name)
		}
	}
	return disallowed
}

func isCheckSet(check *policy.CheckConditions, name string) bool {
	switch name {
	case "direction":
		return check.Direction != ""
	case "ethertype":
		return check.Ethertype != ""
	case "protocol":
		return check.Protocol != ""
	case "port":
		return check.Port != 0
	case "remote_ip_prefix":
		return check.RemoteIPPrefix != ""
	case "status":
		return check.Status != ""
	case "age_gt":
		return check.AgeGT != ""
	case "unused":
		return check.Unused
	case "exempt_names":
		return len(check.ExemptNames) > 0
	case "exempt_metadata":
		return check.ExemptMetadata != nil
	case "image_name":
		return len(check.ImageName) > 0
	default:
		return false
	}
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
