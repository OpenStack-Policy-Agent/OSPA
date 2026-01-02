package generators

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// UpdateServiceFile adds new resources to an existing service file
func UpdateServiceFile(baseDir, serviceName, displayName string, newResources []string) error {
	filePath := filepath.Join(baseDir, "pkg", "services", "services", serviceName+".go")
	
	if !fileExists(filePath) {
		return fmt.Errorf("service file %s does not exist", filePath)
	}

	// Read existing file
	src, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading service file: %w", err)
	}

	// Extract existing resources from the file
	existingResources, err := extractResourcesFromServiceFile(filePath)
	if err != nil {
		return fmt.Errorf("extracting existing resources: %w", err)
	}
	
	// Track existing resources in a set
	existingSet := make(map[string]bool)
	for _, r := range existingResources {
		existingSet[r] = true
	}
	
	// Find resources to add (not already present)
	resourcesToAdd := []string{}
	for _, res := range newResources {
		if !existingSet[res] {
			resourcesToAdd = append(resourcesToAdd, res)
		}
	}

	if len(resourcesToAdd) == 0 {
		return nil // Nothing to add
	}

	// Convert source to string and modify
	content := string(src)
	
	// Add RegisterResource calls in init()
	content = addRegisterResourceCalls(content, serviceName, resourcesToAdd)
	
	// Add cases to GetResourceAuditor switch
	content = addAuditorCases(content, serviceName, displayName, resourcesToAdd)
	
	// Add cases to GetResourceDiscoverer switch
	content = addDiscovererCases(content, displayName, resourcesToAdd)
	
	// Update supported resources comment
	content = updateSupportedResourcesComment(content, resourcesToAdd)

	// Write updated file
	return os.WriteFile(filePath, []byte(content), 0644)
}

// addRegisterResourceCalls adds RegisterResource calls to init() function
func addRegisterResourceCalls(content, serviceName string, resources []string) string {
	// Find the init() function and add RegisterResource calls
	initPattern := `RegisterResource("` + serviceName + `", "`
	
	// Find last RegisterResource call
	lastRegisterIndex := strings.LastIndex(content, initPattern)
	if lastRegisterIndex == -1 {
		// No existing RegisterResource, find init() and add after MustRegister
		mustRegisterIndex := strings.Index(content, "MustRegister(&")
		if mustRegisterIndex != -1 {
			// Find end of MustRegister line
			lineEnd := strings.Index(content[mustRegisterIndex:], "\n")
			if lineEnd != -1 {
				insertPos := mustRegisterIndex + lineEnd + 1
				newCalls := "\t// Register all supported resources for automatic validation\n"
				for _, res := range resources {
					newCalls += fmt.Sprintf("\tRegisterResource(%q, %q)\n", serviceName, res)
				}
				return content[:insertPos] + newCalls + content[insertPos:]
			}
		}
		return content
	}

	// Find the end of the last RegisterResource line
	lineEnd := strings.Index(content[lastRegisterIndex:], "\n")
	if lineEnd == -1 {
		return content
	}

	insertPos := lastRegisterIndex + lineEnd
	newCalls := ""
	for _, res := range resources {
		newCalls += fmt.Sprintf("\tRegisterResource(%q, %q)\n", serviceName, res)
	}

	return content[:insertPos] + newCalls + content[insertPos:]
}

