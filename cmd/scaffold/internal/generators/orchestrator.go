package generators

import (
	"fmt"
)

// GenerateService orchestrates the generation of all files for a service
func GenerateService(serviceName, displayName, serviceType string, resources []string, force bool) error {
	baseDir := "."
	
	// Analyze existing service to determine what needs to be generated
	analysis, err := AnalyzeService(baseDir, serviceName, resources)
	if err != nil && !force {
		// If analysis fails and we're not forcing, continue anyway
		analysis = &ServiceAnalysis{}
	}
	
	// Determine which resources are new
	newResources := resources
	if analysis != nil && len(analysis.ExistingResources) > 0 {
		existingSet := make(map[string]bool)
		for _, r := range analysis.ExistingResources {
			existingSet[r] = true
		}
		newResources = []string{}
		for _, r := range resources {
			if !existingSet[r] {
				newResources = append(newResources, r)
			}
		}
	}
	
	// Generate service file (will update if exists and not forcing)
	if err := GenerateServiceFile(baseDir, serviceName, displayName, serviceType, resources, force); err != nil {
		return fmt.Errorf("generating service file: %w", err)
	}

	// Generate discovery file
	if err := GenerateDiscoveryFile(baseDir, serviceName, displayName, resources, force); err != nil {
		return fmt.Errorf("generating discovery file: %w", err)
	}

	// Generate auditor files (only for new resources if not forcing)
	resourcesToGenerate := resources
	if !force && len(newResources) > 0 {
		resourcesToGenerate = newResources
	}
	if err := GenerateAuditorFiles(baseDir, serviceName, displayName, resourcesToGenerate, force); err != nil {
		return fmt.Errorf("generating auditor files: %w", err)
	}

	// Generate auth client method
	if err := GenerateAuthMethod(baseDir, serviceName, displayName, serviceType, force); err != nil {
		return fmt.Errorf("generating auth method: %w", err)
	}

	// Generate unit test files (only for new resources if not forcing)
	if err := GenerateUnitTests(baseDir, serviceName, displayName, resourcesToGenerate, force); err != nil {
		return fmt.Errorf("generating unit tests: %w", err)
	}

	// Generate e2e test file (update if exists)
	if err := GenerateE2ETest(baseDir, serviceName, displayName, resourcesToGenerate, force); err != nil {
		return fmt.Errorf("generating e2e test: %w", err)
	}

	// Generate policy guide (update if exists)
	if err := GeneratePolicyGuide(baseDir, serviceName, displayName, serviceType, resources, force); err != nil {
		return fmt.Errorf("generating policy guide: %w", err)
	}

	// Generate validation file (update if exists)
	if err := GenerateValidationFile(baseDir, serviceName, displayName, resources, force); err != nil {
		return fmt.Errorf("generating validation file: %w", err)
	}

	return nil
}

