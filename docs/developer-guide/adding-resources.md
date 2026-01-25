# Adding Resources

This guide covers how to add new resource types to existing OpenStack services.

## Overview

Adding a resource to an existing service involves:

1. Creating a discoverer
2. Creating an auditor
3. Registering the resource
4. Updating the validator
5. Writing tests

## Example: Adding "backup" to Cinder

Let's walk through adding `backup` support to the existing Cinder service.

### Step 1: Create Discoverer

Add to `pkg/discovery/services/cinder.go`:

```go
import (
    "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/backups"
)

// CinderBackupDiscoverer discovers cinder/backup resources.
type CinderBackupDiscoverer struct{}

func (d *CinderBackupDiscoverer) ResourceType() string {
    return "backup"
}

func (d *CinderBackupDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
    ch := make(chan discovery.Job)

    go func() {
        defer close(ch)

        opts := backups.ListOpts{}
        if allTenants {
            opts.AllTenants = true
        }

        pages, err := backups.List(client, opts).AllPages()
        if err != nil {
            return
        }

        backupList, err := backups.ExtractBackups(pages)
        if err != nil {
            return
        }

        for _, backup := range backupList {
            select {
            case <-ctx.Done():
                return
            case ch <- discovery.Job{
                Service:      "cinder",
                ResourceType: "backup",
                ResourceID:   backup.ID,
                ProjectID:    backup.ProjectID,
                Resource:     backup,
            }:
            }
        }
    }()

    return ch, nil
}
```

### Step 2: Create Auditor

Create `pkg/audit/cinder/backup.go`:

```go
package cinder

import (
    "context"
    "fmt"
    "path/filepath"
    "time"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
    "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
    "github.com/gophercloud/gophercloud"
    "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/backups"
)

type BackupAuditor struct{}

func (a *BackupAuditor) ResourceType() string {
    return "backup"
}

func (a *BackupAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
    _ = ctx

    backup, ok := resource.(backups.Backup)
    if !ok {
        return nil, fmt.Errorf("expected backups.Backup, got %T", resource)
    }

    result := &audit.Result{
        RuleID:       rule.Name,
        ResourceID:   backup.ID,
        ResourceName: backup.Name,
        ProjectID:    backup.ProjectID,
        Status:       backup.Status,
        Compliant:    true,
        Rule:         rule,
    }

    // Check exemptions first
    if isExemptByName(backup.Name, rule.Check.ExemptNames) {
        result.Compliant = true
        result.Observation = "exempt by name"
        return result, nil
    }

    // Status check
    if rule.Check.Status != "" && backup.Status == rule.Check.Status {
        result.Compliant = false
        result.Observation = fmt.Sprintf("status is %s", backup.Status)
        return result, nil
    }

    // Age check
    if rule.Check.AgeGT != "" {
        threshold, err := rule.Check.ParseAgeGT()
        if err != nil {
            return nil, fmt.Errorf("failed to parse age_gt: %w", err)
        }
        age := time.Since(backup.CreatedAt)
        if age > threshold {
            result.Compliant = false
            result.Observation = fmt.Sprintf("backup is %s old (threshold: %s)", age, threshold)
        }
    }

    return result, nil
}

func (a *BackupAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
    _ = ctx

    if rule.Action == "log" {
        return nil
    }

    c, ok := client.(*gophercloud.ServiceClient)
    if !ok {
        return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
    }

    backup, ok := resource.(backups.Backup)
    if !ok {
        return fmt.Errorf("expected backups.Backup, got %T", resource)
    }

    switch rule.Action {
    case "delete":
        return backups.Delete(c, backup.ID).ExtractErr()
    default:
        return fmt.Errorf("cinder/backup: action %q not implemented", rule.Action)
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

### Step 3: Update Service

Update `pkg/services/services/cinder.go`:

```go
func (s *CinderService) GetResourceAuditor(resourceType string) (audit.Auditor, error) {
    switch resourceType {
    case "volume":
        return &cinder.VolumeAuditor{}, nil
    case "snapshot":
        return &cinder.SnapshotAuditor{}, nil
    case "backup":  // Add this
        return &cinder.BackupAuditor{}, nil
    default:
        return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
    }
}

