package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

func TestJSONLWriter_WriteResult_EmitsExpectedFields(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLWriter(&buf)

	now := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)
	r := &audit.Result{
		RuleID:               "r1",
		ResourceID:           "srv-1",
		ResourceName:         "srv",
		ProjectID:            "proj",
		Status:               "SHUTOFF",
		UpdatedAt:            now,
		Compliant:            false,
		Observation:          "too old",
		RemediationAttempted: true,
		Remediated:           false,
		RemediationError:     nil,
		Rule: &policy.Rule{
			Name:   "r1",
			Action: "delete",
		},
	}

	if err := w.WriteResult(r); err != nil {
		t.Fatalf("write result: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("unmarshal json: %v\nline=%s", err, line)
	}

	if m["rule_id"] != "r1" {
		t.Fatalf("expected rule_id r1, got %#v", m["rule_id"])
	}
	if m["action"] != "delete" {
		t.Fatalf("expected action delete, got %#v", m["action"])
	}
	if m["remediation_attempted"] != true {
		t.Fatalf("expected remediation_attempted true, got %#v", m["remediation_attempted"])
	}
	if _, ok := m["remediation_error"]; ok {
		t.Fatalf("did not expect remediation_error field when nil, got %#v", m["remediation_error"])
	}
}

func TestJSONLWriter_WriteResult_IncludesErrors(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLWriter(&buf)

	r := &audit.Result{
		RuleID:           "r1",
		ResourceID:       "srv-1",
		ResourceName:     "srv",
		Compliant:        false,
		Error:            errString("eval failed"),
		ErrorKind:        audit.ErrorKindAudit,
		RemediationError: errString("delete failed"),
		RemediationErrorKind: audit.ErrorKindRemediation,
		Rule: &policy.Rule{
			Name: "r1",
		},
	}

	if err := w.WriteResult(r); err != nil {
		t.Fatalf("write result: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("unmarshal json: %v\nline=%s", err, line)
	}
	if m["error"] != "eval failed" {
		t.Fatalf("expected error field, got %#v", m["error"])
	}
	if m["remediation_error"] != "delete failed" {
		t.Fatalf("expected remediation_error field, got %#v", m["remediation_error"])
	}
	if m["error_kind"] != string(audit.ErrorKindAudit) {
		t.Fatalf("expected error_kind field, got %#v", m["error_kind"])
	}
	if m["remediation_error_kind"] != string(audit.ErrorKindRemediation) {
		t.Fatalf("expected remediation_error_kind field, got %#v", m["remediation_error_kind"])
	}
}

type errString string

func (e errString) Error() string { return string(e) }


