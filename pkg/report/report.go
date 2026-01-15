package report

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/metrics"
)

// ResultWriter is a streaming result writer.
type ResultWriter interface {
	WriteResult(*audit.Result) error
	Close() error
}

// JSONLWriter writes one JSON object per line.
type JSONLWriter struct {
	enc *json.Encoder
}

func NewJSONLWriter(w io.Writer) *JSONLWriter {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &JSONLWriter{enc: enc}
}

type Finding struct {
	RuleID            string `json:"rule_id"`
	ResourceID        string `json:"resource_id"`
	ResourceName      string `json:"resource_name"`
	ResourceType      string `json:"resource_type,omitempty"`
	Service           string `json:"service,omitempty"`
	ProjectID         string `json:"project_id,omitempty"`
	Status            string `json:"status,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
	Compliant         bool   `json:"compliant"`
	ErrorKind         string `json:"error_kind,omitempty"`
	Mode              string `json:"mode,omitempty"`
	Observation       string `json:"observation,omitempty"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	Action            string `json:"action,omitempty"`
	Error             string `json:"error,omitempty"`

	RemediationAttempted bool   `json:"remediation_attempted,omitempty"`
	Remediated           bool   `json:"remediated,omitempty"`
	RemediationError     string `json:"remediation_error,omitempty"`
	RemediationErrorKind string `json:"remediation_error_kind,omitempty"`
	RemediationSkipped   bool   `json:"remediation_skipped,omitempty"`
	RemediationSkipReason string `json:"remediation_skip_reason,omitempty"`
}

func (w *JSONLWriter) WriteResult(r *audit.Result) error {
	f := Finding{
		RuleID:            r.RuleID,
		ResourceID:        r.ResourceID,
		ResourceName:      r.ResourceName,
		ProjectID:         r.ProjectID,
		Status:            r.Status,
		Compliant:         r.Compliant,
		Observation:       r.Observation,
		RemediationSkipped: r.RemediationSkipped,
		RemediationSkipReason: r.RemediationSkipReason,
		RemediationAttempted: r.RemediationAttempted,
		Remediated:           r.Remediated,
	}

	if r.Rule != nil {
		f.Action = r.Rule.Action
		f.ResourceType = r.Rule.Resource
		f.Service = r.Rule.Service
	}

	if !r.UpdatedAt.IsZero() {
		f.UpdatedAt = r.UpdatedAt.UTC().Format(time.RFC3339)
	}
	if r.Error != nil {
		f.Error = r.Error.Error()
		if r.ErrorKind != "" {
			f.ErrorKind = string(r.ErrorKind)
		}
	}
	if r.RemediationError != nil {
		f.RemediationError = r.RemediationError.Error()
		if r.RemediationErrorKind != "" {
			f.RemediationErrorKind = string(r.RemediationErrorKind)
		}
	}
	return w.enc.Encode(f)
}

func (w *JSONLWriter) Close() error {
	return nil
}

// Summary aggregates result counts.
type Summary struct {
	Scanned              int
	Violations           int
	Errors               int
	Written              int
	RemediationAttempted int
	Remediated           int
	RemediationSkipped   int
}

// ConsumeResults reads results, updates metrics, and writes output (if writer provided).
func ConsumeResults(results <-chan *audit.Result, writer ResultWriter) Summary {
	var summary Summary

	for result := range results {
		summary.Scanned++
		metrics.IncScanned()

		if result.Error != nil {
			summary.Errors++
			metrics.IncErrors()
		}
		if result.RemediationError != nil {
			summary.Errors++
			metrics.IncErrors()
		}
		if !result.Compliant {
			summary.Violations++
			metrics.IncViolations()
		}
		if result.RemediationAttempted {
			summary.RemediationAttempted++
			metrics.IncRemediationAttempted()
		}
		if result.Remediated {
			summary.Remediated++
			metrics.IncRemediated()
		}
		if result.RemediationSkipped {
			summary.RemediationSkipped++
			metrics.IncRemediationSkipped()
		}

		if writer != nil && (!result.Compliant || result.Error != nil || result.RemediationError != nil) {
			if err := writer.WriteResult(result); err == nil {
				summary.Written++
			}
		}
	}

	if writer != nil {
		_ = writer.Close()
	}

	return summary
}

func PrintSummary(out io.Writer, summary Summary) {
	fmt.Fprintln(out, "---- Summary ----")
	fmt.Fprintf(out, "Scanned: %d\nViolations: %d\nErrors: %d\n", summary.Scanned, summary.Violations, summary.Errors)
	fmt.Fprintf(out, "Remediation attempted: %d\nRemediated: %d\nRemediation skipped: %d\n",
		summary.RemediationAttempted, summary.Remediated, summary.RemediationSkipped)
}
