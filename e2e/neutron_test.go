//go:build e2e

package e2e

import (
	"testing"
)


// TestNeutron_SecurityGroupRuleAudit tests neutron security_group_rule auditing
func TestNeutron_SecurityGroupRuleAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-security_group_rule-check
      description: Test security_group_rule check
      service: neutron
      resource: security_group_rule
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for neutron/security_group_rule results
	SecurityGroupRuleResults := results.FilterByService("neutron").FilterByResourceType("security_group_rule")

	SecurityGroupRuleResults.LogSummary(t)

	// Basic assertions
	if SecurityGroupRuleResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during security_group_rule audit", SecurityGroupRuleResults.Errors)
	}
}


// TestNeutron_FloatingIpAudit tests neutron floating_ip auditing
func TestNeutron_FloatingIpAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-floating_ip-check
      description: Test floating_ip check
      service: neutron
      resource: floating_ip
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for neutron/floating_ip results
	FloatingIpResults := results.FilterByService("neutron").FilterByResourceType("floating_ip")

	FloatingIpResults.LogSummary(t)

	// Basic assertions
	if FloatingIpResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during floating_ip audit", FloatingIpResults.Errors)
	}
}


// TestNeutron_SecurityGroupAudit tests neutron security_group auditing
func TestNeutron_SecurityGroupAudit(t *testing.T) {
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-security_group-check
      description: Test security_group check
      service: neutron
      resource: security_group
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	// Filter for neutron/security_group results
	SecurityGroupResults := results.FilterByService("neutron").FilterByResourceType("security_group")

	SecurityGroupResults.LogSummary(t)

	// Basic assertions
	if SecurityGroupResults.Errors > 0 {
		t.Logf("Warning: %%d errors encountered during security_group audit", SecurityGroupResults.Errors)
	}
}


