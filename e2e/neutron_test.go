//go:build e2e

package e2e

import (
	"testing"
)


func TestNeutron_NetworkAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
	engine := NewTestEngine(t)

	policyYAML := `version: v1
defaults:
  workers: 2
policies:
  - neutron:
    - name: test-network-check
      description: Test network check
      service: neutron
      resource: network
      check:
        status: active
      action: log`

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	NetworkResults := results.FilterByService("neutron").FilterByResourceType("network")
	NetworkResults.LogSummary(t)

	if NetworkResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during network audit", NetworkResults.Errors)
	}
}


func TestNeutron_SecurityGroupAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
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

	SecurityGroupResults := results.FilterByService("neutron").FilterByResourceType("security_group")
	SecurityGroupResults.LogSummary(t)

	if SecurityGroupResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during security_group audit", SecurityGroupResults.Errors)
	}
}


func TestNeutron_SecurityGroupRuleAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
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

	SecurityGroupRuleResults := results.FilterByService("neutron").FilterByResourceType("security_group_rule")
	SecurityGroupRuleResults.LogSummary(t)

	if SecurityGroupRuleResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during security_group_rule audit", SecurityGroupRuleResults.Errors)
	}
}


func TestNeutron_FloatingIpAudit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
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

	FloatingIpResults := results.FilterByService("neutron").FilterByResourceType("floating_ip")
	FloatingIpResults.LogSummary(t)

	if FloatingIpResults.Errors > 0 {
		t.Logf("Warning: %d errors encountered during floating_ip audit", FloatingIpResults.Errors)
	}
}


