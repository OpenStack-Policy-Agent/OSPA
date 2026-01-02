package generators

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateDiscoveryFile generates the discovery implementation file
func GenerateDiscoveryFile(baseDir, serviceName, displayName string, resources []string, force bool) error {
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
			discovererName := displayName + strings.Title(res) + "Discoverer"
			if strings.Contains(contentStr, "type "+discovererName) {
				existingResources[res] = true
			}
		}
		
		// Find new resources
		newResources := []string{}
		for _, r := range resources {
			if !existingResources[r] {
				newResources = append(newResources, r)
			}
		}
		
		if len(newResources) == 0 {
			// All resources already have discoverers
			return nil
		}
		
		// Update existing file with new discoverers
		return UpdateDiscoveryFile(baseDir, serviceName, displayName, newResources)
	}

	tmpl := `package discovery

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/{{.ServiceName}}/v2/{{.ResourcePackage}}"
	"github.com/gophercloud/gophercloud/pagination"
)

{{range .Resources}}
// {{$.DisplayName}}{{. | Title}}Discoverer discovers {{$.ServiceName}} resources of type {{.}}
type {{$.DisplayName}}{{. | Title}}Discoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *{{$.DisplayName}}{{. | Title}}Discoverer) ResourceType() string {
	return "{{.}}"
}

// Discover discovers resources and sends them to the returned channel
func (d *{{$.DisplayName}}{{. | Title}}Discoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error) {
	opts := {{. | Title}}s.ListOpts{}
	if allTenants {
		opts.AllTenants = true
		// Adjust based on API: opts.TenantID = "" or similar
	}
	pager := {{. | Title}}s.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		resourceList, err := {{. | Title}}s.Extract{{. | Title}}s(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(resourceList))
		for i := range resourceList {
			resources[i] = resourceList[i]
		}
		return resources, nil
	}

	createJob := SimpleJobCreator(
		"{{$.ServiceName}}",
		func(r interface{}) string {
			return r.({{. | Title}}s.{{. | Title}}).ID
		},
		func(r interface{}) string {
			// Adjust based on resource structure - may be TenantID, ProjectID, or nested field
			return r.({{. | Title}}s.{{. | Title}}).TenantID
		},
	)

	return DiscoverPaged(ctx, client, "{{$.ServiceName}}", d.ResourceType(), pager, extract, createJob)
}

{{end}}
`

	data := struct {
		ServiceName     string
		DisplayName     string
		ResourcePackage string
		Resources       []string
	}{
		ServiceName:     serviceName,
		DisplayName:     displayName,
		ResourcePackage: serviceName, // Default to service name, may need adjustment
		Resources:       resources,
	}

	funcMap := template.FuncMap{
		"Title": strings.Title,
	}

	t, err := template.New("discovery").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(filePath, t, data)
}

