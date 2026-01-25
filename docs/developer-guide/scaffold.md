# Scaffold Tool

The scaffold tool generates boilerplate code for adding new OpenStack services to OSPA.

## Overview

The scaffold tool automates the creation of:

- Service implementation files
- Discovery files
- Auditor files
- Auth client methods
- Unit tests
- E2E tests
- Policy validators
- Policy guides

## Usage

### Basic Usage

```bash
go run ./cmd/scaffold --service <name> --resources <list>
```

### List Available Services

```bash
go run ./cmd/scaffold --list
```

This shows all known OpenStack services and their resources.

### Generate a Service

```bash
go run ./cmd/scaffold \
  --service glance \
  --resources image,member \
```

### Using Make

```bash
make scaffold SERVICE=glance RESOURCES=image,member
```

## Options

| Flag | Required | Description |
|------|----------|-------------|
| `--service` | Yes | Service name (e.g., `nova`, `neutron`) |
| `--resources` | Yes | Comma-separated resource types |
| `--list` | No | List available services |

## Generated Files

```
Generated files for 'glance' with resources [image, member]:

pkg/services/services/glance.go       # Service implementation
pkg/discovery/services/glance.go      # Resource discoverers
pkg/audit/glance/image.go             # Image auditor
pkg/audit/glance/image_test.go        # Image auditor tests
pkg/audit/glance/member.go            # Member auditor
pkg/audit/glance/member_test.go       # Member auditor tests
pkg/auth/auth.go                      # Client method (appended)
pkg/policy/validation/glance.go       # Policy validator
e2e/glance/resource_creator.go        # E2E resource creators
e2e/glance/image_test.go              # E2E image tests
e2e/glance/member_test.go             # E2E member tests
examples/policies/glance-policy-guide.md  # Policy documentation
```

## After Generation

### 1. Update Registry Config

Edit `cmd/scaffold/internal/registry/config/<service>.yaml`:

```yaml
resources:
  image:
    description: Glance images
    checks:
      - status      # Has status field
      - age_gt      # Has timestamps
      - unused      # Can detect unused
      - exempt_names
    actions:
      - log
      - delete
      - tag         # If supported
```

### 2. Implement Discoverers

Edit `pkg/discovery/services/<service>.go`:

```go
func (d *GlanceImageDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
    ch := make(chan discovery.Job)

    go func() {
        defer close(ch)

        opts := images.ListOpts{}
        pager := images.List(client, opts)
        
        pager.EachPage(func(page pagination.Page) (bool, error) {
            select {
            case <-ctx.Done():
                return false, ctx.Err()
            default:
            }

            imageList, err := images.ExtractImages(page)
            if err != nil {
                return false, err
            }

            for _, image := range imageList {
                ch <- discovery.Job{
                    Service:      "glance",
                    ResourceType: "image",
                    ResourceID:   image.ID,
                    ProjectID:    image.Owner,
                    Resource:     image,
                }
            }
            return true, nil
        })
    }()

    return ch, nil
}
```

### 3. Implement Auditors

Edit `pkg/audit/<service>/<resource>.go`:

```go
func (a *ImageAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
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

    check := rule.Check

    // Status check
    if check.Status != "" && string(image.Status) == check.Status {
        result.Compliant = false
        result.Observation = fmt.Sprintf("status is %s", image.Status)
        return result, nil
    }

    // Age check
    if check.AgeGT != "" {
        threshold, err := check.ParseAgeGT()
        if err != nil {
            return nil, err
        }
        if time.Since(image.UpdatedAt) > threshold {
            result.Compliant = false
            result.Observation = fmt.Sprintf("image is older than %s", check.AgeGT)
        }
    }

    return result, nil
}
```

### 4. Implement Validators

Edit `pkg/policy/validation/<service>.go`:

```go
func (v *GlanceValidator) ValidateResource(resourceType, ruleName string, check *policy.CheckConditions) error {
    switch resourceType {
    case "image":
        if check.Status == "" && check.AgeGT == "" {
            return fmt.Errorf("rule %q: must specify status or age_gt", ruleName)
        }
    default:
        return fmt.Errorf("unknown resource type: %s", resourceType)
    }
    return nil
}
```

### 5. Run Tests

```bash
# Unit tests
go test ./pkg/audit/glance/... -v

# Service registration
go test ./pkg/services/... -v

# E2E tests (requires OpenStack)
OS_CLOUD=mycloud go test -tags=e2e ./e2e/glance/... -v
```

## Architecture

```
cmd/scaffold/
├── main.go                     # CLI entry point
└── internal/
    ├── registry/
    │   ├── registry.go         # Service/resource registry
    │   └── config/             # YAML metadata per service
    │       ├── nova.yaml
    │       ├── neutron.yaml
    │       └── ...
    └── generators/
        ├── orchestrator.go     # Main orchestration
        ├── service.go          # Service file generation
        ├── discovery.go        # Discovery file generation
        ├── auditor.go          # Auditor files generation
        ├── auth.go             # Auth method generation
        ├── tests.go            # Unit tests generation
        ├── e2e.go              # E2E tests generation
        ├── validation.go       # Validation file generation
        └── policy_guide.go     # Policy guide generation
```

## Registry Configuration

Each service has a YAML config in `cmd/scaffold/internal/registry/config/`:

```yaml
name: glance
display_name: Image
client_type: image
description: OpenStack Image Service
resources:
  image:
    description: Glance images
    checks:
      - status
      - age_gt
      - unused
      - exempt_names
    actions:
      - log
      - delete
      - tag
  member:
    description: Image sharing members
    checks:
      - status
    actions:
      - log
      - delete
```

## Best Practices

1. **Always use scaffold** - Ensures consistency across services
2. **Review generated code** - Scaffold provides templates, not finished code
3. **Update registry first** - Define accurate checks/actions before generating
4. **Test incrementally** - Unit tests → integration → E2E
5. **Document the service** - Update the generated policy guide

## Regenerating

Regenerating will **overwrite** existing files. To preserve customizations:

1. Backup modified files
2. Run scaffold
3. Merge changes manually

Or use a feature branch:

```bash
git checkout -b regenerate-glance
go run ./cmd/scaffold --service glance --resources image,member
git diff  # Review changes
```

## Troubleshooting

### Service Not in List

Add it to `cmd/scaffold/internal/registry/config/<service>.yaml`.

### Generated Code Doesn't Compile

1. Check import paths
2. Verify Gophercloud package names
3. Review generated TODO comments

### Missing Client Method

The scaffold appends to `pkg/auth/auth.go`. Check if it was added correctly.

