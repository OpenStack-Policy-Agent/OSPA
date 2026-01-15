package orchestrator

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/metrics"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/gophercloud/gophercloud"
)

// Orchestrator coordinates policy execution
type Orchestrator struct {
	policy      *policy.Policy
	session     *auth.Session
	workers     int
	apply       bool
	allTenants  bool
	ctx         context.Context
	cancel      context.CancelFunc
	resultsChan chan *audit.Result

	jobsBuffer    int
	resultsBuffer int

	ruleIndex       map[string]map[string][]*policy.Rule
	clientCache     map[string]*gophercloud.ServiceClient
	clientCacheLock sync.Mutex

	remediationAllowlist map[string]bool
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(p *policy.Policy, session *auth.Session, workers int, apply, allTenants bool) *Orchestrator {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	return &Orchestrator{
		policy:      p,
		session:     session,
		workers:     workers,
		apply:       apply,
		allTenants:  allTenants,
		ctx:         ctx,
		cancel:      cancel,
		resultsChan: make(chan *audit.Result, 100),
		jobsBuffer:  1000,
		resultsBuffer: 100,
		clientCache: make(map[string]*gophercloud.ServiceClient),
	}
}

// SetBuffers configures channel buffer sizes for jobs and results.
func (o *Orchestrator) SetBuffers(jobsBuffer, resultsBuffer int) {
	if jobsBuffer > 0 {
		o.jobsBuffer = jobsBuffer
	}
	if resultsBuffer > 0 {
		o.resultsBuffer = resultsBuffer
	}
}

// SetRemediationAllowlist sets which actions are allowed when apply mode is enabled.
// An empty list means allow all actions (current default behavior).
func (o *Orchestrator) SetRemediationAllowlist(actions []string) {
	if len(actions) == 0 {
		o.remediationAllowlist = nil
		return
	}
	allow := make(map[string]bool, len(actions))
	for _, action := range actions {
		if action == "" {
			continue
		}
		allow[action] = true
	}
	o.remediationAllowlist = allow
}

// Run executes the policy audit
func (o *Orchestrator) Run() (<-chan *audit.Result, error) {
	// Get all rules from policy
	rules := o.policy.GetAllRules()

	// Group rules by service and resource type for efficient discovery
	ruleGroups := make(map[string]map[string][]*policy.Rule)
	for i := range rules {
		rule := &rules[i]
		service := rule.Service
		resourceType := rule.Resource

		if ruleGroups[service] == nil {
			ruleGroups[service] = make(map[string][]*policy.Rule)
		}
		ruleGroups[service][resourceType] = append(ruleGroups[service][resourceType], rule)
	}
	o.ruleIndex = ruleGroups

	// Start worker pool
	var wg sync.WaitGroup
	jobsChan := make(chan discovery.Job, o.jobsBuffer)
	o.resultsChan = make(chan *audit.Result, o.resultsBuffer)

	// Start workers
	for i := 0; i < o.workers; i++ {
		wg.Add(1)
		go o.worker(i, jobsChan, &wg)
	}

	// Start discovery for each service/resource type
	var discoveryWg sync.WaitGroup
	for serviceName, resourceRules := range ruleGroups {
		service, err := services.Get(serviceName)
		if err != nil {
			slog.Warn("service not found", "service", serviceName, "error", err)
			metrics.IncServiceNotFound()
			continue
		}

		client, err := o.getClient(serviceName, service)
		if err != nil {
			slog.Warn("failed to get client", "service", serviceName, "error", err)
			metrics.IncClientErrors()
			continue
		}

		for resourceType, rules := range resourceRules {
			discoverer, err := service.GetResourceDiscoverer(resourceType)
			if err != nil {
				slog.Warn("discoverer not found", "service", serviceName, "resource", resourceType, "error", err)
				metrics.IncDiscovererNotFound()
				continue
			}

			discoveryWg.Add(1)
			go func(svc string, resType string, disc discovery.Discoverer, cli *gophercloud.ServiceClient, rls []*policy.Rule) {
				defer discoveryWg.Done()

				jobChan, err := disc.Discover(o.ctx, cli, o.allTenants)
				if err != nil {
					slog.Error("discovery error", "service", svc, "resource", resType, "error", err)
					metrics.IncDiscoveryErrors()
					return
				}

				for job := range jobChan {
					select {
					case <-o.ctx.Done():
						return
					case jobsChan <- job:
					}
				}
			}(serviceName, resourceType, discoverer, client, rules)
		}
	}

	// Close jobs channel when all discovery is done
	go func() {
		discoveryWg.Wait()
		close(jobsChan)
	}()

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(o.resultsChan)
	}()

	return o.resultsChan, nil
}

// worker processes jobs from the jobs channel
func (o *Orchestrator) worker(id int, jobsChan <-chan discovery.Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobsChan {
		select {
		case <-o.ctx.Done():
			return
		default:
		}

		// Get service and auditor
		service, err := services.Get(job.Service)
		if err != nil {
			slog.Warn("service not found", "worker", id, "service", job.Service, "error", err)
			metrics.IncServiceNotFound()
			continue
		}

		client, err := o.getClient(job.Service, service)
		if err != nil {
			slog.Warn("failed to get client", "worker", id, "service", job.Service, "error", err)
			metrics.IncClientErrors()
			continue
		}

		// Get rules for this service/resource type from the policy
		relevantRules := o.ruleIndex[job.Service][job.ResourceType]
		if len(relevantRules) == 0 {
			continue
		}

		// Process each relevant rule
		for _, rule := range relevantRules {
			// Get auditor
			auditor, err := service.GetResourceAuditor(job.ResourceType)
			if err != nil {
				slog.Warn("auditor not found", "worker", id, "service", job.Service, "resource", job.ResourceType, "error", err)
				metrics.IncAuditorNotFound()
				continue
			}

			// Check resource
			result, err := auditor.Check(o.ctx, job.Resource, rule)
			if err != nil {
				result = &audit.Result{
					RuleID:     rule.Name,
					ResourceID: job.ResourceID,
					Compliant:  false,
					Error:      err,
					ErrorKind:  audit.ErrorKindAudit,
					Rule:       rule,
				}
			}

			// Apply remediation if needed
			if !result.Compliant && result.Error == nil && rule.Action != "log" {
				if !o.apply {
					result.RemediationSkipped = true
					result.RemediationSkipReason = "dry-run"
				} else if !o.isActionAllowed(rule.Action) {
					result.RemediationSkipped = true
					result.RemediationSkipReason = "action_not_allowed"
				} else {
					result.RemediationAttempted = true
					if err := auditor.Fix(o.ctx, client, job.Resource, rule); err != nil {
						result.RemediationError = err
						result.RemediationErrorKind = audit.ErrorKindRemediation
					} else {
						result.Remediated = true
					}
				}
			}

			// Send result
			select {
			case <-o.ctx.Done():
				return
			case o.resultsChan <- result:
			}
		}
	}
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	o.cancel()
}

func (o *Orchestrator) getClient(serviceName string, service services.Service) (*gophercloud.ServiceClient, error) {
	o.clientCacheLock.Lock()
	defer o.clientCacheLock.Unlock()

	if client, ok := o.clientCache[serviceName]; ok {
		return client, nil
	}

	client, err := service.GetClient(o.session)
	if err != nil {
		return nil, err
	}
	o.clientCache[serviceName] = client
	return client, nil
}

func (o *Orchestrator) isActionAllowed(action string) bool {
	if o.remediationAllowlist == nil {
		return true
	}
	return o.remediationAllowlist[action]
}

