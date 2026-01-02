package audit

import (
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// Result represents the outcome of an audit check
type Result struct {
	RuleID       string
	ResourceID   string
	ResourceName string
	ProjectID    string
	Compliant    bool
	Observation  string
	Error        error
	Rule         *policy.Rule

	// Remediation fields
	RemediationAttempted bool
	Remediated           bool
	RemediationError     error

	// Additional metadata
	UpdatedAt time.Time
	Status    string
}

