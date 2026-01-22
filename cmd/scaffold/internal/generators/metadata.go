package generators

import (
	"fmt"
	"strings"

	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/registry"
)

// ResourceSpec captures metadata needed by generators.
type ResourceSpec struct {
	Name        string
	Description string
	Checks      []string
	Actions     []string
	Discovery   DiscoverySpec
}

// DiscoverySpec captures discovery characteristics for a resource.
type DiscoverySpec struct {
	Pagination bool
	AllTenants bool
	Regions    bool
}

func buildResourceSpecs(serviceName string, resources []string) ([]ResourceSpec, error) {
	serviceName = strings.ToLower(strings.TrimSpace(serviceName))
	info, err := registry.GetServiceInfo(serviceName)
	if err != nil {
		return fallbackResourceSpecs(resources), nil
	}

	specs := make([]ResourceSpec, 0, len(resources))
	for _, res := range resources {
		resName := strings.ToLower(strings.TrimSpace(res))
		meta, ok := info.Resources[resName]
		if !ok {
			specs = append(specs, fallbackResourceSpec(resName))
			continue
		}
		specs = append(specs, ResourceSpec{
			Name:        resName,
			Description: meta.Description,
			Checks:      append([]string{}, meta.Checks...),
			Actions:     append([]string{}, meta.Actions...),
			Discovery: DiscoverySpec{
				Pagination: meta.Discovery.Pagination,
				AllTenants: meta.Discovery.AllTenants,
				Regions:    meta.Discovery.Regions,
			},
		})
	}

	return specs, nil
}

func fallbackResourceSpecs(resources []string) []ResourceSpec {
	specs := make([]ResourceSpec, 0, len(resources))
	for _, res := range resources {
		specs = append(specs, fallbackResourceSpec(res))
	}
	return specs
}

func fallbackResourceSpec(resource string) ResourceSpec {
	resource = strings.ToLower(strings.TrimSpace(resource))
	desc := fmt.Sprintf("%s resources", resource)
	return ResourceSpec{
		Name:        resource,
		Description: desc,
	}
}
