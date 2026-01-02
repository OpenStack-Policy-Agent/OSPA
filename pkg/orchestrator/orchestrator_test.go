package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/orchestrator"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
	"github.com/gophercloud/gophercloud"
)

type fakeDiscoverer struct {
	service string
	resType string
}

func (d *fakeDiscoverer) ResourceType() string { return d.resType }
func (d *fakeDiscoverer) Discover(ctx context.Context, _ *gophercloud.ServiceClient, _ bool) (<-chan discovery.Job, error) {
	ch := make(chan discovery.Job, 1)
	ch <- discovery.Job{
		Service:      d.service,
		ResourceType: d.resType,
		ResourceID:   "id-1",
		Resource:     map[string]any{"id": "id-1"},
		ProjectID:    "proj-1",
	}
	close(ch)
	return ch, nil
}

type fakeAuditor struct {
	resType string
	fixErr  error
	fixed   bool
}

func (a *fakeAuditor) ResourceType() string { return a.resType }
func (a *fakeAuditor) Check(context.Context, interface{}, *policy.Rule) (*audit.Result, error) {
	// Always non-compliant so remediation triggers in apply mode.
	return &audit.Result{Compliant: false}, nil
}
func (a *fakeAuditor) Fix(context.Context, interface{}, interface{}, *policy.Rule) error {
	a.fixed = true
	return a.fixErr
}

type fakeService struct {
	name      string
	resType   string
	disc      discovery.Discoverer
	aud       audit.Auditor
}

func (s *fakeService) Name() string { return s.name }
func (s *fakeService) GetClient(*auth.Session) (*gophercloud.ServiceClient, error) {
	return &gophercloud.ServiceClient{}, nil
}
func (s *fakeService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
	if resourceType != s.resType {
		return nil, nil
	}
	return s.aud, nil
}
func (s *fakeService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
	if resourceType != s.resType {
		return nil, nil
	}
	return s.disc, nil
}

func TestOrchestrator_Run_ProducesResultAndCanRemediate(t *testing.T) {
	const (
		svc    = "orchestrator-test-svc"
		res    = "thing"
		ruleID = "r1"
	)

	// Register the service + resource for policy validation and orchestrator lookup.
	services.RegisterResource(svc, res)

	aud := &fakeAuditor{resType: res}
	disc := &fakeDiscoverer{service: svc, resType: res}
	if err := services.Register(&fakeService{name: svc, resType: res, disc: disc, aud: aud}); err != nil {
		t.Fatalf("services.Register() = %v", err)
	}

	p := &policy.Policy{
		Version: "v1",
		Policies: []policy.ServicePolicy{
			{
				Service: svc,
				Rules: []policy.Rule{
					{
						Name:     ruleID,
						Service:  svc,
						Resource: res,
						Check:    policy.CheckConditions{Status: "active"},
						Action:   "delete",
					},
				},
			},
		},
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("policy.Validate() = %v", err)
	}

	o := orchestrator.NewOrchestrator(p, &auth.Session{CloudName: "test"}, 1, true, false)
	results, err := o.Run()
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}

	select {
	case r, ok := <-results:
		if !ok {
			t.Fatalf("results channel closed unexpectedly")
		}
		if r == nil {
			t.Fatalf("got nil result")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for result")
	}

	if !aud.fixed {
		t.Fatalf("expected auditor.Fix to be called in apply mode")
	}
}


