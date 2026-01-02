package generators

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateServiceFile generates the service implementation file
func GenerateServiceFile(baseDir, serviceName, displayName, serviceType string, resources []string, force bool) error {
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
		
		newResources := []string{}
		for _, r := range resources {
			if !existingSet[r] {
				newResources = append(newResources, r)
			}
		}
		
		if len(newResources) == 0 {
			// All resources already exist
			return nil
		}
		
		// Update existing file with new resources
		return UpdateServiceFile(baseDir, serviceName, displayName, newResources)
	}

	tmpl := `package services

import (
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/{{.ServiceName}}"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

// {{.DisplayName}}Service implements the Service interface for {{.DisplayName}}
//
// Supported resources:{{range .Resources}}
//   - {{.}}: {{$.DisplayName}} {{.}} resources{{end}}
//
// To add support for a new resource type:
//   1. Create a discoverer in pkg/discovery/services/{{.ServiceName}}.go
//   2. Create an auditor in pkg/audit/{{.ServiceName}}/
//   3. Add cases in GetResourceAuditor() and GetResourceDiscoverer() below
//   4. Register the resource in init() using RegisterResource()
type {{.DisplayName}}Service struct{}

func init() {
	MustRegister(&{{.DisplayName}}Service{})
	// Register all supported resources for automatic validation{{range .Resources}}
	RegisterResource("{{$.ServiceName}}", "{{.}}"){{end}}
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
	case "{{.}}":
		return &{{$.ServiceName}}.{{. | Title}}Auditor{}, nil
	{{- end}}
	default:
		return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
	}
}

// GetResourceDiscoverer returns a discoverer for the given resource type
func (s *{{.DisplayName}}Service) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	switch resourceType {
	{{- range .Resources}}
	case "{{.}}":
		return &discovery.{{$.DisplayName}}{{. | Title}}Discoverer{}, nil
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

	t, err := template.New("service").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(filePath, t, data)
}

