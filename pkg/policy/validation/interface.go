package validation

import (
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// Validator defines the interface for service-specific policy validation
type Validator interface {
	// ServiceName returns the OpenStack service name this validator handles (e.g., "nova", "neutron")
	ServiceName() string

	// ValidateResource validates check conditions for a specific resource type
	// Returns an error if the check conditions are invalid for the resource
	ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error
}

