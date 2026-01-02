package blockstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
)

// SnapshotAuditor audits Cinder snapshots
type SnapshotAuditor struct{}

func (a *SnapshotAuditor) ResourceType() string {
	return "snapshot"
}

func (a *SnapshotAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	snapshot, ok := resource.(snapshots.Snapshot)
	if !ok {
		return nil, fmt.Errorf("expected snapshots.Snapshot, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   snapshot.ID,
		ResourceName: snapshot.Name,
		// gophercloud snapshots.Snapshot doesn't expose project/tenant ID in the v3 API struct.
		ProjectID:    "",
		Compliant:    true,
		Rule:         rule,
		Status:       snapshot.Status,
		UpdatedAt:    snapshot.UpdatedAt,
	}

	check := rule.Check
	now := time.Now()

	// Check age_gt
	if check.AgeGT != "" {
		ageThreshold, err := check.ParseAgeGT()
		if err != nil {
			return nil, fmt.Errorf("failed to parse age_gt: %w", err)
		}

		evalTime := snapshot.UpdatedAt
		if evalTime.IsZero() {
			evalTime = snapshot.CreatedAt
		}
		if !evalTime.IsZero() {
			age := now.Sub(evalTime)
			if age >= ageThreshold {
				result.Compliant = false
				result.Observation = fmt.Sprintf("Snapshot is %s old (>= %s)", age.Round(time.Hour*24), ageThreshold)
			}
		}
	}

	return result, nil
}

func (a *SnapshotAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	if rule.Action != "delete" {
		return nil
	}

	snapshot, ok := resource.(snapshots.Snapshot)
	if !ok {
		return fmt.Errorf("expected snapshots.Snapshot, got %T", resource)
	}

	serviceClient, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	return snapshots.Delete(serviceClient, snapshot.ID).ExtractErr()
}

