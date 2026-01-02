package network

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

func TestSecurityGroupRuleAuditor_Check_AllConditionsMatch_IsViolation(t *testing.T) {
	a := &SecurityGroupRuleAuditor{}

	sg := rules.SecGroupRule{
		ID:             "r1",
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   22,
		RemoteIPPrefix: "0.0.0.0/0",
		TenantID:       "proj",
	}

	rule := &policy.Rule{
		Name:     "policy-rule",
		Service:  "neutron",
		Resource: "security_group_rule",
		Action:   "log",
		Check: policy.CheckConditions{
			Direction:      "ingress",
			Ethertype:      "IPv4",
			Protocol:       "tcp",
			Port:           22,
			RemoteIPPrefix: "0.0.0.0/0",
		},
	}

	res, err := a.Check(context.Background(), sg, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if res.Compliant {
		t.Fatalf("expected violation (non-compliant) when all conditions match")
	}
}


