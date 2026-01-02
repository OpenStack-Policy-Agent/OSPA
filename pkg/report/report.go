package report

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
)

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
	Mode              string `json:"mode,omitempty"`
	Observation       string `json:"observation,omitempty"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	Action            string `json:"action,omitempty"`
	Error             string `json:"error,omitempty"`

	RemediationAttempted bool   `json:"remediation_attempted,omitempty"`
	Remediated           bool   `json:"remediated,omitempty"`
	RemediationError     string `json:"remediation_error,omitempty"`
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
	}
	if r.RemediationError != nil {
		f.RemediationError = r.RemediationError.Error()
	}
	return w.enc.Encode(f)
}

func PrintSummary(out io.Writer, scanned, violations, errors int) {
	fmt.Fprintln(out, "---- Summary ----")
	fmt.Fprintf(out, "Scanned: %d\nViolations: %d\nErrors: %d\n", scanned, violations, errors)
}
