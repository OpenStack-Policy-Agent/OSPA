package services_test

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/catalog"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/services"

	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services"
	"github.com/gophercloud/gophercloud"
)

type fakeService struct{ name string }

func (s *fakeService) Name() string { return s.name }
func (s *fakeService) GetClient(_ *auth.Session) (*gophercloud.ServiceClient, error) {
	return &gophercloud.ServiceClient{}, nil
}
func (s *fakeService) GetResourceAuditor(string) (audit.Auditor, error)           { return nil, nil }
func (s *fakeService) GetResourceDiscoverer(string) (discovery.Discoverer, error) { return nil, nil }

func TestGet_KnownServiceRegistered(t *testing.T) {
	// Registered via blank import above.
	if _, err := services.Get("nova"); err != nil {
		t.Fatalf("Get(nova) error = %v", err)
	}
}

func TestGetSupportedResources_IncludesKnownResources(t *testing.T) {
	res := services.GetSupportedResources()
	if !res["nova"]["instance"] {
		t.Fatalf("expected nova/instance to be supported")
	}
	if !res["neutron"]["security_group_rule"] {
		t.Fatalf("expected neutron/security_group_rule to be supported")
	}
	if !res["cinder"]["volume"] {
		t.Fatalf("expected cinder/volume to be supported")
	}
}

func TestRegisterResource_WiresToCatalog(t *testing.T) {
	services.RegisterResource("unit-test-svc", "unit-test-res")
	if !catalog.IsResourceSupported("unit-test-svc", "unit-test-res") {
		t.Fatalf("expected catalog to contain unit-test-svc/unit-test-res")
	}
}

func TestRegister_DuplicateServiceFails(t *testing.T) {
	svc := &fakeService{name: "unit-test-dup-service"}
	if err := services.Register(svc); err != nil {
		t.Fatalf("Register() first call error = %v", err)
	}
	if err := services.Register(svc); err == nil {
		t.Fatalf("Register() second call error = nil, want error")
	}
}
