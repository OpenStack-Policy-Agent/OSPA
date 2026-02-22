package audit

import (
	"context"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// Auditor evaluates a resource against a rule
type Auditor interface {
	// Check evaluates a resource against a rule and returns a result
	Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*Result, error)

	// Fix applies remediation to a resource based on the rule
	Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error

	// ResourceType returns the resource type this auditor handles
	ResourceType() string

	// ImplementedChecks returns the list of CheckConditions field names
	// that this auditor actually evaluates. The orchestrator uses this to
	// warn when a policy rule references a check that no auditor handles.
	ImplementedChecks() []string
}
