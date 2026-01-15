package report

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
)

// CSVWriter writes CSV findings for non-compliant resources and errors.
type CSVWriter struct {
	writer     *csv.Writer
	wroteHeader bool
}

func NewCSVWriter(w io.Writer) *CSVWriter {
	return &CSVWriter{writer: csv.NewWriter(w)}
}

func (w *CSVWriter) WriteResult(r *audit.Result) error {
	if !w.wroteHeader {
		header := []string{
			"rule_id",
			"resource_id",
			"resource_name",
			"resource_type",
			"service",
			"project_id",
			"status",
			"updated_at",
			"compliant",
			"error_kind",
			"observation",
			"action",
			"error",
			"remediation_attempted",
			"remediated",
			"remediation_error",
			"remediation_error_kind",
			"remediation_skipped",
			"remediation_skip_reason",
		}
		if err := w.writer.Write(header); err != nil {
			return err
		}
		w.wroteHeader = true
	}

	resourceType := ""
	service := ""
	action := ""
	if r.Rule != nil {
		resourceType = r.Rule.Resource
		service = r.Rule.Service
		action = r.Rule.Action
	}

	updatedAt := ""
	if !r.UpdatedAt.IsZero() {
		updatedAt = r.UpdatedAt.UTC().Format(time.RFC3339)
	}

	errorKind := ""
	if r.ErrorKind != "" {
		errorKind = string(r.ErrorKind)
	}

	remediationErrorKind := ""
	if r.RemediationErrorKind != "" {
		remediationErrorKind = string(r.RemediationErrorKind)
	}

	errorText := ""
	if r.Error != nil {
		errorText = r.Error.Error()
	}

	remediationErrorText := ""
	if r.RemediationError != nil {
		remediationErrorText = r.RemediationError.Error()
	}

	record := []string{
		r.RuleID,
		r.ResourceID,
		r.ResourceName,
		resourceType,
		service,
		r.ProjectID,
		r.Status,
		updatedAt,
		boolToString(r.Compliant),
		errorKind,
		r.Observation,
		action,
		errorText,
		boolToString(r.RemediationAttempted),
		boolToString(r.Remediated),
		remediationErrorText,
		remediationErrorKind,
		boolToString(r.RemediationSkipped),
		r.RemediationSkipReason,
	}

	return w.writer.Write(record)
}

func (w *CSVWriter) Close() error {
	w.writer.Flush()
	return w.writer.Error()
}

func boolToString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

