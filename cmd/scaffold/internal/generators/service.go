package generators

import (
	"path/filepath"
	"text/template"
)

// GenerateServiceFile generates the service implementation file.
func GenerateServiceFile(baseDir, serviceName, displayName, serviceType string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateServiceFileWithSpecs(baseDir, serviceName, displayName, serviceType, specs)
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
