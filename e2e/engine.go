//go:build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/orchestrator"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services" // Register discoverers
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"           // Register services
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services"  // Register service implementations
)

// TestEngine provides a framework for running e2e tests
type TestEngine struct {
	Session     *auth.Session
	CloudName   string
	PolicyPath  string
	Workers     int
	Apply       bool
	AllTenants  bool
}

// NewTestEngine creates a new test engine with default settings
func NewTestEngine(t *testing.T) *TestEngine {
	t.Helper()

	cloudName := os.Getenv("OS_CLOUD")
	if cloudName == "" {
		t.Skip("OS_CLOUD not set, skipping e2e test")
	}

	session, err := auth.NewSession(cloudName)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	policyPath := os.Getenv("OSPA_E2E_POLICY")
	if policyPath == "" {
		policyPath = "../../examples/policies.yaml"
	}

	return &TestEngine{
		Session:    session,
		CloudName:  cloudName,
		PolicyPath: policyPath,
		Workers:    4,
		Apply:      false,
		AllTenants: false,
	}
}

// LoadPolicy loads a policy from the configured path or a custom path
func (e *TestEngine) LoadPolicy(t *testing.T, customPath ...string) *policy.Policy {
	t.Helper()

	path := e.PolicyPath
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	}

	p, err := policy.Load(path)
	if err != nil {
		t.Fatalf("Failed to load policy from %s: %v", path, err)
	}

	return p
}

// RunAudit runs a policy audit and returns results
func (e *TestEngine) RunAudit(t *testing.T, p *policy.Policy) *AuditResults {
	t.Helper()

	orch := orchestrator.NewOrchestrator(p, e.Session, e.Workers, e.Apply, e.AllTenants)
	defer orch.Stop()

	resultsChan, err := orch.Run()
	if err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	results := &AuditResults{
		Scanned:    0,
		Violations: 0,
		Errors:     0,
		Results:    []*ResultSummary{},
	}

	for result := range resultsChan {
		results.Scanned++
		if result.Error != nil {
			results.Errors++
		}
		if result.RemediationError != nil {
			results.Errors++
		}
		if !result.Compliant {
			results.Violations++
		}

		service := ""
		resourceType := ""
		if result.Rule != nil {
			service = result.Rule.Service
			resourceType = result.Rule.Resource
		}

		results.Results = append(results.Results, &ResultSummary{
			RuleID:       result.RuleID,
			ResourceID:   result.ResourceID,
			ResourceName: result.ResourceName,
			Service:      service,
			ResourceType: resourceType,
			Compliant:    result.Compliant,
			Observation:  result.Observation,
			Error:        result.Error,
		})
	}

	return results
}

// AuditResults contains the results of an audit run
type AuditResults struct {
	Scanned    int
	Violations int
	Errors     int
	Results    []*ResultSummary
}

// ResultSummary summarizes a single audit result
type ResultSummary struct {
	RuleID       string
	ResourceID   string
	ResourceName string
	Service      string
	ResourceType string
	Compliant    bool
	Observation  string
	Error        error
}

// FilterByService filters results by service name
func (r *AuditResults) FilterByService(service string) *AuditResults {
	filtered := &AuditResults{
		Scanned:    0,
		Violations: 0,
		Errors:     0,
		Results:    []*ResultSummary{},
	}

	for _, result := range r.Results {
		if result.Service == service {
			filtered.Results = append(filtered.Results, result)
			filtered.Scanned++
			if !result.Compliant {
				filtered.Violations++
			}
			if result.Error != nil {
				filtered.Errors++
			}
		}
	}

	return filtered
}

// FilterByResourceType filters results by resource type
func (r *AuditResults) FilterByResourceType(resourceType string) *AuditResults {
	filtered := &AuditResults{
		Scanned:    0,
		Violations: 0,
		Errors:     0,
		Results:    []*ResultSummary{},
	}

	for _, result := range r.Results {
		if result.ResourceType == resourceType {
			filtered.Results = append(filtered.Results, result)
			filtered.Scanned++
			if !result.Compliant {
				filtered.Violations++
			}
			if result.Error != nil {
				filtered.Errors++
			}
		}
	}

	return filtered
}

// FilterByResourceID filters results by resource ID
func (r *AuditResults) FilterByResourceID(resourceID string) *AuditResults {
	filtered := &AuditResults{
		Scanned:    0,
		Violations: 0,
		Errors:     0,
		Results:    []*ResultSummary{},
	}

	for _, result := range r.Results {
		if result.ResourceID == resourceID {
			filtered.Results = append(filtered.Results, result)
			filtered.Scanned++
			if !result.Compliant {
				filtered.Violations++
			}
			if result.Error != nil {
				filtered.Errors++
			}
		}
	}

	return filtered
}

// LogSummary logs a summary of the audit results
func (r *AuditResults) LogSummary(t *testing.T) {
	t.Logf("Audit Summary: Scanned=%d, Violations=%d, Errors=%d", r.Scanned, r.Violations, r.Errors)
	for _, result := range r.Results {
		if !result.Compliant {
			t.Logf("  Violation: %s/%s - %s: %s", result.Service, result.ResourceType, result.ResourceID, result.Observation)
		}
		if result.Error != nil {
			t.Logf("  Error: %s/%s - %s: %v", result.Service, result.ResourceType, result.ResourceID, result.Error)
		}
	}
}

// LoadPolicyFromYAML loads a policy from a YAML string (useful for testing)
func (e *TestEngine) LoadPolicyFromYAML(t *testing.T, yamlContent string) *policy.Policy {
	t.Helper()

	// Create a temporary file with the YAML content
	tmpFile, err := os.CreateTemp("", "ospa-e2e-policy-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		tmpFile.Close()
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	return e.LoadPolicy(t, tmpFile.Name())
}

