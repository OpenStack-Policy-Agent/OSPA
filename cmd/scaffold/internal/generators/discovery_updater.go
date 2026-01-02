package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateDiscoveryFile adds new resource discoverers to an existing discovery file
func UpdateDiscoveryFile(baseDir, serviceName, displayName string, newResources []string) error {
	filePath := filepath.Join(baseDir, "pkg", "discovery", "services", serviceName+".go")
	
	if !fileExists(filePath) {
		return fmt.Errorf("discovery file %s does not exist", filePath)
	}

	// Read existing file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading discovery file: %w", err)
	}

	contentStr := string(content)

	// Check which resources already have discoverers
	existingResources := make(map[string]bool)
	for _, res := range newResources {
		discovererName := displayName + strings.Title(res) + "Discoverer"
		if strings.Contains(contentStr, "type "+discovererName) {
			existingResources[res] = true
		}
	}

	// Filter out resources that already exist
	resourcesToAdd := []string{}
	for _, res := range newResources {
		if !existingResources[res] {
			resourcesToAdd = append(resourcesToAdd, res)
		}
	}

	if len(resourcesToAdd) == 0 {
		return nil // Nothing to add
	}

	// Generate discoverer code for new resources
	discovererCode := generateDiscovererCode(serviceName, displayName, resourcesToAdd)

	// Append to end of file (safe; avoids injecting inside an existing function's closing brace)
	newContent := contentStr + "\n\n" + discovererCode + "\n"

	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// generateDiscovererCode generates discoverer code for resources
func generateDiscovererCode(serviceName, displayName string, resources []string) string {
	code := ""
	
	for _, resource := range resources {
		titleRes := strings.Title(resource)
		code += fmt.Sprintf(`// %s%sDiscoverer discovers %s resources of type %s
type %s%sDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *%s%sDiscoverer) ResourceType() string {
	return %q
}

// Discover discovers resources and sends them to the returned channel
func (d *%s%sDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error) {
	opts := %ss.ListOpts{}
	if allTenants {
		opts.AllTenants = true
		// Adjust based on API: opts.TenantID = "" or similar
	}
	pager := %ss.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		resourceList, err := %ss.Extract%ss(page)
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
		%q,
		func(r interface{}) string {
			return r.(%ss.%s).ID
		},
		func(r interface{}) string {
			// Adjust based on resource structure - may be TenantID, ProjectID, or nested field
			return r.(%ss.%s).TenantID
		},
	)

	return DiscoverPaged(ctx, client, %q, d.ResourceType(), pager, extract, createJob)
}

`, 
			displayName, titleRes, serviceName, resource,
			displayName, titleRes,
			displayName, titleRes, resource,
			displayName, titleRes,
			titleRes,           // %ss.ListOpts{}
			titleRes,           // %ss.List(...)
			titleRes, titleRes, // %ss.Extract%ss(...)
			serviceName,        // SimpleJobCreator service
			titleRes, titleRes, // r.(%ss.%s).ID
			titleRes, titleRes, // r.(%ss.%s).TenantID
			serviceName)        // DiscoverPaged service
	}

	return code
}

