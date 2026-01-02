package remediate_test

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/remediate"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

type fakeRemediator struct{ action string }

func (r *fakeRemediator) Action() string { return r.action }
func (r *fakeRemediator) Execute(context.Context, interface{}, interface{}, *policy.Rule) error {
	return nil
}

func TestBuiltInRemediatorsRegistered(t *testing.T) {
	for _, action := range []string{"log", "delete", "tag"} {
		if _, err := remediate.Get(action); err != nil {
			t.Fatalf("Get(%q) error = %v", action, err)
		}
	}
}

func TestGet_UnknownActionErrors(t *testing.T) {
	if _, err := remediate.Get("nope"); err == nil {
		t.Fatalf("Get(nope) error = nil, want error")
	}
}

func TestRegister_OverridesExisting(t *testing.T) {
	remediate.Register(&fakeRemediator{action: "log"})
	if _, err := remediate.Get("log"); err != nil {
		t.Fatalf("Get(log) error = %v", err)
	}
}


