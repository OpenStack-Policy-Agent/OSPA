package services

import "github.com/OpenStack-Policy-Agent/OSPA/pkg/catalog"

// RegisterResource registers a resource type for a service.
// Kept in this package for backwards compatibility; implemented in pkg/catalog
// to avoid import cycles between policy/services/audit.
func RegisterResource(serviceName, resourceType string) {
	catalog.RegisterResource(serviceName, resourceType)
}

// IsResourceSupported checks if a service supports a resource type.
func IsResourceSupported(serviceName, resourceType string) bool {
	return catalog.IsResourceSupported(serviceName, resourceType)
}

// GetServiceResources returns all resource types supported by a service.
func GetServiceResources(serviceName string) []string {
	return catalog.GetServiceResources(serviceName)
}
