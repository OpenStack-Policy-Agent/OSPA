# Adding Services

This guide covers how to add support for new OpenStack services to OSPA.

## Overview

Adding a new service involves:

1. Generating scaffold code
2. Implementing discoverers
3. Implementing auditors
4. Adding validators
5. Writing tests

## Quick Start

The fastest way to add a service is using the scaffold tool:

```bash
go run ./cmd/scaffold \
  --service glance \
  --resources image,member
```

See [Scaffold Tool](scaffold.md) for details.

## Manual Implementation

If you prefer to create files manually, follow these steps:

### Step 1: Service Implementation

Create `pkg/services/services/<service>.go`:

```go
package services

import (
    "fmt"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
    "github.com/OpenStack-Policy-Agent/OSPA/pkg/audit/glance"
    "github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
    "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
    discovery_services "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services"
    rootservices "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"
    "github.com/gophercloud/gophercloud"
)

// GlanceService implements the Service interface for OpenStack Glance.
//
// Supported resources:
//   - image: Glance images
//     Checks: status, age_gt, unused, exempt_names
//     Actions: log, delete, tag
//   - member: Image members
//     Checks: status, exempt_names
//     Actions: log, delete
type GlanceService struct{}

func init() {
    rootservices.MustRegister(&GlanceService{})
    rootservices.RegisterResource("glance", "image")
    rootservices.RegisterResource("glance", "member")
}

func (s *GlanceService) Name() string {
    return "glance"
}

func (s *GlanceService) GetClient(session *auth.Session) (*gophercloud.ServiceClient, error) {
    return session.GetGlanceClient()
}

func (s *GlanceService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
    switch resourceType {
    case "image":
        return &glance.ImageAuditor{}, nil
    case "member":
        return &glance.MemberAuditor{}, nil
    default:
        return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
    }
}

func (s *GlanceService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
    switch resourceType {
    case "image":
        return &discovery_services.GlanceImageDiscoverer{}, nil
    case "member":
        return &discovery_services.GlanceMemberDiscoverer{}, nil
    default:
        return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
    }
}
```

### Step 2: Auth Client Method

Add to `pkg/auth/auth.go`:

```go
func (s *Session) GetGlanceClient() (*gophercloud.ServiceClient, error) {
    client, err := clientconfig.NewServiceClient("image", &clientconfig.ClientOpts{
        Cloud: s.CloudName,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create glance client: %w", err)
    }
    return client, nil
}
```

### Step 3: Discovery Implementation

Create `pkg/discovery/services/glance.go`:

```go
package services

import (
    "context"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
    "github.com/gophercloud/gophercloud"
    "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

// GlanceImageDiscoverer discovers glance/image resources.
type GlanceImageDiscoverer struct{}

func (d *GlanceImageDiscoverer) ResourceType() string {
    return "image"
}

func (d *GlanceImageDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
    ch := make(chan discovery.Job)

    go func() {
        defer close(ch)

        opts := images.ListOpts{}
        pages, err := images.List(client, opts).AllPages()
        if err != nil {
            return
        }

        imageList, err := images.ExtractImages(pages)
        if err != nil {
            return
        }

        for _, image := range imageList {
            select {
            case <-ctx.Done():
                return
            case ch <- discovery.Job{
                Service:      "glance",
                ResourceType: "image",
                ResourceID:   image.ID,
                ProjectID:    image.Owner,
                Resource:     image,
            }:
            }
        }
    }()

    return ch, nil
}
```

### Step 4: Auditor Implementation

Create `pkg/audit/glance/image.go`:

