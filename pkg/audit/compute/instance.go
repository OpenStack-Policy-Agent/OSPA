package compute

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

// InstanceAuditor audits Nova instances
type InstanceAuditor struct{}

func (a *InstanceAuditor) ResourceType() string {
	return "instance"
}

func (a *InstanceAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	server, ok := resource.(servers.Server)
	if !ok {
		return nil, fmt.Errorf("expected servers.Server, got %T", resource)
	}

	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   server.ID,
		ResourceName: server.Name,
		ProjectID:    server.TenantID,
		Compliant:    true,
		Rule:         rule,
		Status:       server.Status,
		UpdatedAt:    server.Updated,
	}

	check := rule.Check
	now := time.Now()

	// Check age_gt
	if check.AgeGT != "" {
		ageThreshold, err := check.ParseAgeGT()
		if err != nil {
			return nil, fmt.Errorf("failed to parse age_gt: %w", err)
		}

		evalTime := server.Updated
		if evalTime.IsZero() {
			evalTime = server.Created
		}
		if !evalTime.IsZero() {
			age := now.Sub(evalTime)
			if age >= ageThreshold {
				// Check exemptions
				if check.ExemptMetadata != nil {
					if exempt, _ := checkMetadataExemption(server.Metadata, check.ExemptMetadata); exempt {
						return result, nil // Exempt, so compliant
					}
				}

				result.Compliant = false
				result.Observation = fmt.Sprintf("Instance is %s old (>= %s)", age.Round(time.Hour*24), ageThreshold)
			}
		}
	}

	// Check image_name
	if len(check.ImageName) > 0 {
		imageName := ""
		if server.Image != nil {
			imageName = server.Image["name"].(string)
		}

		for _, deprecatedImage := range check.ImageName {
			if strings.EqualFold(imageName, deprecatedImage) {
				result.Compliant = false
				if result.Observation != "" {
					result.Observation += "; "
				}
				result.Observation += fmt.Sprintf("Instance is using deprecated image: %s", imageName)
				break
			}
		}
	}

	// Check status
	if check.Status != "" {
		if server.Status != check.Status {
			return result, nil // Status doesn't match, but that's not a violation for this check
		}
	}

	return result, nil
}

func (a *InstanceAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	if rule.Action != "tag" {
		return nil
	}

	server, ok := resource.(servers.Server)
	if !ok {
		return fmt.Errorf("expected servers.Server, got %T", resource)
	}

	serviceClient, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	// Tag the instance
	tagName := rule.TagName
	if tagName == "" {
		tagName = rule.ActionTagName
	}
	if tagName == "" {
		return fmt.Errorf("tag_name or action_tag_name is required for tag action")
	}

	// Add metadata tag
	metadata := server.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata[tagName] = "true"

	_, err := servers.UpdateMetadata(serviceClient, server.ID, servers.MetadataOpts(metadata)).Extract()
	return err
}

// checkMetadataExemption checks if metadata matches exemption criteria
func checkMetadataExemption(metadata map[string]string, exempt *policy.MetadataMatch) (bool, error) {
	if metadata == nil {
		return false, nil
	}

	value, exists := metadata[exempt.Key]
	if !exists {
		return false, nil
	}

	return strings.EqualFold(value, exempt.Value), nil
}

