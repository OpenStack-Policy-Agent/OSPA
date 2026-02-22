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
	ErrorKind    ErrorKind
	Rule         *policy.Rule

	// Classification
	Severity string
	Category string
	GuideRef string

	// Remediation fields
	RemediationAttempted  bool
	Remediated            bool
	RemediationError      error
	RemediationErrorKind  ErrorKind
	RemediationSkipped    bool
	RemediationSkipReason string

	// Additional metadata
	UpdatedAt time.Time
	Status    string
}

// ErrorKind captures the source of an error to help reporting and metrics.
type ErrorKind string

const (
	ErrorKindAudit       ErrorKind = "audit"
	ErrorKindRemediation ErrorKind = "remediation"
)
