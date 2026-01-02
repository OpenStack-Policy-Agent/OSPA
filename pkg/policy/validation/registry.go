package validation

import (
	"fmt"
	"strings"
	"sync"
)

var (
	validators     = make(map[string]Validator)
	validatorsLock sync.RWMutex
)

// Register registers a service-specific validator
func Register(validator Validator) {
	validatorsLock.Lock()
	defer validatorsLock.Unlock()
	
	serviceName := strings.ToLower(validator.ServiceName())
	validators[serviceName] = validator
}

// Get retrieves a validator for a given service name
func Get(serviceName string) (Validator, error) {
	validatorsLock.RLock()
	defer validatorsLock.RUnlock()
	
	serviceName = strings.ToLower(serviceName)
	validator, ok := validators[serviceName]
	if !ok {
		return nil, fmt.Errorf("no validator registered for service %q", serviceName)
	}
	return validator, nil
}

// Has checks if a validator exists for a service
func Has(serviceName string) bool {
	validatorsLock.RLock()
	defer validatorsLock.RUnlock()
	
	serviceName = strings.ToLower(serviceName)
	_, ok := validators[serviceName]
	return ok
}

// List returns all registered service names
func List() []string {
	validatorsLock.RLock()
	defer validatorsLock.RUnlock()
	
	services := make([]string, 0, len(validators))
	for service := range validators {
		services = append(services, service)
	}
	return services
}

