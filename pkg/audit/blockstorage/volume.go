package blockstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
)

// VolumeAuditor audits Cinder volumes
type VolumeAuditor struct{}

func (a *VolumeAuditor) ResourceType() string {
	return "volume"
}

func (a *VolumeAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	volume, ok := resource.(volumes.Volume)
	if !ok {
		return nil, fmt.Errorf("expected volumes.Volume, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   volume.ID,
		ResourceName: volume.Name,
		// gophercloud volumes.Volume doesn't expose project/tenant ID in the v3 API struct.
		ProjectID:    "",
		Compliant:    true,
		Rule:         rule,
		Status:       volume.Status,
		UpdatedAt:    volume.UpdatedAt,
	}

	check := rule.Check
	now := time.Now()

	// Check status
	if check.Status != "" {
		if volume.Status != check.Status {
			return result, nil // Status doesn't match, but that's not a violation for this check
		}
	}

	// Check age_gt
	if check.AgeGT != "" {
		ageThreshold, err := check.ParseAgeGT()
		if err != nil {
			return nil, fmt.Errorf("failed to parse age_gt: %w", err)
		}

		evalTime := volume.UpdatedAt
		if evalTime.IsZero() {
			evalTime = volume.CreatedAt
		}
		if !evalTime.IsZero() {
			age := now.Sub(evalTime)
			if age >= ageThreshold {
				result.Compliant = false
				if result.Observation != "" {
					result.Observation += "; "
				}
				result.Observation = fmt.Sprintf("Volume is %s old (>= %s) and status is %s", age.Round(time.Hour*24), ageThreshold, volume.Status)
			}
		}
	}

	return result, nil
}

func (a *VolumeAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	// Volume violations are typically logged, not fixed
	// If delete action is specified, we would delete the volume here
	if rule.Action == "delete" {
		return fmt.Errorf("delete action for volume is not yet implemented (too dangerous)")
	}
	return nil
}

