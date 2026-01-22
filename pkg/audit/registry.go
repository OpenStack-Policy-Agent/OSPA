package audit

import (
	"fmt"
	"sync"
)

var (
	auditorRegistry     = make(map[string]map[string]Auditor)
	auditorRegistryLock sync.RWMutex
)

// Register registers an auditor for a service and resource type
func Register(service, resourceType string, auditor Auditor) {
	auditorRegistryLock.Lock()
	defer auditorRegistryLock.Unlock()

	if auditorRegistry[service] == nil {
		auditorRegistry[service] = make(map[string]Auditor)
	}

	auditorRegistry[service][resourceType] = auditor
}

// Get retrieves an auditor for a service and resource type
func Get(service, resourceType string) (Auditor, error) {
	auditorRegistryLock.RLock()
	defer auditorRegistryLock.RUnlock()

	serviceAuditors, exists := auditorRegistry[service]
	if !exists {
		return nil, fmt.Errorf("no auditors registered for service %q", service)
	}

	auditor, exists := serviceAuditors[resourceType]
	if !exists {
		return nil, fmt.Errorf("no auditor registered for service %q, resource type %q", service, resourceType)
	}

	return auditor, nil
}