```go
package glance

import (
    "context"
    "fmt"
    "path/filepath"
    "time"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
    "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
    "github.com/gophercloud/gophercloud"
    "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

type ImageAuditor struct{}

func (a *ImageAuditor) ResourceType() string {
    return "image"
}

func (a *ImageAuditor) ImplementedChecks() []string {
    return []string{"status", "age_gt", "unused", "exempt_names"}
}

func (a *ImageAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
    _ = ctx

    image, ok := resource.(images.Image)
    if !ok {
        return nil, fmt.Errorf("expected images.Image, got %T", resource)
    }

    result := &audit.Result{
        RuleID:       rule.Name,
        ResourceID:   image.ID,
        ResourceName: image.Name,
        ProjectID:    image.Owner,
        Status:       string(image.Status),
        UpdatedAt:    image.UpdatedAt,
        Compliant:    true,
        Rule:         rule,
    }

    // Check exemptions first
    if isExemptByName(image.Name, rule.Check.ExemptNames) {
        result.Compliant = true
        result.Observation = "exempt by name"
        return result, nil
    }

    // Status check
    if rule.Check.Status != "" && string(image.Status) == rule.Check.Status {
        result.Compliant = false
        result.Observation = fmt.Sprintf("status is %s", image.Status)
        return result, nil
    }

    // Age check
    if rule.Check.AgeGT != "" {
        threshold, err := rule.Check.ParseAgeGT()
        if err != nil {
            return nil, fmt.Errorf("failed to parse age_gt: %w", err)
        }
        age := time.Since(image.UpdatedAt)
        if age > threshold {
            result.Compliant = false
            result.Observation = fmt.Sprintf("image is %s old (threshold: %s)", age, threshold)
        }
    }

    return result, nil
}

func (a *ImageAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
    _ = ctx

    if rule.Action == "log" {
        return nil
    }

    c, ok := client.(*gophercloud.ServiceClient)
    if !ok {
        return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
    }

    image, ok := resource.(images.Image)
    if !ok {
        return fmt.Errorf("expected images.Image, got %T", resource)
    }

    switch rule.Action {
    case "delete":
        return images.Delete(c, image.ID).ExtractErr()
    case "tag":
        // Implement tagging if supported
        return fmt.Errorf("glance/image: tag action not implemented")
    default:
        return fmt.Errorf("glance/image: action %q not implemented", rule.Action)
    }
}

func isExemptByName(name string, patterns []string) bool {
    for _, pattern := range patterns {
        if matched, _ := filepath.Match(pattern, name); matched {
            return true
        }
        if pattern == name {
            return true
        }
    }
    return false
}
```

### Step 5: Validator

Create `pkg/policy/validation/glance.go`:

```go
package validation

import (
    "fmt"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// GlanceValidator validates Glance service policies.
type GlanceValidator struct{}

func init() {
    policy.RegisterValidator(&GlanceValidator{})
}

func (v *GlanceValidator) ServiceName() string {
    return "glance"
}

func (v *GlanceValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
    switch resourceType {
    case "image":
        if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names"}); err != nil {
            return fmt.Errorf("rule %q: %w", ruleName, err)
        }
    case "member":
        if err := validateAllowedChecks(check, []string{"status", "exempt_names"}); err != nil {
            return fmt.Errorf("rule %q: %w", ruleName, err)
        }
    default:
        return fmt.Errorf("rule %q: unsupported resource type %q for glance service", ruleName, resourceType)
    }
    return nil
}
```

## Testing

### Unit Tests

Create `pkg/audit/glance/image_test.go`:

```go
package glance

import (
    "context"
    "testing"
    "time"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
    "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

func TestImageAuditor_ResourceType(t *testing.T) {
    a := &ImageAuditor{}
    if got := a.ResourceType(); got != "image" {
        t.Errorf("ResourceType() = %q, want %q", got, "image")
    }
}

func TestImageAuditor_Check_StatusMatch(t *testing.T) {
    a := &ImageAuditor{}

    image := images.Image{
        ID:        "img-1",
        Name:      "test-image",
        Status:    images.ImageStatusActive,
        Owner:     "project-1",
        UpdatedAt: time.Now(),
    }

    rule := &policy.Rule{
        Name:   "test-rule",
        Action: "log",
        Check: policy.CheckConditions{
            Status: "active",
        },
    }

    result, err := a.Check(context.Background(), image, rule)
    if err != nil {
        t.Fatalf("Check() error = %v", err)
    }

    if result.Compliant {
        t.Error("expected non-compliant for matching status")
    }
}
```

### Run Tests

```bash
# Unit tests
go test ./pkg/audit/glance/... -v

# Service registration
go test ./pkg/services/... -v

# All tests
go test ./...
```

### E2E Tests

See [Testing](testing.md) for E2E test guidelines.

## Checklist

Use this checklist when adding a new service:

- [ ] Service implementation created in `pkg/services/services/<service>.go`
- [ ] Auth client method added to `pkg/auth/auth.go`
- [ ] Discoverers implemented in `pkg/discovery/services/<service>.go`
- [ ] Auditors implemented in `pkg/audit/<service>/<resource>.go`
- [ ] Auditor declares checks via `ImplementedChecks()`
- [ ] Validator created in `pkg/policy/validation/<service>.go`
- [ ] Resources registered in `init()` using `rootservices.RegisterResource()`
- [ ] Unit tests written
- [ ] Unit tests pass
- [ ] E2E tests written
- [ ] E2E tests pass
- [ ] Documentation updated

