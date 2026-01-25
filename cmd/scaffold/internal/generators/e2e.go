package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateE2ETest generates the e2e test files for a service.
// Creates:
//   - e2e/<service>/resource_creator.go - Helper to create test resources
//   - e2e/<service>/<resource>_test.go  - Test file for each resource
func GenerateE2ETest(baseDir, serviceName, displayName string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateE2ETestWithSpecs(baseDir, serviceName, displayName, specs)
}

func generateE2ETestWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec) error {
	// Create service directory under e2e/
	e2eServiceDir := filepath.Join(baseDir, "e2e", serviceName)
	if err := os.MkdirAll(e2eServiceDir, 0755); err != nil {
		return err
	}

	// Generate resource_creator.go
	if err := generateResourceCreator(e2eServiceDir, serviceName, displayName, resources); err != nil {
		return err
	}

	// Generate individual <resource>_test.go files
	for _, resource := range resources {
		if err := generateResourceTestFile(e2eServiceDir, serviceName, displayName, resource); err != nil {
			return err
		}
	}

	// Add Get<Service>Client method to e2e/engine.go
	if err := appendEngineClientMethod(baseDir, serviceName, displayName); err != nil {
		return err
	}

	return nil
}

// generateResourceCreator creates the resource_creator.go file with instructions
func generateResourceCreator(dir, serviceName, displayName string, resources []ResourceSpec) error {
	filePath := filepath.Join(dir, "resource_creator.go")

	// Check if file exists and filter out resources that already have implementations
	if existingContent, err := os.ReadFile(filePath); err == nil {
		newResources := filterUnimplementedCreators(string(existingContent), resources)
		if len(newResources) == 0 {
			fmt.Printf("Info: All requested resources already have creators in resource_creator.go\n")
			return nil
		}
		// Append new creators to existing file
		return appendResourceCreators(filePath, string(existingContent), serviceName, displayName, newResources)
	}

	tmpl := `//go:build e2e

// Package {{.ServiceName}} contains e2e tests for the {{.DisplayName}} service.
//
// =============================================================================
// RESOURCE CREATOR - READ THIS FIRST
// =============================================================================
//
// This file provides helper functions to create test resources for e2e tests.
// Each resource may have dependencies that must be created first.
//
// HOW TO USE:
// 1. Implement the Create<Resource>() functions below
// 2. Each function should create the resource AND its dependencies
// 3. Return a cleanup function that deletes resources in reverse order
// 4. Use these functions in the corresponding <resource>_test.go files
//
// DEPENDENCY GRAPH FOR {{.DisplayName}}:
// =============================================================================
{{range .Resources}}
// {{.Name | Pascal}}:
//   Description: {{.Description}}
//   Gophercloud: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
//   OpenStack API: https://docs.openstack.org/api-ref/{{$.ServiceName}}
{{end}}
// =============================================================================

package {{.ServiceName}}

import (
	"fmt"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	// TODO: Import the specific gophercloud packages you need:
	// "github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	// "github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

const testPrefix = "ospa-e2e-"

// =============================================================================
// RESOURCE CREATORS - IMPLEMENT THESE
// =============================================================================
{{range .Resources}}

// Create{{.Name | Pascal}} creates a test {{.Name}} and returns:
//   - resourceID: The ID of the created resource (for filtering audit results)
//   - cleanup: A function to delete the resource and its dependencies
func Create{{.Name | Pascal}}(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
	t.Helper()
	
	// TODO: Implement resource creation
	// See the example above and the gophercloud documentation
	
	t.Skip("Create{{.Name | Pascal}} not implemented - implement in resource_creator.go")
	return "", func() {}
}
{{end}}

// =============================================================================
// CLEANUP HELPER
// =============================================================================

// CleanupOrphans deletes any leaked test resources (those with testPrefix).
// Run this manually if tests fail and leave resources behind:
//   go test -tags=e2e ./e2e/{{.ServiceName}}/... -run TestCleanupOrphans
func CleanupOrphans(t *testing.T, client *gophercloud.ServiceClient) {
	t.Helper()
	
	// TODO: Implement cleanup for orphaned resources
	// List all resources, filter by testPrefix, delete them
	
	t.Log("TODO: Implement orphan cleanup")
}
`

	data := struct {
		ServiceName string
		DisplayName string
		Resources   []ResourceSpec
	}{
		ServiceName: serviceName,
		DisplayName: displayName,
		Resources:   resources,
	}

	funcMap := template.FuncMap{
		"Pascal":  ToPascal,
		"ToUpper": func(s string) string { return s },
	}

	t, err := template.New("resource_creator").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(filePath, t, data)
}

