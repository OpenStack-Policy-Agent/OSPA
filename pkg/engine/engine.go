package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

type Evaluator interface {
	EvaluateServer(s servers.Server) []Result
}

// Job represents a unit of work (a resource to check).
type Job struct {
	Server servers.Server
}

// Result represents the outcome of a check/fix.
type Result struct {
	RuleID            string
	ResourceID        string
	ResourceName      string
	ProjectID         string
	Status            string
	UpdatedAt         time.Time
	Compliant         bool
	Observation       string
	RecommendedAction string
	Error             error

	// Mode controls whether remediation should be attempted (audit/enforce).
	Mode string

	// Remediation fields are populated only when Mode == "enforce".
	RemediationAttempted bool
	Remediated           bool
	RemediationError     error
}

type ServerRule interface {
	ID() string
	Evaluate(s servers.Server, now time.Time) Result
}

type RuleSet struct {
	Now   func() time.Time
	Rules []ServerRule
}

func (rs RuleSet) EvaluateServer(s servers.Server) []Result {
	nowFn := rs.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	now := nowFn()

	out := make([]Result, 0, len(rs.Rules))
	for _, r := range rs.Rules {
		out = append(out, r.Evaluate(s, now))
	}
	return out
}

// StoppedOlderThanRule flags servers in a given status (typically SHUTOFF) whose
// Updated timestamp is older than a threshold.
type StoppedOlderThanRule struct {
	RuleID            string
	MatchStatus       string
	OlderThan         time.Duration
	RecommendedAction string
	Mode              string
}

func (r StoppedOlderThanRule) ID() string { return r.RuleID }

func (r StoppedOlderThanRule) Evaluate(s servers.Server, now time.Time) Result {
	res := Result{
		RuleID:            r.RuleID,
		ResourceID:        s.ID,
		ResourceName:      s.Name,
		ProjectID:         s.TenantID,
		Status:            s.Status,
		UpdatedAt:         s.Updated,
		Compliant:         true,
		RecommendedAction: r.RecommendedAction,
		Mode:              r.Mode,
	}

	// Some Nova list responses may omit Updated; fall back to Created when needed.
	evalTime := s.Updated
	if evalTime.IsZero() {
		evalTime = s.Created
		// Keep UpdatedAt as-is (for reporting). Observation will still describe the status/age.
	}
	if evalTime.IsZero() {
		res.Compliant = false
		res.Error = fmt.Errorf("server.updated and server.created are zero (cannot evaluate age)")
		return res
	}

	if r.MatchStatus != "" && s.Status != r.MatchStatus {
		return res
	}

	age := now.Sub(evalTime)
	if age >= r.OlderThan {
		res.Compliant = false
		res.Observation = fmt.Sprintf("server is %s and updated %s ago (>= %s)", s.Status, age.Round(time.Second), r.OlderThan)
	}

	return res
}

// StartWorkerPool spins up 'count' goroutines to process jobs.
func StartWorkerPool(count int, apply bool, client *gophercloud.ServiceClient, evaluator Evaluator, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	if count <= 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go worker(i, apply, client, evaluator, jobs, results, wg)
	}
}

// worker processes individual jobs.
func worker(id int, apply bool, client *gophercloud.ServiceClient, evaluator Evaluator, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		rs := evaluator.EvaluateServer(job.Server)
		for _, r := range rs {
			// Remediate in-worker to keep "Correct" parallelized.
			if r.Mode == "enforce" && !r.Compliant && r.Error == nil {
				r.RemediationAttempted = true
				if apply {
					if err := remediate(client, r); err != nil {
						r.RemediationError = err
					} else {
						r.Remediated = true
					}
				}
			}
			results <- r
		}
	}
}

func remediate(client *gophercloud.ServiceClient, r Result) error {
	switch r.RecommendedAction {
	case "delete":
		// Delete the server (best-effort). Caller decides whether this is allowed.
		return servers.Delete(client, r.ResourceID).ExtractErr()
	default:
		return fmt.Errorf("unsupported remediation action %q", r.RecommendedAction)
	}
}
