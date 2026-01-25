package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateDiscoveryFile generates the discovery implementation file.
func GenerateDiscoveryFile(baseDir, serviceName, displayName string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateDiscoveryFileWithSpecs(baseDir, serviceName, displayName, specs)
}

func generateDiscoveryFileWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec) error {
	filePath := filepath.Join(baseDir, "pkg", "discovery", "services", serviceName+".go")

	// Check if file exists and filter out resources that already have implementations
	if existingContent, err := os.ReadFile(filePath); err == nil {
		resources = filterUnimplementedDiscoverers(string(existingContent), displayName, resources)
		if len(resources) == 0 {
			fmt.Printf("Info: All requested resources already have discoverer implementations in %s\n", filePath)
			return nil
		}
		// If file exists but we have new resources to add, append them
		return appendDiscoverers(filePath, string(existingContent), serviceName, displayName, resources)
	}

	tmpl := `package services

import (
	"context"

	discovery "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/gophercloud/gophercloud"
	// TODO: Import the correct gophercloud package for {{.ServiceName}}.
	// Example for Nova: "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	// See: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
)

{{range .Resources}}
// {{$.DisplayName}}{{.Name | Pascal}}Discoverer discovers {{$.ServiceName}}/{{.Name}} resources.
//
// TODO: Implement Discover() using gophercloud to list {{.Name}} resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/{{$.ServiceName}}
//
// Discovery hints from registry:
//   pagination: {{.Discovery.Pagination}}
//   all_tenants: {{.Discovery.AllTenants}}
//   regions: {{.Discovery.Regions}}
type {{$.DisplayName}}{{.Name | Pascal}}Discoverer struct{}

func (d *{{$.DisplayName}}{{.Name | Pascal}}Discoverer) ResourceType() string {
	return "{{.Name}}"
}

func (d *{{$.DisplayName}}{{.Name | Pascal}}Discoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List {{.Name}} resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "{{$.ServiceName}}", ResourceType: "{{.Name}}", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

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

// filterUnimplementedDiscoverers returns only resources that don't have real implementations.
// A "real implementation" is one that doesn't have the TODO placeholder in its Discover method.
func filterUnimplementedDiscoverers(content, displayName string, resources []ResourceSpec) []ResourceSpec {
	var unimplemented []ResourceSpec
	for _, r := range resources {
		if !hasDiscovererImplementation(content, displayName, r.Name) {
			unimplemented = append(unimplemented, r)
		} else {
			fmt.Printf("Info: %s%sDiscoverer already has an implementation, skipping\n", displayName, ToPascal(r.Name))
		}
	}
	return unimplemented
}

// hasDiscovererImplementation checks if a discoverer already exists in the file.
// Returns true if the discoverer type exists (regardless of whether it's a placeholder or real).
func hasDiscovererImplementation(content, displayName, resourceName string) bool {
	discovererType := displayName + ToPascal(resourceName) + "Discoverer"

	// Check if the type exists - if it does, don't regenerate it
	typeDecl := "type " + discovererType + " struct{}"
	return strings.Contains(content, typeDecl)
}

// appendDiscoverers appends new discoverer implementations to an existing file.
func appendDiscoverers(filePath, existingContent, serviceName, displayName string, resources []ResourceSpec) error {
	tmpl := `
{{range .Resources}}
// {{$.DisplayName}}{{.Name | Pascal}}Discoverer discovers {{$.ServiceName}}/{{.Name}} resources.
//
// TODO: Implement Discover() using gophercloud to list {{.Name}} resources.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/{{$.ServiceName}}
//
// Discovery hints from registry:
//   pagination: {{.Discovery.Pagination}}
//   all_tenants: {{.Discovery.AllTenants}}
//   regions: {{.Discovery.Regions}}
type {{$.DisplayName}}{{.Name | Pascal}}Discoverer struct{}

func (d *{{$.DisplayName}}{{.Name | Pascal}}Discoverer) ResourceType() string {
	return "{{.Name}}"
}

func (d *{{$.DisplayName}}{{.Name | Pascal}}Discoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job)

	go func() {
		defer close(ch)

		// TODO: List {{.Name}} resources using gophercloud and send jobs.
		// Example pattern:
		//   pages, err := <resource>.List(client, <opts>).AllPages()
		//   resources, err := <resource>.ExtractResources(pages)
		//   for _, r := range resources {
		//       select {
		//       case <-ctx.Done():
		//           return
		//       case ch <- discovery.Job{Service: "{{$.ServiceName}}", ResourceType: "{{.Name}}", ResourceID: r.ID, ProjectID: r.TenantID, Resource: r}:
		//       }
		//   }
		_ = ctx
		_ = client
		_ = allTenants
	}()

	return ch, nil
}
{{end}}`

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

	t, err := template.New("discovery_append").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	var newContent strings.Builder
	if err := t.Execute(&newContent, data); err != nil {
		return err
	}

	// Append to existing content
	finalContent := strings.TrimRight(existingContent, "\n\t ") + "\n" + newContent.String()

	return os.WriteFile(filePath, []byte(finalContent), 0644)
}
