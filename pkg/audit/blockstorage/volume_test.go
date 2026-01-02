package blockstorage

import (
	"context"
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
)

func TestVolumeAuditor_Check_AgeGT_MarksNonCompliant(t *testing.T) {
	a := &VolumeAuditor{}

	old := time.Now().Add(-40 * 24 * time.Hour)
	v := volumes.Volume{
		ID:        "vol-1",
		Name:      "vol",
		Status:    "available",
		CreatedAt: old,
		UpdatedAt: old,
	}

	rule := &policy.Rule{
		Name:     "r1",
		Service:  "cinder",
		Resource: "volume",
		Action:   "log",
		Check: policy.CheckConditions{
			AgeGT: "30d",
		},
	}

	res, err := a.Check(context.Background(), v, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if res.Compliant {
		t.Fatalf("expected non-compliant for old volume")
	}
}


