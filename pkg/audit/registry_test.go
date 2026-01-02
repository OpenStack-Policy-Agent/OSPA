package audit_test

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

type fakeAuditor struct{ rt string }

func (a *fakeAuditor) ResourceType() string { return a.rt }
func (a *fakeAuditor) Check(context.Context, interface{}, *policy.Rule) (*audit.Result, error) {
	return &audit.Result{Compliant: true}, nil
}
func (a *fakeAuditor) Fix(context.Context, interface{}, interface{}, *policy.Rule) error { return nil }

func TestGet_NotFoundErrors(t *testing.T) {
	if _, err := audit.Get("nope", "x"); err == nil {
		t.Fatalf("Get(nope,x) error = nil, want error")
	}
}

func TestRegisterAndGet(t *testing.T) {
	audit.Register("svc-unit", "res-unit", &fakeAuditor{rt: "res-unit"})
	got, err := audit.Get("svc-unit", "res-unit")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ResourceType() != "res-unit" {
		t.Fatalf("auditor.ResourceType = %q, want %q", got.ResourceType(), "res-unit")
	}
}


