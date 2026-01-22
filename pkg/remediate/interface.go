package remediate

import (
	"context"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// Remediator applies remediation actions to resources
type Remediator interface {
	// Execute executes the remediation action
	Execute(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error

	// Action returns the action name this remediator handles
	Action() string
}
