package generators

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/astutil"
)

// GenerateValidationFile generates or updates a validation file for a service.
func GenerateValidationFile(baseDir, serviceName, displayName string, resources []string, force bool) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateValidationFileWithSpecs(baseDir, serviceName, displayName, specs, force)
}

func generateValidationFileWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec, force bool) error {
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
func updateValidationFile(filePath, serviceName, displayName string, resources []ResourceSpec) error {
	fset, file, err := astutil.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing existing validation file: %w", err)
	}

	fn := astutil.FindFunc(file, "ValidateResource")
	sw := astutil.FindSwitchOnIdent(fn, "resourceType")
	if sw == nil {
		return fmt.Errorf("ValidateResource switch not found in %s", filePath)
	}

	existing := astutil.CaseValues(sw)
	var cases []*ast.CaseClause
	for _, resource := range resources {
		if existing[strconv.Quote(resource.Name)] {
			continue
		}
		cases = append(cases, validationCase(resource))
	}

	if len(cases) == 0 {
		return nil
	}

	astutil.InsertCasesBeforeDefault(sw, cases)
	return astutil.WriteFile(filePath, fset, file)
}

func validationCase(resource ResourceSpec) *ast.CaseClause {
	return &ast.CaseClause{
		List: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(resource.Name)}},
		Body: []ast.Stmt{
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun:  ast.NewIdent("validateAllowedChecks"),
							Args: []ast.Expr{ast.NewIdent("check"), stringSlice(resource.Checks)},
						},
					},
				},
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent("err"),
					Op: token.NEQ,
					Y:  ast.NewIdent("nil"),
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: ast.NewIdent("fmt.Errorf"),
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("rule %q: %w")},
										ast.NewIdent("ruleName"),
										ast.NewIdent("err"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func stringSlice(items []string) ast.Expr {
	elts := make([]ast.Expr, 0, len(items))
	for _, item := range items {
		elts = append(elts, &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(item)})
	}
	return &ast.CompositeLit{
		Type: &ast.ArrayType{Elt: ast.NewIdent("string")},
		Elts: elts,
	}
}
