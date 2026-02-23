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
	"testing"

	"github.com/gophercloud/gophercloud"
	// TODO: Import the specific gophercloud packages you need:
	// "github.com/gophercloud/gophercloud/openstack/<service>/<version>/<resource>"
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
// - [x] Status check (status: ACTIVE, DOWN, ERROR, etc.)
// - [x] Age check (age_gt: 30d)
// - [x] Unused check (unused: true) - if applicable
// - [x] Exempt names (exempt_names: [...])
// - [x] Discovery (multiple resources)
// - [x] Classification propagation (severity/category/guide_ref)
// - [x] Output JSON
// - [x] Output CSV{{if HasAction .Resource.Actions "delete"}}
// - [x] Delete action{{end}}{{if HasAction .Resource.Actions "tag"}}
// - [x] Tag action{{end}}{{if or (HasAction .Resource.Actions "delete") (HasAction .Resource.Actions "tag")}}
// - [x] Dry-run remediation skip
// - [x] Allow-actions filtering{{end}}{{range .Resource.RichChecks}}
// - [ ] Domain check: {{.Name}} ({{.Description}}){{end}}
//
// RUNNING TESTS:
//   OS_CLOUD=mycloud go test -tags=e2e ./e2e/{{.ServiceName}}/... -v -run {{.Resource.Name | Pascal}}
//
// =============================================================================

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/e2e"
)

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_StatusCheck verifies status-based auditing.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_StatusCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-status
      description: Find {{.Resource.Name}} by status
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors: %d", resourceResults.Errors)
	}

	// TODO: Adjust expected status for this resource type (e.g. "available", "in-use", "BUILD")
	// TODO: Add a non-compliant status test (create resource, check for a status it does NOT have)
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_AgeGTCheck verifies age-based auditing.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_AgeGTCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-age
      description: Find {{.Resource.Name}} older than 30 days
      resource: {{.Resource.Name}}
      check:
        age_gt: 30d
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	// Freshly created resource should be compliant (younger than 30 days)
	if resourceResults.Violations > 0 {
		t.Error("Freshly created resource should not be flagged by age_gt: 30d")
	}

	// TODO: Add an AgeGTViolation test: create resource, sleep briefly, audit with age_gt: 1s, expect violation
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_UnusedCheck verifies unused detection.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_UnusedCheck(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-unused
      description: Find unused {{.Resource.Name}}
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

	// TODO: Assert based on resource-specific unused semantics
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_ExemptNames verifies name exemptions work.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_ExemptNames(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-exempt
      description: Test exemption by name prefix
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

	if resourceResults.Violations > 0 {
		t.Error("Expected resource to be exempt by name")
	}
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_Discovery verifies batch discovery.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_Discovery(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	id1, cleanup1 := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup1()
	id2, cleanup2 := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup2()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-discovery
      description: Discover {{.Resource.Name}} resources
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	r1 := results.FilterByService("{{.ServiceName}}").FilterByResourceType("{{.Resource.Name}}").FilterByResourceID(id1)
	r2 := results.FilterByService("{{.ServiceName}}").FilterByResourceType("{{.Resource.Name}}").FilterByResourceID(id2)

	if r1.Scanned == 0 {
		t.Error("First resource was not discovered")
	}
	if r2.Scanned == 0 {
		t.Error("Second resource was not discovered")
	}
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_Classification verifies severity/category/guide_ref propagation.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_Classification(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-classify
      description: Test classification fields
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
      action: log
      severity: high
      category: security
      guide_ref: TEST-001` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}
	resourceResults.AssertClassification(t, "high", "security", "TEST-001")

	// TODO: Replace TEST-001 with a real OpenStack Security Guide reference for this resource
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_OutputJSON verifies JSON output contains required fields.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_OutputJSON(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	_, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-json
      description: Output format test
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
      action: log
      severity: medium
      category: compliance` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results, filePath := engine.RunAuditToFile(t, policy, "json")
	defer os.Remove(filePath)

	if results.Scanned == 0 {
		t.Skip("No resources scanned, cannot validate output")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read JSON output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("JSON output file is empty")
	}

	var finding map[string]interface{}
	if err := json.Unmarshal(data[:e2e.FindLineEnd(data)], &finding); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	for _, field := range []string{"rule_id", "resource_id", "compliant", "severity", "category"} {
		if _, ok := finding[field]; !ok {
			t.Errorf("JSON output missing required field: %s", field)
		}
	}
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_OutputCSV verifies CSV output has the correct headers.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_OutputCSV(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	client := engine.Get{{.ClientMethod}}Client(t)

	_, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-csv
      description: Output format test
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
      action: log
      severity: low
      category: operations` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results, filePath := engine.RunAuditToFile(t, policy, "csv")
	defer os.Remove(filePath)

	if results.Scanned == 0 {
		t.Skip("No resources scanned, cannot validate output")
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open CSV output: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}

	expected := map[string]bool{"rule_id": false, "resource_id": false, "compliant": false, "severity": false, "category": false, "guide_ref": false}
	for _, col := range header {
		if _, ok := expected[col]; ok {
			expected[col] = true
		}
	}
	for col, found := range expected {
		if !found {
			t.Errorf("CSV header missing expected column: %s", col)
		}
	}

	row, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV data row: %v", err)
	}
	if len(row) != len(header) {
		t.Errorf("CSV data row has %d columns, expected %d", len(row), len(header))
	}
}
{{if HasAction .Resource.Actions "delete"}}
// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_DeleteAction verifies delete remediation.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_DeleteAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-delete
      description: Delete {{.Resource.Name}} resources
      resource: {{.Resource.Name}}
      check:
        unused: true
      action: delete` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during delete: %d", resourceResults.Errors)
	}

	// TODO: Verify the resource was actually deleted (e.g. GET returns 404)
}
{{end}}{{if HasAction .Resource.Actions "tag"}}
// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_TagAction verifies tag remediation.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_TagAction(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-tag
      description: Tag {{.Resource.Name}} resources
      resource: {{.Resource.Name}}
      check:
        status: ACTIVE
      action: tag
      tag_name: ospa-e2e-tagged` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Errors > 0 {
		t.Errorf("Unexpected errors during tag: %d", resourceResults.Errors)
	}

	// TODO: Verify the tag was actually applied (e.g. GET resource, check tags list)
}
{{end}}{{if or (HasAction .Resource.Actions "delete") (HasAction .Resource.Actions "tag")}}
// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_DryRunSkip verifies dry-run skips remediation.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_DryRunSkip(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = false
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-dryrun
      description: Dry run delete
      resource: {{.Resource.Name}}
      check:
        unused: true
      action: delete` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)
	resourceResults.AssertRemediationSkipped(t, "dry-run")

	// TODO: Verify the resource still exists after dry-run (e.g. GET returns 200)
}

// Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_AllowActionsFiltering verifies the allow-actions filter.
func Test{{.DisplayName}}_{{.Resource.Name | Pascal}}_AllowActionsFiltering(t *testing.T) {
	engine := e2e.NewTestEngine(t)
	engine.Apply = true
	engine.AllowActions = []string{"tag"}
	client := engine.Get{{.ClientMethod}}Client(t)

	resourceID, cleanup := Create{{.Resource.Name | Pascal}}(t, client)
	defer cleanup()

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{.ServiceName}}:
    - name: test-{{.Resource.Name}}-allowlist
      description: Delete not in allowlist
      resource: {{.Resource.Name}}
      check:
        unused: true
      action: delete` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	resourceResults := results.FilterByService("{{.ServiceName}}").
		FilterByResourceType("{{.Resource.Name}}").
		FilterByResourceID(resourceID)

	resourceResults.LogSummary(t)

	if resourceResults.Scanned == 0 {
		t.Error("Expected resource to be scanned")
	}

	resourceResults.AssertRemediationSkipped(t, "action_not_allowed")
}
{{end}}{{range .Resource.RichChecks}}
// Test{{$.DisplayName}}_{{$.Resource.Name | Pascal}}_Check_{{.Name | Pascal}} tests the {{.Name}} domain check.
func Test{{$.DisplayName}}_{{$.Resource.Name | Pascal}}_Check_{{.Name | Pascal}}(t *testing.T) {
	// TODO: Implement domain-specific check test for {{.Name}}
	// Description: {{.Description}}
	// Category: {{.Category}}, Severity: {{.Severity}}
	t.Skip("Domain check test for {{.Name}} not yet implemented")
}
{{end}}
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
		"Pascal":    ToPascal,
		"HasAction": hasAction,
		"HasCheck":  hasCheck,
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

// hasAction checks if a given action is in the actions slice.
func hasAction(actions []string, action string) bool {
	for _, a := range actions {
		if strings.EqualFold(a, action) {
			return true
		}
	}
	return false
}

// hasCheck checks if a given check is in the checks slice.
func hasCheck(checks []string, check string) bool {
	for _, c := range checks {
		if strings.EqualFold(c, check) {
			return true
		}
	}
	return false
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
