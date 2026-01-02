package compute

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
)

func TestKeypairAuditor_Check_UnusedMarksNonCompliant(t *testing.T) {
	a := &KeypairAuditor{}

	kp := keypairs.KeyPair{
		Name:   "kp1",
		UserID: "user1",
	}

	rule := &policy.Rule{
		Name:     "r1",
		Service:  "nova",
		Resource: "keypair",
		Action:   "log",
		Check: policy.CheckConditions{
			Unused: true,
		},
	}

	res, err := a.Check(context.Background(), kp, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if res.Compliant {
		t.Fatalf("expected non-compliant when check.unused is true")
	}
	if res.Observation == "" {
		t.Fatalf("expected observation to be set")
	}
}


