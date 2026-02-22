package common

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// ResourceAdapter provides a uniform interface for extracting common fields
// from any OpenStack resource. Each service's audit package implements thin
// wrappers that satisfy this interface for its gophercloud types.
type ResourceAdapter interface {
	GetID() string
	GetName() string
	GetProjectID() string
	GetStatus() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// BuildBaseResult constructs an audit.Result pre-populated with fields from
// the adapter and the rule. Auditors should call this first, then layer on
// service-specific logic.
func BuildBaseResult(a ResourceAdapter, rule *policy.Rule) *audit.Result {
	return &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   a.GetID(),
		ResourceName: a.GetName(),
		ProjectID:    a.GetProjectID(),
		Status:       a.GetStatus(),
		UpdatedAt:    a.GetUpdatedAt(),
		Compliant:    true,
		Rule:         rule,
	}
}

// CheckExemptByName returns true (and sets the observation) when the
// resource name matches any pattern in rule.Check.ExemptNames.
// Supports both exact matches and glob patterns.
func CheckExemptByName(a ResourceAdapter, rule *policy.Rule, result *audit.Result) bool {
	name := a.GetName()
	for _, pattern := range rule.Check.ExemptNames {
		if name == pattern {
			result.Compliant = true
			result.Observation = "exempt by name"
			return true
		}
		if matched, _ := filepath.Match(pattern, name); matched {
			result.Compliant = true
			result.Observation = "exempt by name"
			return true
		}
	}
	return false
}

// CheckStatus marks the result non-compliant when the resource status
// matches the value declared in the policy rule.
func CheckStatus(a ResourceAdapter, rule *policy.Rule, result *audit.Result) {
	if rule.Check.Status == "" {
		return
	}
	if a.GetStatus() == rule.Check.Status {
		result.Compliant = false
		result.Observation = fmt.Sprintf("status is %s", a.GetStatus())
	}
}

// CheckAgeGT marks the result non-compliant when the resource is older
// than the duration encoded in rule.Check.AgeGT. It prefers UpdatedAt
// and falls back to CreatedAt.
func CheckAgeGT(a ResourceAdapter, rule *policy.Rule, result *audit.Result) error {
	if rule.Check.AgeGT == "" {
		return nil
	}

	age, err := rule.Check.ParseAgeGT()
	if err != nil {
		return fmt.Errorf("parsing age_gt: %w", err)
	}

	ts := a.GetUpdatedAt()
	if ts.IsZero() {
		ts = a.GetCreatedAt()
	}
	if ts.IsZero() {
		return nil
	}

	if time.Since(ts) > age {
		result.Compliant = false
		result.Observation = fmt.Sprintf("resource is older than %s (last updated: %s)",
			rule.Check.AgeGT, ts.Format(time.RFC3339))
	}
	return nil
}

// RunCommonChecks executes the universal check sequence that applies to
// every resource type: exempt_names -> status -> age_gt.
//
// It returns true if the resource is exempt (and therefore the auditor
// should short-circuit). The caller is responsible for unused and any
// domain-specific checks.
func RunCommonChecks(a ResourceAdapter, rule *policy.Rule, result *audit.Result) (exempt bool, err error) {
	if CheckExemptByName(a, rule, result) {
		return true, nil
	}

	CheckStatus(a, rule, result)

	if err := CheckAgeGT(a, rule, result); err != nil {
		return false, err
	}

	return false, nil
}
