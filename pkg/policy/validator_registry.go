package policy

import (
	"fmt"
	"strings"
	"sync"
)

// ResourceValidator validates check conditions for a specific OpenStack service.
// Kept in pkg/policy to avoid import cycles (audit -> policy, and validation implementations -> policy).
type ResourceValidator interface {
	ServiceName() string
	ValidateResource(check *CheckConditions, resourceType, ruleName string) error
}

var (
	validatorMu sync.RWMutex
	validators  = make(map[string]ResourceValidator)
)

// RegisterValidator registers a service-specific validator.
func RegisterValidator(v ResourceValidator) {
	validatorMu.Lock()
	defer validatorMu.Unlock()
	validators[strings.ToLower(v.ServiceName())] = v
}

// GetValidator returns a validator for serviceName (if any).
func GetValidator(serviceName string) (ResourceValidator, bool) {
	validatorMu.RLock()
	defer validatorMu.RUnlock()
	v, ok := validators[strings.ToLower(serviceName)]
	return v, ok
}

// MustGetValidator returns the validator or an error (useful for tests).
func MustGetValidator(serviceName string) (ResourceValidator, error) {
	v, ok := GetValidator(serviceName)
	if !ok {
		return nil, fmt.Errorf("no validator registered for service %q", serviceName)
	}
	return v, nil
}


