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
	RichChecks  []CheckSpec
	Actions     []string
	Discovery   DiscoverySpec
}

// CheckSpec carries rich metadata for a single check, sourced from the
// registry's CheckInfo. Used by the policy guide generator.
type CheckSpec struct {
	Name        string
	Type        string
	Description string
	Category    string
	Severity    string
	GuideRef    string
}

// DiscoverySpec captures discovery characteristics for a resource.
type DiscoverySpec struct {
	Pagination bool
	AllTenants bool
	Regions    bool
}

// HasCommonChecks returns true if the resource supports the universal
// checks (status, age_gt, unused, exempt_names).
func (r ResourceSpec) HasCommonChecks() bool {
	for _, c := range r.Checks {
		if c == "status" || c == "age_gt" || c == "unused" || c == "exempt_names" {
			return true
		}
	}
	return false
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

		var richChecks []CheckSpec
		for _, rc := range meta.RichChecks {
			richChecks = append(richChecks, CheckSpec{
				Name:        rc.Name,
				Type:        rc.Type,
				Description: rc.Description,
				Category:    rc.Category,
				Severity:    rc.Severity,
				GuideRef:    rc.GuideRef,
			})
		}

		specs = append(specs, ResourceSpec{
			Name:        resName,
			Description: meta.Description,
			Checks:      append([]string{}, meta.Checks...),
			RichChecks:  richChecks,
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

// GuideChecklistSpec is a generator-facing copy of registry.GuideChecklistItem.
type GuideChecklistSpec struct {
	ID          string
	Description string
	Section     string
	Manual      bool
}

// buildGuideChecklist loads the OpenStack Security Guide checklist items
// for a service from the registry. Returns nil if the service or its
// checklist is not found.
func buildGuideChecklist(serviceName string) []GuideChecklistSpec {
	serviceName = strings.ToLower(strings.TrimSpace(serviceName))
	info, err := registry.GetServiceInfo(serviceName)
	if err != nil {
		return nil
	}
	specs := make([]GuideChecklistSpec, 0, len(info.GuideChecklist))
	for _, item := range info.GuideChecklist {
		specs = append(specs, GuideChecklistSpec{
			ID:          item.ID,
			Description: item.Description,
			Section:     item.Section,
			Manual:      item.Manual,
		})
	}
	return specs
}
