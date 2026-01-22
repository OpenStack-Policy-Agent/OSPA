package generators

import (
	"fmt"
	"path/filepath"
	"text/template"
)

// GenerateServiceFile generates the service implementation file.
func GenerateServiceFile(baseDir, serviceName, displayName, serviceType string, resources []string, force bool) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateServiceFileWithSpecs(baseDir, serviceName, displayName, serviceType, specs, force)
}

func generateServiceFileWithSpecs(baseDir, serviceName, displayName, serviceType string, resources []ResourceSpec, force bool) error {
	filePath := filepath.Join(baseDir, "pkg", "services", "services", serviceName+".go")

	// If file exists and not forcing, try to update it instead
	if !force && fileExists(filePath) {
		// Check which resources are new
		existingResources, err := extractResourcesFromServiceFile(filePath)
		if err != nil {
			return fmt.Errorf("file %s already exists and could not be analyzed (use --force to overwrite): %w", filePath, err)
		}

		// Find new resources
		existingSet := make(map[string]bool)
		for _, r := range existingResources {
			existingSet[r] = true
		}

		newResources := []ResourceSpec{}
		for _, r := range resources {
			if !existingSet[r.Name] {
				newResources = append(newResources, r)
			}
		}

		if len(newResources) == 0 {
			// All resources already exist
			return nil
		}

		// Update existing file with new resources
		return UpdateServiceFile(baseDir, serviceName, displayName, namesFromSpecs(newResources))
	}

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

// {{.DisplayName}}Service implements the Service interface for {{.DisplayName}}
//
// Supported resources:{{range .Resources}}
//   - {{.Name}}: {{.Description}}
//     Checks: {{JoinOrNone .Checks}}
//     Actions: {{JoinOrNone .Actions}}{{end}}
//
// TODO(OSPA): Ensure pkg/auth/auth.go has Get{{.DisplayName}}Client() that returns a
// *gophercloud.ServiceClient for service type "{{.ServiceType}}". The scaffold tool adds this
// method automatically, but verify it is correct for your cloud/provider.
//
// TODO(OSPA): Next steps to finish this service:
//   1. Implement discovery in pkg/discovery/services/{{.ServiceName}}.go for each resource.
//   2. Implement auditors in pkg/audit/{{.ServiceName}}/ and update rule evaluation.
//   3. Update allowed checks/actions per resource in cmd/scaffold/internal/registry/config/{{.ServiceName}}.yaml.
//   4. Run unit tests and e2e tests once real discovery/auditing exists.
type {{.DisplayName}}Service struct{}

func init() {
	rootservices.MustRegister(&{{.DisplayName}}Service{})
	// Register all supported resources for automatic validation{{range .Resources}}
	rootservices.RegisterResource("{{$.ServiceName}}", "{{.Name}}"){{end}}
}

// Name returns the service name
func (s *{{.DisplayName}}Service) Name() string {
	return "{{.ServiceName}}"
}

// GetClient returns an authenticated service client
func (s *{{.DisplayName}}Service) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
	return session.Get{{.DisplayName}}Client()
}

// GetResourceAuditor returns an auditor for the given resource type
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

// GetResourceDiscoverer returns a discoverer for the given resource type
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
