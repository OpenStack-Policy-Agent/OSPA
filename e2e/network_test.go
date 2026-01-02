//go:build e2e

package e2e

import (
	"testing"
)

// TestNetwork_SecurityGroupRuleAudit tests Neutron security group rule auditing
func TestNetwork_SecurityGroupRuleAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-rule-ssh
      description: Test SSH rule check
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 22
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	sgRuleResults := results.FilterByService("neutron").FilterByResourceType("security_group_rule")
	sgRuleResults.LogSummary(t)
}

// TestNetwork_FloatingIPAudit tests Neutron floating IP auditing
func TestNetwork_FloatingIPAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-fip-unassociated
      description: Test unassociated floating IP check
      service: neutron
      resource: floating_ip
      check:
        status: DOWN
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	fipResults := results.FilterByService("neutron").FilterByResourceType("floating_ip")
	fipResults.LogSummary(t)
}

// TestNetwork_SecurityGroupAudit tests Neutron security group auditing
func TestNetwork_SecurityGroupAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-sg-unused
      description: Test unused security group check
      service: neutron
      resource: security_group
      check:
        unused: true
        exempt_names:
          - default
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	sgResults := results.FilterByService("neutron").FilterByResourceType("security_group")
	sgResults.LogSummary(t)
}

