//go:build e2e

package e2e

import (
	"sync"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/engine"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

func TestAgent_Audit_DetectsButDoesNotFix(t *testing.T) {
	client := getComputeClient(t)

	// Create one compliant (ACTIVE) and one noncompliant (SHUTOFF) server.
	compliant := createServer(t, client, uniqueName("ospa-e2e-compliant"))
	noncompliant := createServer(t, client, uniqueName("ospa-e2e-noncompliant"))
	stopServer(t, client, noncompliant.ID)

	// Re-fetch after stop so Status/Updated are current.
	noncompliantPtr, err := servers.Get(client, noncompliant.ID).Extract()
	if err != nil {
		t.Fatalf("servers.Get noncompliant after stop: %v", err)
	}
	noncompliant = *noncompliantPtr

	// Force a violation immediately by using OlderThan=0 in the rule (test-only).
	evaluator := engine.RuleSet{
		Now: time.Now,
		Rules: []engine.ServerRule{
			engine.StoppedOlderThanRule{
				RuleID:            "e2e.stopped_older_than_0",
				MatchStatus:       "SHUTOFF",
				OlderThan:         0,
				RecommendedAction: "delete",
				Mode:              "audit",
			},
		},
	}

	jobs := make(chan engine.Job, 8)
	results := make(chan engine.Result, 16)
	var wg sync.WaitGroup
	engine.StartWorkerPool(2, false, client, evaluator, jobs, results, &wg)

	jobs <- engine.Job{Server: compliant}
	jobs <- engine.Job{Server: noncompliant}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var (
		gotCompliant    *engine.Result
		gotNonCompliant *engine.Result
	)
	for r := range results {
		r := r
		switch r.ResourceID {
		case compliant.ID:
			gotCompliant = &r
		case noncompliant.ID:
			gotNonCompliant = &r
		}
	}
	if gotCompliant == nil || gotNonCompliant == nil {
		t.Fatalf("expected results for both servers (compliant=%v, noncompliant=%v)", gotCompliant != nil, gotNonCompliant != nil)
	}
	if !gotCompliant.Compliant {
		t.Fatalf("expected compliant server to be compliant; got %+v", *gotCompliant)
	}
	if gotNonCompliant.Compliant {
		t.Fatalf("expected noncompliant server to be noncompliant; got %+v", *gotNonCompliant)
	}
	if gotNonCompliant.RemediationAttempted {
		t.Fatalf("audit mode should not attempt remediation; got %+v", *gotNonCompliant)
	}

	// Ensure audit did NOT delete the server.
	if _, err := servers.Get(client, noncompliant.ID).Extract(); err != nil {
		t.Fatalf("expected server to still exist after audit; get err: %v", err)
	}
}

func TestAgent_EnforceApply_FixesByDeletingNonCompliant(t *testing.T) {
	client := getComputeClient(t)

	target := createServer(t, client, uniqueName("ospa-e2e-delete"))
	stopServer(t, client, target.ID)

	// Re-fetch after stop so Status/Updated are current.
	targetGet, err := servers.Get(client, target.ID).Extract()
	if err != nil {
		t.Fatalf("servers.Get target after stop: %v", err)
	}

	// Force a violation immediately by using OlderThan=0 in the rule (test-only).
	evaluator := engine.RuleSet{
		Now: time.Now,
		Rules: []engine.ServerRule{
			engine.StoppedOlderThanRule{
				RuleID:            "e2e.stopped_older_than_0",
				MatchStatus:       "SHUTOFF",
				OlderThan:         0,
				RecommendedAction: "delete",
				Mode:              "enforce",
			},
		},
	}

	jobs := make(chan engine.Job, 1)
	results := make(chan engine.Result, 4)
	var wg sync.WaitGroup

	engine.StartWorkerPool(1, true, client, evaluator, jobs, results, &wg)
	jobs <- engine.Job{Server: *targetGet}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var r *engine.Result
	for res := range results {
		res := res
		r = &res
	}
	if r == nil {
		t.Fatalf("expected a result")
	}
	if r.Compliant {
		t.Fatalf("expected noncompliant; got %+v", *r)
	}
	if !r.RemediationAttempted {
		t.Fatalf("expected remediation attempted; got %+v", *r)
	}
	if !r.Remediated {
		t.Fatalf("expected remediated=true (delete requested); got %+v", *r)
	}
	if r.RemediationError != nil {
		t.Fatalf("expected remediation error nil; got %+v", *r)
	}

	waitForDeleted(t, client, target.ID, 10*time.Minute)
}


