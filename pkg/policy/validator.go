package policy

import (
	"fmt"
	"strings"
	
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/catalog"
)

// Validate validates the policy structure and rules
func (p *Policy) Validate() error {
	if p.Version == "" {
		return fmt.Errorf("policy.version is required")
	}

	if len(p.Policies) == 0 {
		return fmt.Errorf("policy.policies must contain at least one service policy")
	}

	seenRuleNames := make(map[string]struct{})
	
	// Dynamically discover supported services and resources from the registry
	// This allows new services to be added without modifying the validator
	supportedResources := catalog.GetSupportedResources()
	supportedServices := make(map[string]bool)
	for serviceName := range supportedResources {
		supportedServices[serviceName] = true
	}
	
	// If no services are registered yet, provide a helpful error message
	if len(supportedServices) == 0 {
		return fmt.Errorf("no services are registered - ensure service packages are imported")
	}

	supportedActions := map[string]bool{
		"log":    true,
		"delete": true,
		"tag":    true,
	}

	for i, sp := range p.Policies {
		service := strings.ToLower(sp.Service)
		if !supportedServices[service] {
			// List available services for better error message
			availableServices := make([]string, 0, len(supportedServices))
			for svc := range supportedServices {
				availableServices = append(availableServices, svc)
			}
			return fmt.Errorf("policies[%d]: unsupported service %q (available services: %v)", i, sp.Service, availableServices)
		}

		if len(sp.Rules) == 0 {
			return fmt.Errorf("policies[%d].%s: must contain at least one rule", i, sp.Service)
		}

		for j, rule := range sp.Rules {
			ruleName := rule.Name
			if ruleName == "" {
				return fmt.Errorf("policies[%d].%s.rules[%d]: name is required", i, sp.Service, j)
			}

			// Check for duplicate rule names
			if _, ok := seenRuleNames[ruleName]; ok {
				return fmt.Errorf("duplicate rule name %q", ruleName)
			}
			seenRuleNames[ruleName] = struct{}{}

			// Validate service matches parent
			if rule.Service != "" && strings.ToLower(rule.Service) != service {
				return fmt.Errorf("rule %q: service %q does not match parent service %q", ruleName, rule.Service, sp.Service)
			}

			// Validate resource
			resource := strings.ToLower(rule.Resource)
			if resource == "" {
				return fmt.Errorf("rule %q: resource is required", ruleName)
			}
			if !supportedResources[service][resource] {
				return fmt.Errorf("rule %q: unsupported resource %q for service %q", ruleName, rule.Resource, sp.Service)
			}

			// Validate action
			action := strings.ToLower(rule.Action)
			if action == "" {
				return fmt.Errorf("rule %q: action is required", ruleName)
			}
			if !supportedActions[action] {
				return fmt.Errorf("rule %q: unsupported action %q (supported: log, delete, tag)", ruleName, rule.Action)
			}

			// Validate action-specific fields
			if action == "tag" {
				if rule.TagName == "" {
					return fmt.Errorf("rule %q: tag_name is required when action is 'tag'", ruleName)
				}
			}

			// Validate check conditions using service-specific validator
			if err := validateCheckConditions(service, &rule.Check, resource, ruleName); err != nil {
				return err
			}

			// Validate age_gt format if present
			if rule.Check.AgeGT != "" {
				if _, err := rule.Check.ParseAgeGT(); err != nil {
					return fmt.Errorf("rule %q: %w", ruleName, err)
				}
			}
		}
	}

	return nil
}

// validateCheckConditions validates check conditions using service-specific validators
func validateCheckConditions(serviceName string, check *CheckConditions, resource, ruleName string) error {
	// Try to get service-specific validator
	if validator, ok := GetValidator(serviceName); ok {
		return validator.ValidateResource(check, resource, ruleName)
	}
	
	// Fallback: if no validator is registered for this service, skip resource-specific validation
	// This allows services without validators to still work (though not recommended)
	return nil
}

