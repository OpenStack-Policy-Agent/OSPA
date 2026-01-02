package compute

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

func TestInstanceAuditor_Check_AgeGT_WithMetadataExemption(t *testing.T) {
	a := &InstanceAuditor{}

	old := time.Now().Add(-40 * 24 * time.Hour)
	srv := servers.Server{
		ID:       "srv-1",
		Name:     "srv",
		TenantID: "proj",
		Created:  old,
		Updated:  old,
		Metadata: map[string]string{"exempt": "true"},
		Status:   "ACTIVE",
	}

	rule := &policy.Rule{
		Name:     "r1",
		Service:  "nova",
		Resource: "instance",
		Action:   "log",
		Check: policy.CheckConditions{
			AgeGT: "30d",
			ExemptMetadata: &policy.MetadataMatch{
				Key:   "exempt",
				Value: "true",
			},
		},
	}

	res, err := a.Check(context.Background(), srv, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if res == nil {
		t.Fatalf("result is nil")
	}
	if !res.Compliant {
		t.Fatalf("expected compliant due to exemption, got non-compliant: %s", res.Observation)
	}
}