// generateResourceTestFile creates an individual <resource>_test.go file
func generateResourceTestFile(dir, serviceName, displayName string, resource ResourceSpec) error {
	filePath := filepath.Join(dir, resource.Name+"_test.go")

	// Check if test file already exists and has real tests
	if hasE2ETestImplementation(filePath, resource.Name) {
		fmt.Printf("Info: %s_test.go already has e2e tests, skipping\n", resource.Name)
		return nil
	}

	tmpl := `//go:build e2e

package {{.ServiceName}}

// =============================================================================
// {{.Resource.Name | Pascal}} E2E TESTS
// =============================================================================
//
// BEFORE WRITING TESTS:
// 1. Implement Create{{.Resource.Name | Pascal}}() in resource_creator.go
// 2. The creator should handle all dependencies (network, subnet, etc.)
// 3. The creator returns a cleanup function - always defer it!
//
// TEST COVERAGE CHECKLIST:
// - [ ] Status check (status: ACTIVE, DOWN, ERROR, etc.)
// - [ ] Age check (age_gt: 30d)
// - [ ] Unused check (unused: true) - if applicable
// - [ ] Exempt names (exempt_names: [...])
// - [ ] Multiple resources (batch discovery)
// - [ ] Error handling (invalid resource, missing permissions)
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/{{.ServiceName}}/... -v -run {{.Resource.Name | Pascal}}
//
// =============================================================================

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_StatusCheck verifies status-based auditing.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	// Create test resource using the helper from resource_creator.go
	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	// Run audit with status check
	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-status
      description: Find {{.Resource.Name}} by status
      service: {{.ServiceName}}
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter to our specific resource
	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// Verify the resource was discovered
	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors: %d", resourceResults.Errors)
	}
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_UnusedCheck verifies unused detection.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	// Create an "unused" resource (no attachments/dependencies)
	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-unused
      description: Find unused {{.Resource.Name}}
      service: {{.ServiceName}}
      resource: {{.Resource.Name}}
      check:
        unused: true
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	// TODO: Add assertions based on whether the resource should be flagged
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_ExemptNames verifies name exemptions work.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	// The resource name starts with "ospa-e2e-" - exempt it
	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-exempt
      description: Test exemption by name prefix
      service: {{.ServiceName}}
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
        exempt_names:
          - ospa-e2e-*
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	// Resource should be compliant (exempted by name)
	if resourceResults.Violations > 0 {
		t.Error("Expected resource to be exempt by name")
	}
}

// TestCleanup_{{.Resource.Name | Pascal}} cleans up orphaned test resources.
// Run manually: go test -tags=e2e ./e2e/{{.ServiceName}}/... -run TestCleanup
func TestCleanup_{{.Resource.Name | Pascal}}(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup in short mode")
	}

	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	CleanupOrphans(t, client)
}
`

	data := struct {
		ServiceName  string
		DisplayName  string
		ClientMethod string
		Resource     ResourceSpec
	}{
		ServiceName:  serviceName,
		DisplayName:  displayName,
		ClientMethod: getClientMethodName(serviceName),
		Resource:     resource,
	}

	funcMap := template.FuncMap{
		"Pascal": ToPascal,
	}

	t, err := template.New("resource_test").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(filePath, t, data)
}

// getClientMethodName returns the method name suffix for getting the service client.
func getClientMethodName(serviceName string) string {
	switch serviceName {
	case "nova":
		return "Compute"
	case "neutron":
		return "Network"
	case "cinder":
		return "BlockStorage"
	case "glance":
		return "Image"
	case "keystone":
		return "Identity"
	case "heat":
		return "Orchestration"
	case "swift":
		return "ObjectStorage"
	case "octavia":
		return "LoadBalancer"
	default:
		return ToPascal(serviceName)
	}
}

// appendEngineClientMethod adds a Get<Service>Client method to e2e/engine.go
func appendEngineClientMethod(baseDir, serviceName, displayName string) error {
	enginePath := filepath.Join(baseDir, "e2e", "engine.go")

	// Read existing content
	content, err := os.ReadFile(enginePath)
	if err != nil {
		// engine.go doesn't exist, skip
		return nil
	}

	clientMethodName := getClientMethodName(serviceName)
	methodSignature := "func (e *TestEngine) Get" + clientMethodName + "Client"

	// Check if method already exists
	if contains(string(content), methodSignature) {
		return nil
	}

	// Get the auth method name for this service
	authMethodName := getAuthMethodName(serviceName, displayName)

	// Generate the new method
	newMethod := `
// Get` + clientMethodName + `Client returns a gophercloud client for the ` + displayName + ` service.
func (e *TestEngine) Get` + clientMethodName + `Client(t *testing.T) *gophercloud.ServiceClient {
	t.Helper()
	client, err := e.Session.` + authMethodName + `()
	if err != nil {
		t.Fatalf("Failed to get ` + serviceName + ` client: %v", err)
	}
	return client
}
`

	// Find a good insertion point - after the last Get*Client method or before LoadPolicy
	contentStr := string(content)

	// Try to find "// LoadPolicy" as insertion point
	insertPoint := -1
	loadPolicyIdx := indexOf(contentStr, "\n// LoadPolicy")
	if loadPolicyIdx != -1 {
		insertPoint = loadPolicyIdx
	}

	if insertPoint == -1 {
		// Fallback: append before the last closing newlines
		insertPoint = len(contentStr)
		// Trim trailing whitespace and add back
		contentStr = trimRight(contentStr, "\n\t ") + "\n"
		insertPoint = len(contentStr)
	}

	// Insert the new method
	newContent := contentStr[:insertPoint] + newMethod + contentStr[insertPoint:]

	return os.WriteFile(enginePath, []byte(newContent), 0644)
}

