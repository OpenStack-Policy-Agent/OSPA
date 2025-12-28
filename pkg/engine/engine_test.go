package engine

import (
	"sync"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

type fakeEvaluator struct {
	results []Result
}

func (f fakeEvaluator) EvaluateServer(_ servers.Server) []Result { return f.results }

func TestStoppedOlderThanRule_Evaluate_CompliantWhenStatusDoesNotMatch(t *testing.T) {
	now := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)
	s := servers.Server{
		ID:       "srv-1",
		Name:     "test",
		TenantID: "proj-1",
		Status:   "ACTIVE",
		Updated:  now.Add(-365 * 24 * time.Hour),
	}

	rule := StoppedOlderThanRule{
		RuleID:            "r1",
		MatchStatus:       "SHUTOFF",
		OlderThan:         30 * 24 * time.Hour,
		RecommendedAction: "delete",
		Mode:              "enforce",
	}

	res := rule.Evaluate(s, now)
	if !res.Compliant {
		t.Fatalf("expected compliant when status does not match; got %+v", res)
	}
	if res.Mode != "enforce" {
		t.Fatalf("expected mode propagated; got %q", res.Mode)
	}
}

func TestStoppedOlderThanRule_Evaluate_NonCompliantWhenOlderThanThreshold(t *testing.T) {
	now := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)
	s := servers.Server{
		ID:       "srv-2",
		Name:     "zombie",
		TenantID: "proj-2",
		Status:   "SHUTOFF",
		Updated:  now.Add(-31 * 24 * time.Hour),
	}

	rule := StoppedOlderThanRule{
		RuleID:            "r2",
		MatchStatus:       "SHUTOFF",
		OlderThan:         30 * 24 * time.Hour,
		RecommendedAction: "delete",
		Mode:              "enforce",
	}

	res := rule.Evaluate(s, now)
	if res.Compliant {
		t.Fatalf("expected non-compliant; got %+v", res)
	}
	if res.RecommendedAction != "delete" {
		t.Fatalf("expected recommended action delete; got %q", res.RecommendedAction)
	}
	if res.Observation == "" {
		t.Fatalf("expected observation to be set")
	}
}

func TestStoppedOlderThanRule_Evaluate_ErrorWhenUpdatedIsZero(t *testing.T) {
	now := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)
	s := servers.Server{
		ID:      "srv-3",
		Name:    "bad-updated",
		Status:  "SHUTOFF",
		Updated: time.Time{},
	}

	rule := StoppedOlderThanRule{
		RuleID:      "r3",
		MatchStatus: "SHUTOFF",
		OlderThan:   30 * 24 * time.Hour,
		Mode:        "enforce",
	}

	res := rule.Evaluate(s, now)
	if res.Compliant {
		t.Fatalf("expected non-compliant when updated is zero; got %+v", res)
	}
	if res.Error == nil {
		t.Fatalf("expected error when updated is zero")
	}
}

func TestStartWorkerPool_RemediationAttemptedInEnforceModeAndNonCompliant(t *testing.T) {
	// We don't want real OpenStack calls here, so we run with apply=false.
	eval := fakeEvaluator{
		results: []Result{
			{
				RuleID:            "r-enforce",
				ResourceID:        "srv-x",
				ResourceName:      "x",
				Compliant:         false,
				RecommendedAction: "delete",
				Mode:              "enforce",
			},
			{
				RuleID:       "r-audit",
				ResourceID:   "srv-y",
				ResourceName: "y",
				Compliant:    false,
				Mode:         "audit",
			},
			{
				RuleID:       "r-ok",
				ResourceID:   "srv-z",
				ResourceName: "z",
				Compliant:    true,
				Mode:         "enforce",
			},
		},
	}

	jobs := make(chan Job, 1)
	results := make(chan Result, 16)

	var wg sync.WaitGroup
	StartWorkerPool(1, false, &gophercloud.ServiceClient{}, eval, jobs, results, &wg)

	jobs <- Job{Server: servers.Server{ID: "ignored"}}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var got []Result
	for r := range results {
		got = append(got, r)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}

	// enforce + noncompliant => remediation attempted (even in dry-run)
	if !got[0].RemediationAttempted {
		t.Fatalf("expected remediation attempted for enforce/noncompliant; got %+v", got[0])
	}
	if got[0].Remediated {
		t.Fatalf("did not expect remediated when apply=false; got %+v", got[0])
	}
	if got[0].RemediationError != nil {
		t.Fatalf("did not expect remediation error when apply=false; got %+v", got[0])
	}

	// audit => no remediation attempt
	if got[1].RemediationAttempted {
		t.Fatalf("did not expect remediation attempt in audit mode; got %+v", got[1])
	}

	// compliant => no remediation attempt
	if got[2].RemediationAttempted {
		t.Fatalf("did not expect remediation attempt when compliant; got %+v", got[2])
	}
}


