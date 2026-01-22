package catalog_test

import (
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/catalog"
)

func TestRegisterAndQuery(t *testing.T) {
	catalog.RegisterResource("svc", "res")
	if !catalog.IsResourceSupported("svc", "res") {
		t.Fatalf("expected svc/res supported")
	}
	if got := catalog.GetServiceResources("svc"); len(got) == 0 {
		t.Fatalf("expected service resources to be non-empty")
	}
	if got := catalog.GetSupportedResources(); !got["svc"]["res"] {
		t.Fatalf("expected supported resources map to include svc/res")
	}
}
