package cinder

import (
	"context"
	"fmt"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	// TODO: Import the gophercloud resource type for snapshot.
	// Example: "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

// SnapshotAuditor audits cinder/snapshot resources.
//
// Allowed checks: status, age_gt, unused, exempt_names
// Allowed actions: log, delete, tag
//
// TODO: Cast 'resource' to the correct gophercloud type and implement checks.
// Gophercloud docs: https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack
// OpenStack API: https://docs.openstack.org/api-ref/cinder
type SnapshotAuditor struct{}

func (a *SnapshotAuditor) ResourceType() string {
	return "snapshot"
}

func (a *SnapshotAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	// TODO: Cast resource to the correct type.
	// Example: r := resource.(servers.Server)
	//
	// Then populate the result:
	//   result.ResourceID = r.ID
	//   result.ResourceName = r.Name
	//   result.ProjectID = r.TenantID
	//   result.Status = r.Status
	//   result.UpdatedAt = r.Updated
	//
	// Implement checks based on rule.Check fields:
	//   - Status: compare r.Status with rule.Check.Status
	//   - AgeGT: compare time.Since(r.Updated) with rule.Check.AgeGT
	//   - Unused: implement resource-specific unused detection
	//   - ExemptNames: skip if r.Name matches any exempt pattern

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   "unknown",
		ResourceName: "unknown",
		ProjectID:    "",
		Compliant:    true,
		Rule:         rule,
		Status:       "",
	}

	_ = resource
	return result, nil
}

func (a *SnapshotAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx
	_ = client
	_ = resource

	// TODO: Implement remediation actions.
	// Cast client to *gophercloud.ServiceClient.
	// Allowed actions: log, delete, tag
	//
	// Example for delete:
	//   c := client.(*gophercloud.ServiceClient)
	//   r := resource.(servers.Server)
	//   return servers.Delete(c, r.ID).ExtractErr()

	switch rule.Action {
	case "log":
		return nil
	default:
		return fmt.Errorf("%s/%s: action %q not implemented", "cinder", "snapshot", rule.Action)
	}
}
