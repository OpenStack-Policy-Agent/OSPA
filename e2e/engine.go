//go:build e2e

package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/orchestrator"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/report"
	"github.com/gophercloud/gophercloud"
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services"
)

// TestEngine provides a framework for running e2e tests
type TestEngine struct {
	Session      *auth.Session
	CloudName    string
	PolicyPath   string
	Workers      int
	Apply        bool
	AllTenants   bool
	AllowActions []string
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

// GetNetworkClient returns a gophercloud client for the Neutron service.
func (e *TestEngine) GetNetworkClient(t *testing.T) *gophercloud.ServiceClient {
	t.Helper()
	client, err := e.Session.GetNeutronClient()
	if err != nil {
		skipOrFail(t, "neutron", err)
	}
	return client
}

// GetComputeClient returns a gophercloud client for the Nova service.
func (e *TestEngine) GetComputeClient(t *testing.T) *gophercloud.ServiceClient {
	t.Helper()
	client, err := e.Session.GetNovaClient()
	if err != nil {
		skipOrFail(t, "nova", err)
	}
	return client
}

// GetBlockStorageClient returns a gophercloud client for the Cinder service.
func (e *TestEngine) GetBlockStorageClient(t *testing.T) *gophercloud.ServiceClient {
	t.Helper()
	client, err := e.Session.GetCinderClient()
	if err != nil {
		skipOrFail(t, "cinder", err)
	}
	return client
}

// skipOrFail skips the test if the service is not available in the catalog,
// otherwise fails fatally.
func skipOrFail(t *testing.T, service string, err error) {
	t.Helper()
	if strings.Contains(err.Error(), "unable to create a service client") {
		t.Skipf("%s service not available in catalog, skipping: %v", service, err)
	}
	t.Fatalf("Failed to get %s client: %v", service, err)
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
	if len(e.AllowActions) > 0 {
		orch.SetRemediationAllowlist(e.AllowActions)
	}
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
			RuleID:                result.RuleID,
			ResourceID:            result.ResourceID,
			ResourceName:          result.ResourceName,
			Service:               service,
			ResourceType:          resourceType,
			Compliant:             result.Compliant,
			Observation:           result.Observation,
			Error:                 result.Error,
			Severity:              result.Severity,
			Category:              result.Category,
			GuideRef:              result.GuideRef,
			RemediationAttempted:  result.RemediationAttempted,
			Remediated:            result.Remediated,
			RemediationSkipped:    result.RemediationSkipped,
			RemediationSkipReason: result.RemediationSkipReason,
			RemediationError:      result.RemediationError,
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

	Severity  string
	Category  string
	GuideRef  string

	RemediationAttempted  bool
	Remediated            bool
	RemediationSkipped    bool
	RemediationSkipReason string
	RemediationError      error
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

// RunAuditToFile runs an audit and writes results to a file in the given format.
// Returns the collected AuditResults and the path to the written file.
func (e *TestEngine) RunAuditToFile(t *testing.T, p *policy.Policy, format string) (results *AuditResults, filePath string) {
	t.Helper()

	outFile, err := os.CreateTemp("", "ospa-e2e-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	filePath = outFile.Name()

	writer, err := report.NewWriter(format, outFile)
	if err != nil {
		outFile.Close()
		t.Fatalf("Failed to create report writer: %v", err)
	}

	orch := orchestrator.NewOrchestrator(p, e.Session, e.Workers, e.Apply, e.AllTenants)
	defer orch.Stop()

	resultsChan, err := orch.Run()
	if err != nil {
		outFile.Close()
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	results = &AuditResults{Results: []*ResultSummary{}}
	for result := range resultsChan {
		if writeErr := writer.WriteResult(result); writeErr != nil {
			t.Errorf("Failed to write result: %v", writeErr)
		}
		results.Scanned++
		if result.Error != nil {
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
			Severity:     result.Severity,
			Category:     result.Category,
			GuideRef:     result.GuideRef,
		})
	}

	if closeErr := writer.Close(); closeErr != nil {
		t.Errorf("Failed to close writer: %v", closeErr)
	}
	outFile.Close()

	return results, filePath
}

// FilterByRuleID filters results by rule ID
func (r *AuditResults) FilterByRuleID(ruleID string) *AuditResults {
	filtered := &AuditResults{Results: []*ResultSummary{}}
	for _, result := range r.Results {
		if result.RuleID == ruleID {
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

// AssertClassification checks that all results have the expected severity, category, and guide_ref.
func (r *AuditResults) AssertClassification(t *testing.T, severity, category, guideRef string) {
	t.Helper()
	for _, result := range r.Results {
		if severity != "" && result.Severity != severity {
			t.Errorf("Result %s: expected severity %q, got %q", result.ResourceID, severity, result.Severity)
		}
		if category != "" && result.Category != category {
			t.Errorf("Result %s: expected category %q, got %q", result.ResourceID, category, result.Category)
		}
		if guideRef != "" && result.GuideRef != guideRef {
			t.Errorf("Result %s: expected guide_ref %q, got %q", result.ResourceID, guideRef, result.GuideRef)
		}
	}
}

// AssertRemediationSkipped checks that all non-compliant results have remediation skipped.
func (r *AuditResults) AssertRemediationSkipped(t *testing.T, expectedReason string) {
	t.Helper()
	for _, result := range r.Results {
		if result.Compliant {
			continue
		}
		if !result.RemediationSkipped {
			t.Errorf("Result %s: expected RemediationSkipped=true", result.ResourceID)
		}
		if expectedReason != "" && result.RemediationSkipReason != expectedReason {
			t.Errorf("Result %s: expected skip reason %q, got %q", result.ResourceID, expectedReason, result.RemediationSkipReason)
		}
	}
}

// FindLineEnd returns the index of the first newline in data, or len(data) if none.
// Useful for parsing NDJSON output (one JSON object per line).
func FindLineEnd(data []byte) int {
	for i, b := range data {
		if b == '\n' {
			return i
		}
	}
	return len(data)
}

