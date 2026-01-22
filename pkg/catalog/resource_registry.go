package catalog

import "sync"

// ResourceRegistry tracks which resources are supported by which services.
// This package intentionally has no dependencies on higher-level packages
// (policy/services/audit) to avoid import cycles.
type ResourceRegistry struct {
	mu               sync.RWMutex
	serviceResources map[string]map[string]bool // service -> resource -> true
}

var global = &ResourceRegistry{
	serviceResources: make(map[string]map[string]bool),
}

// RegisterResource registers a resource type for a service.
func RegisterResource(serviceName, resourceType string) {
	global.mu.Lock()
	defer global.mu.Unlock()

	if global.serviceResources[serviceName] == nil {
		global.serviceResources[serviceName] = make(map[string]bool)
	}
	global.serviceResources[serviceName][resourceType] = true
}

// IsResourceSupported checks if a service supports a resource type.
func IsResourceSupported(serviceName, resourceType string) bool {
	global.mu.RLock()
	defer global.mu.RUnlock()

	if resources, ok := global.serviceResources[serviceName]; ok {
		return resources[resourceType]
	}
	return false
}

// GetServiceResources returns all resource types supported by a service.
func GetServiceResources(serviceName string) []string {
	global.mu.RLock()
	defer global.mu.RUnlock()

	if resources, ok := global.serviceResources[serviceName]; ok {
		result := make([]string, 0, len(resources))
		for resourceType := range resources {
			result = append(result, resourceType)
		}
		return result
	}
	return nil
}

// GetSupportedResources returns a copy of all service -> resource mappings.
func GetSupportedResources() map[string]map[string]bool {
	global.mu.RLock()
	defer global.mu.RUnlock()

	result := make(map[string]map[string]bool)
	for service, resources := range global.serviceResources {
		result[service] = make(map[string]bool)
		for resource := range resources {
			result[service][resource] = true
		}
	}
	return result
}
