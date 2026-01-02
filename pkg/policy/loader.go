package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Load reads and parses a policy YAML file
func Load(path string) (*Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}

	// First, unmarshal into a map to handle the service-keyed structure
	var raw map[string]interface{}
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("parse policy yaml: %w", err)
	}

	// Now unmarshal into the Policy struct
	var p Policy
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("parse policy structure: %w", err)
	}

	// Handle the service-keyed structure in policies
	// The YAML has policies as an array of maps where each map has one service key
	if policiesRaw, ok := raw["policies"].([]interface{}); ok {
		var servicePolicies []ServicePolicy
		for _, policyRaw := range policiesRaw {
			if policyMap, ok := policyRaw.(map[interface{}]interface{}); ok {
				for serviceName, rulesRaw := range policyMap {
					serviceStr := fmt.Sprintf("%v", serviceName)
					if rulesList, ok := rulesRaw.([]interface{}); ok {
						var rules []Rule
						for _, ruleRaw := range rulesList {
							ruleBytes, err := yaml.Marshal(ruleRaw)
							if err != nil {
								continue
							}
							var rule Rule
							if err := yaml.Unmarshal(ruleBytes, &rule); err != nil {
								continue
							}
							// Set service from parent if not set in rule
							if rule.Service == "" {
								rule.Service = serviceStr
							}
							rules = append(rules, rule)
						}
						servicePolicies = append(servicePolicies, ServicePolicy{
							Service: serviceStr,
							Rules:   rules,
						})
					}
				}
			}
		}
		if len(servicePolicies) > 0 {
			p.Policies = servicePolicies
		}
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return &p, nil
}