func (s *CinderService) GetResourceDiscoverer(resourceType string) (discovery.Discoverer, error) {
    switch resourceType {
    case "volume":
        return &discovery_services.CinderVolumeDiscoverer{}, nil
    case "snapshot":
        return &discovery_services.CinderSnapshotDiscoverer{}, nil
    case "backup":  // Add this
        return &discovery_services.CinderBackupDiscoverer{}, nil
    default:
        return nil, fmt.Errorf("unsupported resource type %q for service %q", resourceType, s.Name())
    }
}

func init() {
    rootservices.MustRegister(&CinderService{})
    rootservices.RegisterResource("cinder", "volume")
    rootservices.RegisterResource("cinder", "snapshot")
    rootservices.RegisterResource("cinder", "backup")  // Add this
}
```

### Step 4: Update Validator

Update `pkg/policy/validation/cinder.go`:

```go
func (v *CinderValidator) ValidateResource(check *policy.CheckConditions, resourceType, ruleName string) error {
    switch resourceType {
    case "volume":
        if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names"}); err != nil {
            return fmt.Errorf("rule %q: %w", ruleName, err)
        }
    case "snapshot":
        if err := validateAllowedChecks(check, []string{"status", "age_gt", "unused", "exempt_names"}); err != nil {
            return fmt.Errorf("rule %q: %w", ruleName, err)
        }
    case "backup":  // Add this
        if err := validateAllowedChecks(check, []string{"status", "age_gt", "exempt_names"}); err != nil {
            return fmt.Errorf("rule %q: %w", ruleName, err)
        }
    default:
        return fmt.Errorf("rule %q: unsupported resource type %q for cinder service", ruleName, resourceType)
    }
    return nil
}
```

### Step 5: Write Unit Tests

Create `pkg/audit/cinder/backup_test.go`:

```go
package cinder

import (
    "context"
    "testing"
    "time"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
    "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/backups"
)

func TestBackupAuditor_ResourceType(t *testing.T) {
    a := &BackupAuditor{}
    if got := a.ResourceType(); got != "backup" {
        t.Errorf("ResourceType() = %q, want %q", got, "backup")
    }
}

func TestBackupAuditor_Check_StatusMatch(t *testing.T) {
    a := &BackupAuditor{}

    backup := backups.Backup{
        ID:        "backup-1",
        Name:      "test-backup",
        Status:    "available",
        ProjectID: "project-1",
        CreatedAt: time.Now(),
    }

    rule := &policy.Rule{
        Name:   "test-rule",
        Action: "log",
        Check: policy.CheckConditions{
            Status: "available",
        },
    }

    result, err := a.Check(context.Background(), backup, rule)
    if err != nil {
        t.Fatalf("Check() error = %v", err)
    }

    if result.Compliant {
        t.Error("expected non-compliant for matching status")
    }
}

func TestBackupAuditor_Check_AgeGT(t *testing.T) {
    a := &BackupAuditor{}

    oldTime := time.Now().Add(-31 * 24 * time.Hour) // 31 days ago
    backup := backups.Backup{
        ID:        "backup-1",
        Name:      "old-backup",
        Status:    "available",
        ProjectID: "project-1",
        CreatedAt: oldTime,
    }

    rule := &policy.Rule{
        Name:   "test-rule",
        Action: "log",
        Check: policy.CheckConditions{
            AgeGT: "30d",
        },
    }

    result, err := a.Check(context.Background(), backup, rule)
    if err != nil {
        t.Fatalf("Check() error = %v", err)
    }

    if result.Compliant {
        t.Error("expected non-compliant for old backup")
    }
}

