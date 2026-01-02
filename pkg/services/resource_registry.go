package services

import (
	"fmt"
	"sync"
)

// ResourceRegistry tracks which resources are supported by which services.
// This is populated automatically when services are registered.
type ResourceRegistry struct {
	mu              sync.RWMutex
	serviceResources map[string]map[string]bool // service -> resource -> true
}

var globalResourceRegistry = &ResourceRegistry{
	serviceResources: make(map[string]map[string]bool),
}

// RegisterResource registers a resource type for a service.
// This is called automatically by services during registration.
func (r *ResourceRegistry) RegisterResource(serviceName, resourceType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.serviceResources[serviceName] == nil {
		r.serviceResources[serviceName] = make(map[string]bool)
	}
	r.serviceResources[serviceName][resourceType] = true
}

// IsResourceSupported checks if a service supports a resource type.
func (r *ResourceRegistry) IsResourceSupported(serviceName, resourceType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if resources, ok := r.serviceResources[serviceName]; ok {
		return resources[resourceType]
	}
	return false
}

// GetServiceResources returns all resource types supported by a service.
func (r *ResourceRegistry) GetServiceResources(serviceName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if resources, ok := r.serviceResources[serviceName]; ok {
		result := make([]string, 0, len(resources))
		for resourceType := range resources {
			result = append(result, resourceType)
		}
		return result
	}
	return nil
}

// GetAllResources returns a map of all service -> resource mappings.
func (r *ResourceRegistry) GetAllResources() map[string]map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]map[string]bool)
	for service, resources := range r.serviceResources {
		result[service] = make(map[string]bool)
		for resource := range resources {
			result[service][resource] = true
		}
	}
	return result
}

// RegisterResource is a convenience function for the global registry.
func RegisterResource(serviceName, resourceType string) {
	globalResourceRegistry.RegisterResource(serviceName, resourceType)
}

// IsResourceSupported is a convenience function for the global registry.
func IsResourceSupported(serviceName, resourceType string) bool {
	return globalResourceRegistry.IsResourceSupported(serviceName, resourceType)
}

// GetServiceResources is a convenience function for the global registry.
func GetServiceResources(serviceName string) []string {
	return globalResourceRegistry.GetServiceResources(serviceName)
}

