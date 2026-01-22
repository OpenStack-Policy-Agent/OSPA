package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateDiscoveryFile generates the discovery implementation file.
func GenerateDiscoveryFile(baseDir, serviceName, displayName string, resources []string, force bool) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateDiscoveryFileWithSpecs(baseDir, serviceName, displayName, specs, force)
}

func generateDiscoveryFileWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec, force bool) error {
	filePath := filepath.Join(baseDir, "pkg", "discovery", "services", serviceName+".go")

	// If file exists and not forcing, try to update it instead
	if !force && fileExists(filePath) {
		// Check which resources already have discoverers
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("file %s already exists and could not be read (use --force to overwrite): %w", filePath, err)
		}

		contentStr := string(content)
		existingResources := make(map[string]bool)
		for _, res := range resources {
			discovererName := displayName + ToPascal(res.Name) + "Discoverer"
			if strings.Contains(contentStr, "type "+discovererName) {
				existingResources[res.Name] = true
			}
		}

		// Find new resources
		newResources := []ResourceSpec{}
		for _, r := range resources {
			if !existingResources[r.Name] {
				newResources = append(newResources, r)
			}
		}

		if len(newResources) == 0 {
			// All resources already have discoverers
			return nil
		}

		// Update existing file with new discoverers
		return UpdateDiscoveryFile(baseDir, serviceName, displayName, namesFromSpecs(newResources))
	}

	tmpl := `package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
)

{{range .Resources}}
// {{$.DisplayName}}{{.Name | Pascal}}Discoverer discovers {{$.ServiceName}} resources of type {{.Name}}.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement discovery by listing {{$.ServiceName}} {{.Name}} resources from OpenStack:
// - Call the appropriate gophercloud API
// - Handle pagination
// - Emit discovery.Job{Service, ResourceType, ResourceID, ProjectID, Resource}
// - Respect allTenants where applicable
// Discovery traits: pagination={{.Discovery.Pagination}}, all_tenants={{.Discovery.AllTenants}}, regions={{.Discovery.Regions}}
// TODO(OSPA): Add unit tests in pkg/discovery/services/{{$.ServiceName}}_test.go once implemented.
type {{$.DisplayName}}{{.Name | Pascal}}Discoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *{{$.DisplayName}}{{.Name | Pascal}}Discoverer) ResourceType() string {
	return "{{.Name}}"
}

// Discover discovers resources and sends them to the returned channel
func (d *{{$.DisplayName}}{{.Name | Pascal}}Discoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants

	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}

{{end}}
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
		"Pascal": ToPascal,
	}

	t, err := template.New("discovery").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(filePath, t, data)
}