// addAuditorCases adds cases to GetResourceAuditor switch statement
func addAuditorCases(content, serviceName, displayName string, resources []string) string {
	// Find GetResourceAuditor function
	funcStart := strings.Index(content, "func (s *"+displayName+"Service) GetResourceAuditor")
	if funcStart == -1 {
		return content
	}

	// Find the switch statement
	switchStart := strings.Index(content[funcStart:], "switch resourceType {")
	if switchStart == -1 {
		return content
	}

	switchPos := funcStart + switchStart
	
	// Find the default case
	defaultCase := strings.Index(content[switchPos:], "\n\tdefault:")
	if defaultCase == -1 {
		// No default case, find closing brace
		closingBrace := strings.LastIndex(content[switchPos:], "\t}")
		if closingBrace == -1 {
			return content
		}
		insertPos := switchPos + closingBrace
		newCases := "\n"
		for _, res := range resources {
			titleRes := strings.Title(res)
			newCases += fmt.Sprintf("\tcase %q:\n\t\treturn &%s.%sAuditor{}, nil\n", res, serviceName, titleRes)
		}
		return content[:insertPos] + newCases + content[insertPos:]
	}

	// Insert before default case
	insertPos := switchPos + defaultCase
	newCases := "\n"
	for _, res := range resources {
		titleRes := strings.Title(res)
		newCases += fmt.Sprintf("\tcase %q:\n\t\treturn &%s.%sAuditor{}, nil\n", res, serviceName, titleRes)
	}

	return content[:insertPos] + newCases + content[insertPos:]
}

// addDiscovererCases adds cases to GetResourceDiscoverer switch statement
func addDiscovererCases(content, displayName string, resources []string) string {
	// Find GetResourceDiscoverer function
	funcStart := strings.Index(content, "func (s *"+displayName+"Service) GetResourceDiscoverer")
	if funcStart == -1 {
		return content
	}

	// Find the switch statement
	switchStart := strings.Index(content[funcStart:], "switch resourceType {")
	if switchStart == -1 {
		return content
	}

	switchPos := funcStart + switchStart
	
	// Find the default case
	defaultCase := strings.Index(content[switchPos:], "\n\tdefault:")
	if defaultCase == -1 {
		// No default case, find closing brace
		closingBrace := strings.LastIndex(content[switchPos:], "\t}")
		if closingBrace == -1 {
			return content
		}
		insertPos := switchPos + closingBrace
		newCases := "\n"
		for _, res := range resources {
			titleRes := strings.Title(res)
			newCases += fmt.Sprintf("\tcase %q:\n\t\treturn &discovery.%s%sDiscoverer{}, nil\n", res, displayName, titleRes)
		}
		return content[:insertPos] + newCases + content[insertPos:]
	}

	// Insert before default case
	insertPos := switchPos + defaultCase
	newCases := "\n"
	for _, res := range resources {
		titleRes := strings.Title(res)
		newCases += fmt.Sprintf("\tcase %q:\n\t\treturn &discovery.%s%sDiscoverer{}, nil\n", res, displayName, titleRes)
	}

	return content[:insertPos] + newCases + content[insertPos:]
}

// updateSupportedResourcesComment updates the supported resources comment
func updateSupportedResourcesComment(content string, resources []string) string {
	// Find the "Supported resources:" comment
	commentStart := strings.Index(content, "// Supported resources:")
	if commentStart == -1 {
		return content
	}

	// Find the end of the comment block (next non-comment line)
	lines := strings.Split(content[commentStart:], "\n")
	commentEnd := 0
	for i, line := range lines {
		if i > 0 && !strings.HasPrefix(strings.TrimSpace(line), "//") && strings.TrimSpace(line) != "" {
			commentEnd = i
			break
		}
	}

	if commentEnd == 0 {
		return content
	}

	// Build new comment lines
	newCommentLines := []string{"// Supported resources:"}
	for _, res := range resources {
		newCommentLines = append(newCommentLines, fmt.Sprintf("//   - %s: Resource of type %s", res, res))
	}

	// Reconstruct content
	beforeComment := content[:commentStart]
	afterComment := ""
	for i := commentEnd; i < len(lines); i++ {
		if i == commentEnd {
			afterComment = lines[i]
		} else {
			afterComment += "\n" + lines[i]
		}
	}

	return beforeComment + strings.Join(newCommentLines, "\n") + "\n" + afterComment
}

// formatGoFile formats a Go source file
func formatGoFile(filePath string) error {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	formatted, err := format.Source(src)
	if err != nil {
		// If formatting fails, don't fail the whole operation
		return nil
	}

	return os.WriteFile(filePath, formatted, 0644)
}

