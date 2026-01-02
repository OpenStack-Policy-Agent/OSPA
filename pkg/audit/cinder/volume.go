package cinder

import (
	"context"
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// VolumeAuditor audits cinder resources of type volume
//
// TODO(OSPA): Replace placeholder logic with real field extraction + rule evaluation for cinder/volume.
type VolumeAuditor struct{}

// ResourceType returns the resource type this auditor handles
func (a *VolumeAuditor) ResourceType() string {
	return "volume"
}

// Check evaluates a resource against a policy rule
func (a *VolumeAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx
	_ = resource

	// TODO(OSPA): Parse 'resource' into the correct OpenStack SDK type for cinder/volume.
	// Populate ResourceID/ResourceName/ProjectID/Status/UpdatedAt, and implement checks (status, age_gt, unused, etc.).
	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   "unknown",
		ResourceName: "unknown",
		ProjectID:    "",
		Compliant:    true,
		Rule:         rule,
		Status:       "",
	}

	return result, nil
}

// Fix applies remediation to a resource based on the rule action
func (a *VolumeAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx
	_ = client
	_ = resource

	// TODO(OSPA): Implement remediation actions using the correct OpenStack client calls:
	// - delete: delete the resource
	// - tag: apply policy tag/metadata
	// - log: no-op (already supported)
	switch rule.Action {
	case "log":
		return nil
	default:
		return fmt.Errorf("%s.%s fix action %q not implemented", "cinder", "volume", rule.Action)
	}
}
