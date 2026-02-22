package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateServiceFile generates or updates the service implementation file.
// If the file already has a proper service structure (RegisterResource,
// GetResourceAuditor, GetResourceDiscoverer), only new resources are appended.
// Otherwise, the file is generated from scratch.
func GenerateServiceFile(baseDir, serviceName, displayName, serviceType string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}

	filePath := filepath.Join(baseDir, "pkg", "services", "services", serviceName+".go")

	if existing, err := os.ReadFile(filePath); err == nil {
		content := string(existing)
		if strings.Contains(content, "RegisterResource(") &&
			strings.Contains(content, "GetResourceAuditor") &&
			strings.Contains(content, "GetResourceDiscoverer") {
			return appendServiceResources(filePath, content, serviceName, displayName, specs)
		}
	}

	return generateServiceFileWithSpecs(baseDir, serviceName, displayName, serviceType, specs)
}

// appendServiceResources adds new resource registrations, auditor cases, discoverer
// cases and doc-comment lines to an existing service file without rewriting it.
func appendServiceResources(filePath, content, serviceName, displayName string, resources []ResourceSpec) error {
	var toAdd []ResourceSpec
	for _, r := range resources {
		registerCall := fmt.Sprintf(`RegisterResource("%s", "%s")`, serviceName, r.Name)
		if strings.Contains(content, registerCall) {
			fmt.Printf("Info: Service registration for %q already exists, skipping\n", r.Name)
		} else {
			toAdd = append(toAdd, r)
		}
	}

	if len(toAdd) == 0 {
		fmt.Printf("Info: All requested resources already registered in service file\n")
		return nil
	}

	var registrations, docLines, auditorCases, discovererCases strings.Builder
	for _, r := range toAdd {
		pascal := ToPascal(r.Name)
		fmt.Fprintf(&registrations, "\trootservices.RegisterResource(%q, %q)\n", serviceName, r.Name)
		fmt.Fprintf(&docLines, "//   - %s: %s\n//     Checks: %s\n//     Actions: %s\n",
			r.Name, r.Description, JoinOrNone(r.Checks), JoinOrNone(r.Actions))
		fmt.Fprintf(&auditorCases, "\tcase %q:\n\t\treturn &%s.%sAuditor{}, nil\n",
			r.Name, serviceName, pascal)
		fmt.Fprintf(&discovererCases, "\tcase %q:\n\t\treturn &discovery_services.%s%sDiscoverer{}, nil\n",
			r.Name, displayName, pascal)
	}

	// Apply insertions bottom-to-top so earlier positions stay valid.

	// 4. GetResourceDiscoverer — insert cases before default
	discSig := "func (s *" + displayName + "Service) GetResourceDiscoverer"
	if idx := strings.Index(content, discSig); idx != -1 {
		if defIdx := strings.Index(content[idx:], "\tdefault:"); defIdx != -1 {
			at := idx + defIdx
			content = content[:at] + discovererCases.String() + content[at:]
		}
	}

	// 3. GetResourceAuditor — insert cases before default
	audSig := "func (s *" + displayName + "Service) GetResourceAuditor"
	if idx := strings.Index(content, audSig); idx != -1 {
		if defIdx := strings.Index(content[idx:], "\tdefault:"); defIdx != -1 {
			at := idx + defIdx
			content = content[:at] + auditorCases.String() + content[at:]
		}
	}

	// 2. Doc comment — insert before "type XXXService struct{}"
	typeDecl := "type " + displayName + "Service struct{}"
	if idx := strings.Index(content, typeDecl); idx != -1 {
		content = content[:idx] + docLines.String() + content[idx:]
	}

	// 1. RegisterResource — insert after the last RegisterResource line
	lastReg := strings.LastIndex(content, "RegisterResource(")
	if lastReg != -1 {
		if eol := strings.Index(content[lastReg:], "\n"); eol != -1 {
			at := lastReg + eol + 1
			content = content[:at] + registrations.String() + content[at:]
		}
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

func generateServiceFileWithSpecs(baseDir, serviceName, displayName, serviceType string, resources []ResourceSpec) error {
	filePath := filepath.Join(baseDir, "pkg", "services", "services", serviceName+".go")

	tmpl := `package services

import (
	"fmt"

	rootservices "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/{{.ServiceName}}"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	discovery_services "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
	"github.com/gophercloud/gophercloud"
)

// {{.DisplayName}}Service implements the Service interface for OpenStack {{.DisplayName}}.
//
// Supported resources:{{range .Resources}}
//   - {{.Name}}: {{.Description}}
//     Checks: {{JoinOrNone .Checks}}
//     Actions: {{JoinOrNone .Actions}}{{end}}
type {{.DisplayName}}Service struct{}

func init() {
	rootservices.MustRegister(&{{.DisplayName}}Service{}){{range .Resources}}
	rootservices.RegisterResource("{{$.ServiceName}}", "{{.Name}}"){{end}}
}

func (s *{{.DisplayName}}Service) Name() string {
	return "{{.ServiceName}}"
}

func (s *{{.DisplayName}}Service) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.Get{{.DisplayName}}Client()
}

func (s *{{.DisplayName}}Service) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	switch resourceType {
	{{- range .Resources}}
	case "{{.Name}}":
		return &{{$.ServiceName}}.{{.Name | Pascal}}Auditor{}, nil
	{{- end}}
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

func (s *{{.DisplayName}}Service) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	{{- range .Resources}}
	case "{{.Name}}":
		return &discovery_services.{{$.DisplayName}}{{.Name | Pascal}}Discoverer{}, nil
	{{- end}}
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}
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
		"Pascal":     ToPascal,
		"JoinOrNone": JoinOrNone,
	}

	t, err := template.New("service").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(filePath, t, data)
}