func TestBackupAuditor_Check_Exempt(t *testing.T) {
    a := &BackupAuditor{}

    backup := backups.Backup{
        ID:        "backup-1",
        Name:      "system-backup",
        Status:    "available",
        ProjectID: "project-1",
        CreatedAt: time.Now(),
    }

    rule := &policy.Rule{
        Name:   "test-rule",
        Action: "log",
        Check: policy.CheckConditions{
            Status:      "available",
            ExemptNames: []string{"system-*"},
        },
    }

    result, err := a.Check(context.Background(), backup, rule)
    if err != nil {
        t.Fatalf("Check() error = %v", err)
    }

    if !result.Compliant {
        t.Error("expected compliant for exempt backup")
    }
}
```

### Step 6: Test

```bash
# Run unit tests
go test ./pkg/audit/cinder/... -v

# Verify service registration
go test ./pkg/services/... -v

# Test with agent
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy test-backup-policy.yaml \
  --out findings.json
```

## Resource Types Reference

When adding resources, consider these common patterns:

### Status-Based Resources

Most resources have a status field:

```go
if rule.Check.Status != "" && resource.Status == rule.Check.Status {
    result.Compliant = false
    result.Observation = fmt.Sprintf("status is %s", resource.Status)
}
```

### Time-Based Resources

Resources with timestamps support age checks:

```go
if rule.Check.AgeGT != "" {
    threshold, _ := rule.Check.ParseAgeGT()
    if time.Since(resource.UpdatedAt) > threshold {
        result.Compliant = false
    }
}
```

### Relationship-Based Resources

Some resources have "unused" detection:

```go
if rule.Check.Unused {
    // Query related resources to check if this one is in use
    ports, _ := listPortsUsingSecurityGroup(client, sg.ID)
    if len(ports) == 0 {
        result.Compliant = false
        result.Observation = "security group is not in use"
    }
}
```

## Naming Conventions

Follow these naming conventions for consistency:

| Component | Pattern | Example |
|-----------|---------|---------|
| Discoverer | `<Service><Resource>Discoverer` | `CinderBackupDiscoverer` |
| Auditor | `<Resource>Auditor` | `BackupAuditor` |
| Auditor Package | `pkg/audit/<service>` | `pkg/audit/cinder` |
| Discoverer File | `pkg/discovery/services/<service>.go` | `pkg/discovery/services/cinder.go` |

## Checklist

Use this checklist when adding a resource:

- [ ] Discoverer created in `pkg/discovery/services/<service>.go`
- [ ] Auditor created in `pkg/audit/<service>/<resource>.go`
- [ ] Auditor `Check()` method implemented
- [ ] Auditor `Fix()` method implemented
- [ ] Service `GetResourceAuditor` updated
- [ ] Service `GetResourceDiscoverer` updated
- [ ] Resource registered in `init()` with `rootservices.RegisterResource()`
- [ ] Validator updated in `pkg/policy/validation/<service>.go`
- [ ] Unit tests written
- [ ] Unit tests pass
- [ ] E2E tests written (if applicable)
- [ ] Documentation updated

## Common Patterns

### Gophercloud Pagination

```go
pages, err := resources.List(client, opts).AllPages()
if err != nil {
    return
}

items, err := resources.ExtractItems(pages)
if err != nil {
    return
}

for _, item := range items {
    select {
    case <-ctx.Done():
        return
    case ch <- discovery.Job{
        Service:      "service",
        ResourceType: "resource",
        ResourceID:   item.ID,
        ProjectID:    item.ProjectID,
        Resource:     item,
    }:
    }
}
```

### Context Cancellation

Always check for context cancellation in loops:

```go
for _, item := range items {
    select {
    case <-ctx.Done():
        return
    case ch <- discovery.Job{...}:
    }
}
```

### Type Assertions

Always check type assertions:

```go
resource, ok := rawResource.(ResourceType)
if !ok {
    return nil, fmt.Errorf("expected ResourceType, got %T", rawResource)
}
```

