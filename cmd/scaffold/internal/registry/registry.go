package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	registryOnce sync.Once
	registryErr  error
	registryData map[string]ServiceInfo
)

// ServiceDefaults defines default actions/checks for a service.
type ServiceDefaults struct {
	Actions []string `yaml:"actions"`
	Checks  []string `yaml:"checks"`
}

// DiscoveryMetadata captures discovery characteristics for a resource.
type DiscoveryMetadata struct {
	Pagination bool `yaml:"pagination"`
	AllTenants bool `yaml:"all_tenants"`
	Regions    bool `yaml:"regions"`
}

// ServiceInfo contains information about an OpenStack service
type ServiceInfo struct {
	ServiceType string
	DisplayName string
	Defaults    ServiceDefaults
	Resources   map[string]ResourceInfo
}

// ResourceInfo contains information about a resource type
type ResourceInfo struct {
	Description string
	Checks      []string
	Actions     []string
	Discovery   DiscoveryMetadata
}

// ServiceMetadata represents the YAML format for service registry files.
type ServiceMetadata struct {
	Name        string                  `yaml:"name"`
	DisplayName string                  `yaml:"display_name"`
	ServiceType string                  `yaml:"service_type"`
	Defaults    ServiceDefaults         `yaml:"defaults"`
	Resources   map[string]ResourceInfo `yaml:"resources"`
}

// ValidateService checks if a service exists in OpenStack
func ValidateService(serviceName string) error {
	serviceName = strings.ToLower(serviceName)
	if err := ensureLoaded(); err != nil {
		return err
	}
	_, exists := registryData[serviceName]
	if !exists {
		available := ListServices()
		return fmt.Errorf("service %q is not a known OpenStack service. Available services: %v", serviceName, available)
	}
	return nil
}

// ValidateResources checks if resources exist for a given service
func ValidateResources(serviceName string, resources []string) error {
	serviceName = strings.ToLower(serviceName)
	if err := ensureLoaded(); err != nil {
		return err
	}
	serviceInfo, exists := registryData[serviceName]
	if !exists {
		return ValidateService(serviceName)
	}

	var invalidResources []string
	for _, resource := range resources {
		resource = strings.ToLower(resource)
		if _, exists := serviceInfo.Resources[resource]; !exists {
			invalidResources = append(invalidResources, resource)
		}
	}

	if len(invalidResources) > 0 {
		available := make([]string, 0, len(serviceInfo.Resources))
		for name := range serviceInfo.Resources {
			available = append(available, name)
		}
		sort.Strings(available)
		return fmt.Errorf("invalid resources for service %q: %v. Available resources: %v",
			serviceName, invalidResources, available)
	}

	return nil
}

// GetServiceInfo returns information about a service
func GetServiceInfo(serviceName string) (ServiceInfo, error) {
	serviceName = strings.ToLower(serviceName)
	if err := ensureLoaded(); err != nil {
		return ServiceInfo{}, err
	}
	info, exists := registryData[serviceName]
	if !exists {
		return ServiceInfo{}, fmt.Errorf("service %q not found", serviceName)
	}
	return info, nil
}

// GetServiceType returns the OpenStack service type for a service name
func GetServiceType(serviceName string) (string, error) {
	info, err := GetServiceInfo(serviceName)
	if err != nil {
		return "", err
	}
	return info.ServiceType, nil
}

// GetDisplayName returns the display name for a service
func GetDisplayName(serviceName string) (string, error) {
	info, err := GetServiceInfo(serviceName)
	if err != nil {
		return "", err
	}
	return info.DisplayName, nil
}

// ListServices returns all available OpenStack services
func ListServices() []string {
	if err := ensureLoaded(); err != nil {
		return nil
	}
	services := make([]string, 0, len(registryData))
	for name := range registryData {
		services = append(services, name)
	}
	sort.Strings(services)
	return services
}

// ListResources returns all available resources for a service
func ListResources(serviceName string) ([]string, error) {
	serviceName = strings.ToLower(serviceName)
	if err := ensureLoaded(); err != nil {
		return nil, err
	}
	serviceInfo, exists := registryData[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %q not found", serviceName)
	}

	resources := make([]string, 0, len(serviceInfo.Resources))
	for name := range serviceInfo.Resources {
		resources = append(resources, name)
	}
	sort.Strings(resources)
	return resources, nil
}

func ensureLoaded() error {
	registryOnce.Do(func() {
		registryData, registryErr = loadRegistry()
	})
	return registryErr
}

func loadRegistry() (map[string]ServiceInfo, error) {
	dir, err := findRegistryDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading registry dir: %w", err)
	}

	registry := make(map[string]ServiceInfo)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading registry file %q: %w", name, err)
		}

		var meta ServiceMetadata
		if err := yaml.Unmarshal(content, &meta); err != nil {
			return nil, fmt.Errorf("parsing registry file %q: %w", name, err)
		}

		if meta.Name == "" {
			meta.Name = strings.TrimSuffix(name, filepath.Ext(name))
		}
		meta.Name = strings.ToLower(strings.TrimSpace(meta.Name))
		if meta.Name == "" {
			return nil, fmt.Errorf("registry file %q missing service name", name)
		}
		if meta.DisplayName == "" || meta.ServiceType == "" {
			return nil, fmt.Errorf("registry file %q missing display_name or service_type", name)
		}
		if len(meta.Resources) == 0 {
			return nil, fmt.Errorf("registry file %q has no resources", name)
		}

		for resName, resInfo := range meta.Resources {
			if resName == "" {
				return nil, fmt.Errorf("registry file %q has empty resource name", name)
			}
			if resInfo.Description == "" {
				return nil, fmt.Errorf("registry file %q resource %q missing description", name, resName)
			}
			if len(resInfo.Actions) == 0 && len(meta.Defaults.Actions) > 0 {
				resInfo.Actions = append([]string{}, meta.Defaults.Actions...)
			}
			if len(resInfo.Checks) == 0 && len(meta.Defaults.Checks) > 0 {
				resInfo.Checks = append([]string{}, meta.Defaults.Checks...)
			}
			meta.Resources[resName] = resInfo
		}

		registry[meta.Name] = ServiceInfo{
			ServiceType: meta.ServiceType,
			DisplayName: meta.DisplayName,
			Defaults:    meta.Defaults,
			Resources:   meta.Resources,
		}
	}

	if len(registry) == 0 {
		return nil, fmt.Errorf("no registry files found in %s", dir)
	}
	return registry, nil
}

func findRegistryDir() (string, error) {
	if env := strings.TrimSpace(os.Getenv("OSPA_SCAFFOLD_REGISTRY_PATH")); env != "" {
		if info, err := os.Stat(env); err == nil && info.IsDir() {
			return env, nil
		}
		return "", fmt.Errorf("registry path %q is not a directory", env)
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	dir := wd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "cmd", "scaffold", "internal", "registry", "config")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("unable to locate scaffold registry directory (set OSPA_SCAFFOLD_REGISTRY_PATH)")
}