// getAuthMethodName returns the Session method name for getting the service client.
func getAuthMethodName(serviceName, displayName string) string {
	switch serviceName {
	case "nova":
		return "GetNovaClient"
	case "neutron":
		return "GetNeutronClient"
	case "cinder":
		return "GetCinderClient"
	case "glance":
		return "GetGlanceClient"
	case "keystone":
		return "GetKeystoneClient"
	case "heat":
		return "GetHeatClient"
	case "swift":
		return "GetSwiftClient"
	case "octavia":
		return "GetOctaviaClient"
	default:
		return "Get" + displayName + "Client"
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return indexOf(s, substr) != -1
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// trimRight trims characters from the right of s
func trimRight(s, cutset string) string {
	for len(s) > 0 {
		found := false
		for _, c := range cutset {
			if rune(s[len(s)-1]) == c {
				s = s[:len(s)-1]
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return s
}

// hasE2ETestImplementation checks if an e2e test file already exists.
// Returns true if the file exists (regardless of whether it's a placeholder or real).
func hasE2ETestImplementation(filePath, resourceName string) bool {
	_, err := os.Stat(filePath)
	return err == nil // File exists
}

// filterUnimplementedCreators returns only resources that don't have real Create<Resource> implementations.
func filterUnimplementedCreators(content string, resources []ResourceSpec) []ResourceSpec {
	var unimplemented []ResourceSpec
	for _, r := range resources {
		if !hasCreatorImplementation(content, r.Name) {
			unimplemented = append(unimplemented, r)
		} else {
			fmt.Printf("Info: Create%s already has an implementation, skipping\n", ToPascal(r.Name))
		}
	}
	return unimplemented
}

// hasCreatorImplementation checks if a Create<Resource> function already exists.
// Returns true if the function exists (regardless of whether it's a placeholder or real).
func hasCreatorImplementation(content, resourceName string) bool {
	funcName := "Create" + ToPascal(resourceName)
	funcSignature := "func " + funcName + "("
	return strings.Contains(content, funcSignature)
}

// appendResourceCreators appends new Create<Resource> functions to an existing resource_creator.go.
func appendResourceCreators(filePath, existingContent, serviceName, displayName string, resources []ResourceSpec) error {
	tmpl := `
{{range .Resources}}
// Create{{.Name | Pascal}} creates a test {{.Name}} and returns:
//   - resourceID: The ID of the created resource (for filtering audit results)
//   - cleanup: A function to delete the resource and its dependencies
func Create{{.Name | Pascal}}(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
	t.Helper()
	
	// TODO: Implement resource creation
	// See the example above and the gophercloud documentation
	
	t.Skip("Create{{.Name | Pascal}} not implemented - implement in resource_creator.go")
	return "", func() {}
}
{{end}}`

	data := struct {
		ServiceName string
		DisplayName string
		Resources   []ResourceSpec
	}{
		ServiceName: serviceName,
		DisplayName: displayName,
		Resources:   resources,
	}

	funcMap := template.FuncMap{
		"Pascal": ToPascal,
	}

	t, err := template.New("creator_append").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	var newContent strings.Builder
	if err := t.Execute(&newContent, data); err != nil {
		return err
	}

	// Find a good insertion point - before CleanupOrphans or at the end
	insertPoint := strings.Index(existingContent, "// CleanupOrphans")
	if insertPoint == -1 {
		insertPoint = strings.Index(existingContent, "func CleanupOrphans")
	}

	var finalContent string
	if insertPoint != -1 {
		// Insert before CleanupOrphans
		finalContent = existingContent[:insertPoint] + newContent.String() + "\n" + existingContent[insertPoint:]
	} else {
		// Append at end
		finalContent = strings.TrimRight(existingContent, "\n\t ") + "\n" + newContent.String()
	}

	return os.WriteFile(filePath, []byte(finalContent), 0644)
}

