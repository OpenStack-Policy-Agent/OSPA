package neutron

import (
	"context"
	"testing"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

func TestSecurityGroupRuleAuditor_ResourceType(t *testing.T) {
	auditor := &SecurityGroupRuleAuditor{}
	if got := auditor.ResourceType(); got != "security_group_rule" {
		t.Errorf("ResourceType() = %q, want %q", got, "security_group_rule")
	}
}

func TestSecurityGroupRuleAuditor_Check_SSHOpenToWorld(t *testing.T) {
	auditor := &SecurityGroupRuleAuditor{}

	// Create a "dangerous" SSH rule - port 22 open to world
	resource := rules.SecGroupRule{
		ID:             "test-rule-id",
		SecGroupID:     "test-sg-id",
		TenantID:       "test-tenant-id",
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   22,
		PortRangeMax:   22,
		RemoteIPPrefix: "0.0.0.0/0",
	}

	rule := &policy.Rule{
		Name:     "test-ssh-open-to-world",
		Service:  "neutron",
		Resource: "security_group_rule",
		Check: policy.CheckConditions{
			Direction:      "ingress",
			Ethertype:      "IPv4",
			Protocol:       "tcp",
			Port:           22,
			RemoteIPPrefix: "0.0.0.0/0",
		},
		Action: "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result == nil {
		t.Fatal("Check() returned nil result")
	}
	if result.RuleID != rule.Name {
		t.Errorf("Result.RuleID = %q, want %q", result.RuleID, rule.Name)
	}
	if result.ResourceID != resource.ID {
		t.Errorf("Result.ResourceID = %q, want %q", result.ResourceID, resource.ID)
	}
	// Rule should be non-compliant (it's a dangerous SSH rule)
	if result.Compliant {
		t.Error("Expected SSH open to world rule to be non-compliant")
	}
}

func TestSecurityGroupRuleAuditor_Check_SafeRule(t *testing.T) {
	auditor := &SecurityGroupRuleAuditor{}

	// Create a "safe" SSH rule - port 22 from private network
	resource := rules.SecGroupRule{
		ID:             "test-rule-id",
		SecGroupID:     "test-sg-id",
		TenantID:       "test-tenant-id",
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   22,
		PortRangeMax:   22,
		RemoteIPPrefix: "10.0.0.0/8", // Private network, not 0.0.0.0/0
	}

	rule := &policy.Rule{
		Name:     "test-ssh-open-to-world",
		Service:  "neutron",
		Resource: "security_group_rule",
		Check: policy.CheckConditions{
			Direction:      "ingress",
			Ethertype:      "IPv4",
			Protocol:       "tcp",
			Port:           22,
			RemoteIPPrefix: "0.0.0.0/0", // Looking for open to world
		},
		Action: "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result == nil {
		t.Fatal("Check() returned nil result")
	}
	// Rule should be compliant (it's NOT open to world)
	if !result.Compliant {
		t.Error("Expected safe SSH rule (private network) to be compliant")
	}
}

func TestSecurityGroupRuleAuditor_Check_PortRange(t *testing.T) {
	auditor := &SecurityGroupRuleAuditor{}

	// Create a rule with a port range
	resource := rules.SecGroupRule{
		ID:             "test-rule-id",
		SecGroupID:     "test-sg-id",
		TenantID:       "test-tenant-id",
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   20,
		PortRangeMax:   25, // Includes port 22
		RemoteIPPrefix: "0.0.0.0/0",
	}

	rule := &policy.Rule{
		Name:     "test-ssh-port-range",
		Service:  "neutron",
		Resource: "security_group_rule",
		Check: policy.CheckConditions{
			Direction:      "ingress",
			Protocol:       "tcp",
			Port:           22, // Should match since 22 is in range 20-25
			RemoteIPPrefix: "0.0.0.0/0",
		},
		Action: "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	// Rule should be non-compliant (port 22 is in the range)
	if result.Compliant {
		t.Error("Expected rule with port range including 22 to be non-compliant")
	}
}

func TestSecurityGroupRuleAuditor_Check_PartialMatch(t *testing.T) {
	auditor := &SecurityGroupRuleAuditor{}

	// Create a rule that matches some but not all criteria
	resource := rules.SecGroupRule{
		ID:             "test-rule-id",
		SecGroupID:     "test-sg-id",
		TenantID:       "test-tenant-id",
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "udp", // Different protocol
		PortRangeMin:   22,
		PortRangeMax:   22,
		RemoteIPPrefix: "0.0.0.0/0",
	}

	rule := &policy.Rule{
		Name:     "test-ssh-tcp-only",
		Service:  "neutron",
		Resource: "security_group_rule",
		Check: policy.CheckConditions{
			Direction:      "ingress",
			Protocol:       "tcp", // Looking for TCP, but resource is UDP
			Port:           22,
			RemoteIPPrefix: "0.0.0.0/0",
		},
		Action: "log",
	}

	result, err := auditor.Check(context.Background(), resource, rule)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	// Rule should be compliant (protocol doesn't match)
	if !result.Compliant {
		t.Error("Expected UDP rule to be compliant when looking for TCP")
	}
}

func TestSecurityGroupRuleAuditor_Fix(t *testing.T) {
	t.Skip("Fix() requires a mock gophercloud client")
}

func TestBuildRuleName(t *testing.T) {
	tests := []struct {
		name     string
		rule     rules.SecGroupRule
		expected string
	}{
		{
			name: "SSH ingress from anywhere",
			rule: rules.SecGroupRule{
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   22,
				PortRangeMax:   22,
				RemoteIPPrefix: "0.0.0.0/0",
			},
			expected: "ingress/tcp:22 from 0.0.0.0/0",
		},
		{
			name: "Port range",
			rule: rules.SecGroupRule{
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   80,
				PortRangeMax:   443,
				RemoteIPPrefix: "10.0.0.0/8",
			},
			expected: "ingress/tcp:80-443 from 10.0.0.0/8",
		},
		{
			name: "All traffic",
			rule: rules.SecGroupRule{
				Direction:      "egress",
				Protocol:       "",
				PortRangeMin:   0,
				PortRangeMax:   0,
				RemoteIPPrefix: "",
			},
			expected: "egress/any from any",
		},
		{
			name: "Remote security group",
			rule: rules.SecGroupRule{
				Direction:     "ingress",
				Protocol:      "tcp",
				PortRangeMin:  22,
				PortRangeMax:  22,
				RemoteGroupID: "sg-12345",
			},
			expected: "ingress/tcp:22 from sg:sg-12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildRuleName(tt.rule)
			if got != tt.expected {
				t.Errorf("buildRuleName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPortMatches(t *testing.T) {
	tests := []struct {
		name     string
		min      int
		max      int
		port     int
		expected bool
	}{
		{"exact match", 22, 22, 22, true},
		{"in range", 20, 25, 22, true},
		{"below range", 20, 25, 19, false},
		{"above range", 20, 25, 26, false},
		{"all ports (0-0)", 0, 0, 22, true},
		{"all ports matches any", 0, 0, 443, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := portMatches(tt.min, tt.max, tt.port)
			if got != tt.expected {
				t.Errorf("portMatches(%d, %d, %d) = %v, want %v", tt.min, tt.max, tt.port, got, tt.expected)
			}
		})
	}
}
