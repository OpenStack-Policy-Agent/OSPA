package generators

import (
	"fmt"
)

// GenerateService orchestrates the generation of all files for a service.
// Files are always overwritten if they exist.
func GenerateService(serviceName, displayName, serviceType string, resources []string) error {
	baseDir := "."

	// Generate service file
	if err := GenerateServiceFile(baseDir, serviceName, displayName, serviceType, resources); err != nil {
		return fmt.Errorf("generating service file: %w", err)
	}

	// Generate discovery file
	if err := GenerateDiscoveryFile(baseDir, serviceName, displayName, resources); err != nil {
		return fmt.Errorf("generating discovery file: %w", err)
	}

	// Generate auditor files
	if err := GenerateAuditorFiles(baseDir, serviceName, displayName, resources); err != nil {
		return fmt.Errorf("generating auditor files: %w", err)
	}

	// Generate auth client method
	if err := GenerateAuthMethod(baseDir, serviceName, displayName, serviceType); err != nil {
		return fmt.Errorf("generating auth method: %w", err)
	}

	// Generate unit test files
	if err := GenerateUnitTests(baseDir, serviceName, displayName, resources); err != nil {
		return fmt.Errorf("generating unit tests: %w", err)
	}

	// Generate e2e test file
	if err := GenerateE2ETest(baseDir, serviceName, displayName, resources); err != nil {
		return fmt.Errorf("generating e2e test: %w", err)
	}

	// Generate policy guide
	if err := GeneratePolicyGuide(baseDir, serviceName, displayName, serviceType, resources); err != nil {
		return fmt.Errorf("generating policy guide: %w", err)
	}

	// Generate validation file
	if err := GenerateValidationFile(baseDir, serviceName, displayName, resources); err != nil {
		return fmt.Errorf("generating validation file: %w", err)
	}

	return nil
}
