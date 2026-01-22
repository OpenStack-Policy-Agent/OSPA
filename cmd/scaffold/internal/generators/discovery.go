package generators

import (
	"path/filepath"
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
