package remediate

import (
	"context"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// LogRemediator handles "log" action (no-op, just logs)
type LogRemediator struct{}

func (r *LogRemediator) Action() string {
	return "log"
}

func (r *LogRemediator) Execute(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	// Log action is a no-op - violations are just logged
	return nil
}

// DeleteRemediator handles "delete" action
// Note: Actual deletion is handled by the specific auditor's Fix() method
type DeleteRemediator struct{}

func (r *DeleteRemediator) Action() string {
	return "delete"
}

func (r *DeleteRemediator) Execute(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	// Deletion is handled by the auditor's Fix() method
	// This is a placeholder that could coordinate deletion if needed
	return nil
}

// TagRemediator handles "tag" action
// Note: Actual tagging is handled by the specific auditor's Fix() method
type TagRemediator struct{}

func (r *TagRemediator) Action() string {
	return "tag"
}

func (r *TagRemediator) Execute(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	// Tagging is handled by the auditor's Fix() method
	// This is a placeholder that could coordinate tagging if needed
	return nil
}

// ExecuteRemediation executes remediation using the appropriate remediator or auditor
func ExecuteRemediation(ctx context.Context, auditor audit.Auditor, client interface{}, resource interface{}, rule *policy.Rule) error {
	// Use the auditor's Fix() method directly
	return auditor.Fix(ctx, client, resource, rule)
}

func init() {
	Register(&LogRemediator{})
	Register(&DeleteRemediator{})
	Register(&TagRemediator{})
}

