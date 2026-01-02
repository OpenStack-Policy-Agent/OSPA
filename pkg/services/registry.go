package services

import (
	"fmt"
	"sync"
)

var (
	registry     = make(map[string]Service)
	registryLock sync.RWMutex
)

// Register registers a service implementation.
// Services are automatically registered via init() functions in their respective files.
// This should typically be called using MustRegister() from an init() function.
func Register(service Service) error {
	registryLock.Lock()
	defer registryLock.Unlock()

	name := service.Name()
	if _, exists := registry[name]; exists {
		return fmt.Errorf("service %q is already registered", name)
	}

	registry[name] = service
	return nil
}

// Get retrieves a service by name.
// Returns an error if the service is not registered.
func Get(name string) (Service, error) {
	registryLock.RLock()
	defer registryLock.RUnlock()

	service, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("service %q is not registered", name)
	}

	return service, nil
}

// List returns all registered service names.
// Useful for debugging and validation.
func List() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// GetSupportedResources returns a map of service names to their supported resource types.
// This uses the resource registry which is populated when services register their resources.
func GetSupportedResources() map[string]map[string]bool {
	return globalResourceRegistry.GetAllResources()
}

// MustRegister registers a service and panics on error.
// This is intended for use in init() functions where registration failures should be fatal.
func MustRegister(service Service) {
	if err := Register(service); err != nil {
		panic(err)
	}
}
