package neutron

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/common"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/attributestags"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

type secGroupAdapter struct{ sg groups.SecGroup }

func (a secGroupAdapter) GetID() string        { return a.sg.ID }
func (a secGroupAdapter) GetName() string      { return a.sg.Name }
func (a secGroupAdapter) GetProjectID() string { return a.sg.TenantID }
func (a secGroupAdapter) GetStatus() string    { return "ACTIVE" }
func (a secGroupAdapter) GetCreatedAt() time.Time { return a.sg.CreatedAt }
func (a secGroupAdapter) GetUpdatedAt() time.Time { return a.sg.UpdatedAt }

// SecurityGroupAuditor audits neutron/security_group resources.
//
// Allowed checks: status, age_gt, unused, exempt_names
// Allowed actions: log, delete, tag
type SecurityGroupAuditor struct{}

func (a *SecurityGroupAuditor) ResourceType() string {
	return "security_group"
}

func (a *SecurityGroupAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	_ = ctx

	sg, ok := resource.(groups.SecGroup)
	if !ok {
		return nil, fmt.Errorf("expected groups.SecGroup, got %T", resource)
	}

	adapter := secGroupAdapter{sg: sg}
	result := common.BuildBaseResult(adapter, rule)

	exempt, err := common.RunCommonChecks(adapter, rule, result)
	if exempt || err != nil {
		return result, err
	}

	if rule.Check.Unused {
		result.Observation = "unused check pending - requires port enumeration"
	}

	return result, nil
}

func (a *SecurityGroupAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	_ = ctx

	// Log action doesn't require client or resource validation
	if rule.Action == "log" {
		return nil
	}

	c, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}

	sg, ok := resource.(groups.SecGroup)
	if !ok {
		return fmt.Errorf("expected groups.SecGroup, got %T", resource)
	}

	switch rule.Action {
	case "delete":
		// Check if the security group is in use by any ports
		portPages, err := ports.List(c, ports.ListOpts{}).AllPages()
		if err != nil {
			return fmt.Errorf("listing ports: %w", err)
		}
		portList, err := ports.ExtractPorts(portPages)
		if err != nil {
			return fmt.Errorf("extracting ports: %w", err)
		}

		// Check if any port references this security group
		for _, port := range portList {
			for _, sgID := range port.SecurityGroups {
				if sgID == sg.ID {
					return fmt.Errorf("cannot delete security group %s: in use by port %s", sg.ID, port.ID)
				}
			}
		}

		// Delete the security group
		if err := groups.Delete(c, sg.ID).ExtractErr(); err != nil {
			return fmt.Errorf("deleting security group %s: %w", sg.ID, err)
		}
		return nil

	case "tag":
		// Neutron security groups support tags via the standard-attr-tag extension
		tagName := rule.TagName
		if tagName == "" {
			tagName = rule.ActionTagName
		}
		if tagName == "" {
			return fmt.Errorf("neutron/security_group: tag action requires tag_name")
		}

		// Add the tag to the security group
		// The resource type for security groups in the tagging API is "security-groups"
		err := attributestags.Add(c, "security-groups", sg.ID, tagName).ExtractErr()
		if err != nil {
			return fmt.Errorf("tagging security group %s with %q: %w", sg.ID, tagName, err)
		}
		return nil

	default:
		return fmt.Errorf("neutron/security_group: action %q not implemented", rule.Action)
	}
}

// CheckUnused checks if a security group is unused (not attached to any ports).
// This is a helper function that can be called when port information is available.
func (a *SecurityGroupAuditor) CheckUnused(ctx context.Context, client *gophercloud.ServiceClient, sg groups.SecGroup) (bool, error) {
	_ = ctx

	portPages, err := ports.List(client, ports.ListOpts{}).AllPages()
	if err != nil {
		return false, fmt.Errorf("listing ports: %w", err)
	}
	portList, err := ports.ExtractPorts(portPages)
	if err != nil {
		return false, fmt.Errorf("extracting ports: %w", err)
	}

	// Check if any port references this security group
	for _, port := range portList {
		for _, sgID := range port.SecurityGroups {
			if sgID == sg.ID {
				return false, nil // In use
			}
		}
	}

	return true, nil // Unused
}
