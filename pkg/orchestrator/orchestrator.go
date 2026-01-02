package orchestrator

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
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
	}
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

	// Start worker pool
	var wg sync.WaitGroup
	jobsChan := make(chan discovery.Job, 1000)

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
			log.Printf("Warning: service %q not found: %v", serviceName, err)
			continue
		}

		client, err := service.GetClient(o.session)
		if err != nil {
			log.Printf("Warning: failed to get client for service %q: %v", serviceName, err)
			continue
		}

		for resourceType, rules := range resourceRules {
			discoverer, err := service.GetResourceDiscoverer(resourceType)
			if err != nil {
				log.Printf("Warning: discoverer not found for %q/%q: %v", serviceName, resourceType, err)
				continue
			}

			discoveryWg.Add(1)
			go func(svc string, resType string, disc discovery.Discoverer, cli *gophercloud.ServiceClient, rls []*policy.Rule) {
				defer discoveryWg.Done()

				jobChan, err := disc.Discover(o.ctx, cli, o.allTenants)
				if err != nil {
					log.Printf("Error discovering %q/%q: %v", svc, resType, err)
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
			log.Printf("Worker %d: service %q not found: %v", id, job.Service, err)
			continue
		}

		client, err := service.GetClient(o.session)
		if err != nil {
			log.Printf("Worker %d: failed to get client for service %q: %v", id, job.Service, err)
			continue
		}

		// Get rules for this service/resource type from the policy
		var relevantRules []*policy.Rule
		for _, sp := range o.policy.Policies {
			if sp.Service != job.Service {
				continue
			}
			for i := range sp.Rules {
				rule := &sp.Rules[i]
				if rule.Resource == job.ResourceType {
					relevantRules = append(relevantRules, rule)
				}
			}
		}

		// Process each relevant rule
		for _, rule := range relevantRules {
			// Get auditor
			auditor, err := service.GetResourceAuditor(job.ResourceType)
			if err != nil {
				log.Printf("Worker %d: auditor not found for %q/%q: %v", id, job.Service, job.ResourceType, err)
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
					Rule:       rule,
				}
			}

			// Apply remediation if needed
			if !result.Compliant && result.Error == nil && o.apply && rule.Action != "log" {
				result.RemediationAttempted = true
				if err := auditor.Fix(o.ctx, client, job.Resource, rule); err != nil {
					result.RemediationError = err
				} else {
					result.Remediated = true
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

