# Developer Guide (OSPA)

This comprehensive guide covers everything you need to know about developing OSPA, from initial setup to adding new services and resources.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Development Setup](#development-setup)
3. [Quick Start: Adding a New Service](#quick-start-adding-a-new-service)
4. [Complete Workflow Example](#complete-workflow-example)
5. [Adding a New Service (Detailed)](#adding-a-new-service-detailed)
6. [Adding a Resource to Existing Service](#adding-a-resource-to-existing-service)
7. [Testing](#testing)
8. [Running the Agent](#running-the-agent)
9. [Troubleshooting](#troubleshooting)
10. [Development Best Practices](#development-best-practices)
11. [Common Development Patterns](#common-development-patterns)
12. [Testing Checklist](#testing-checklist)

## Prerequisites

- **Go**: Version matching `go.mod` (check with `go version`)
- **OpenStack Access**: Access to an OpenStack cloud with a valid `clouds.yaml` file
- **Environment Variables**:
  ```bash
  export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml
  export OS_CLOUD=mycloud  # Name of cloud in clouds.yaml
  ```

## Development Setup

### 1. Clone and Build

```bash
git clone <repository-url>
cd OSPA
go mod download
go build ./cmd/agent
```

### 2. Verify Setup

```bash
# Run unit tests
go test ./...

# Verify scaffold tool works
go run ./cmd/scaffold --help
```

### 3. Test Agent (Optional)

```bash
# Test with a minimal policy
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/policies.yaml --out /dev/null
```

## Quick Start: Adding a New Service

The easiest way to add a new service is to use the scaffold tool:

**First, list available services:**
```bash
go run ./cmd/scaffold --list
```

**Using Make:**
```bash
make scaffold SERVICE=glance RESOURCES=image,member DISPLAY_NAME=Glance TYPE=image
```

**Or directly:**
```bash
go run ./cmd/scaffold --service glance --display-name Glance --resources image,member --type image
```

**Note:** The scaffold tool validates that services and resources exist in OpenStack. If you specify an invalid service or resource, it will show an error with available options. Use `--list` to see all supported services and resources.

This generates all the necessary files automatically. See [`scaffold-README.md`](scaffold-README.md) for details.

### Manual Checklist

If you prefer to create files manually:

- [ ] Create service implementation (`pkg/services/services/<servicename>.go`)
- [ ] Add client method to auth (`pkg/auth/auth.go`)
- [ ] Create discovery file (`pkg/discovery/services/<servicename>.go`)
- [ ] Create audit directory (`pkg/audit/<servicename>/`)
- [ ] Create auditor for each resource type
- [ ] Create validation file (`pkg/policy/validation/<servicename>.go`)
- [ ] Register service and resources in `init()`
- [ ] Register validator in validation file's `init()`
- [ ] Import validation package in `pkg/policy/validator.go`
- [ ] Test with a sample policy

## Complete Workflow Example

This example demonstrates adding support for Glance (Image Service) from start to finish.

### Step 1: Setup Development Environment

```bash
# Clone repository
git clone <repository-url>
cd OSPA

# Install dependencies
go mod download

# Verify setup
go test ./...
```

### Step 2: Generate Service Code

```bash
# Use scaffold tool to generate all necessary files
go run ./cmd/scaffold \
  --service glance \
  --display-name Glance \
  --resources image,member \
  --type image
```

**Generated Files:**
- `pkg/services/services/glance.go` - Service implementation
- `pkg/discovery/services/glance.go` - Resource discovery
- `pkg/audit/glance/image.go` - Image auditor
- `pkg/audit/glance/member.go` - Member auditor
- `pkg/auth/auth.go` - Client method (appended)
- `pkg/policy/validation/glance.go` - Policy validation for Glance
- `pkg/audit/glance/image_test.go` - Unit tests
- `pkg/audit/glance/member_test.go` - Unit tests
- `e2e/glance_test.go` - E2E tests
- `examples/policies/glance-policy-guide.md` - Policy writing guide

### Step 3: Customize Generated Code

**Update Discovery** (`pkg/discovery/services/glance.go`):

```go
// Update imports to match your OpenStack client library
import (
    "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

// Update resource extraction
imageList, err := images.ExtractImages(page)
```

**Update Auditor** (`pkg/audit/glance/image.go`):

```go
// Update type assertions
image, ok := resource.(images.Image)
if !ok {
    return nil, fmt.Errorf("expected images.Image, got %T", resource)
}

// Customize check logic based on Image struct fields
if check.Status != "" && image.Status != check.Status {
    return result, nil
}
```

### Step 4: Run Unit Tests

```bash
# Run unit tests for the new service
go test ./pkg/audit/glance/... -v

# Fix any issues found
# Add additional test cases as needed
```

### Step 5: Test Service Registration

```bash
# Verify service is registered
go test ./pkg/services/... -v
```

### Step 6: Review Policy Guide

The scaffold tool generates a comprehensive policy guide at `examples/policies/glance-policy-guide.md`. This guide includes:
- Service overview and supported resources
- Policy structure and syntax
- Available check conditions
- Action types (log, delete, tag)
- Resource-specific examples
- OpenStack API documentation references
- Troubleshooting tips

### Step 7: Create Test Policy

Create `test-glance.yaml`:

```yaml
version: v1
defaults:
  workers: 2
policies:
  - glance:
    - name: test-image-active
      description: Test active images
      service: glance
      resource: image
      check:
        status: active
      action: log
    - name: test-image-old
      description: Test old images
      service: glance
      resource: image
      check:
        age_gt: 30d
      action: log
```

### Step 8: Test with Agent

```bash
# Run agent in audit mode
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy test-glance.yaml \
  --out test-findings.jsonl

# Review findings
cat test-findings.jsonl | jq .
```

### Step 8: Run E2E Tests

```bash
# Run e2e tests (requires OpenStack access)
OS_CLOUD=mycloud go test -tags=e2e ./e2e/glance_test.go -v
```

## Adding a New Service (Detailed)

### Step 1: Use Scaffold Tool

The scaffold tool generates all necessary files automatically:

```bash
# Example: Adding Glance (Image Service)
go run ./cmd/scaffold \
  --service glance \
  --display-name Glance \
  --resources image,member \
  --type image
```

### Step 2: Review Generated Code

**Service File** (`pkg/services/services/glance.go`):
```go
// Review and ensure:
// - Service name matches OpenStack service name
// - Resource types are correct
// - Client method name matches display name
```

**Discovery File** (`pkg/discovery/services/glance.go`):
```go
// Customize:
// - Import paths for OpenStack client libraries
// - Resource struct types
// - Field names (e.g., TenantID, ProjectID)
// - Pagination handling
```

**Auditor Files** (`pkg/audit/glance/*.go`):
```go
// Customize:
// - Resource struct type assertions
// - Check conditions based on resource capabilities
// - Fix() implementation for remediation actions
```

**Validation File** (`pkg/policy/validation/glance.go`):
```go
// Customize:
// - Replace TODO comments with actual validation rules
// - Add checks for required fields (e.g., status, age_gt, unused)
// - Validate field values (e.g., status must be one of allowed values)
// - Ensure at least one check condition is specified for each resource
// Example:
// case "image":
//     if check.Status == "" && check.AgeGT == "" {
//         return fmt.Errorf("rule %q: check must specify at least one of status or age_gt", ruleName)
//     }
```

### Step 3: Service Implementation Details

**File:** `pkg/services/services/<servicename>.go`

Key points:
- Implement `Service` interface
- Register in `init()` with `MustRegister()`
- Register each resource with `RegisterResource()`
- Document supported resources in comments

### Step 4: Auth Client

**File:** `pkg/auth/auth.go`

The scaffold tool automatically adds the client method. If adding manually:

```go
func (s *Session) Get<ServiceName>Client() (*gophercloud.ServiceClient, error) {
	client, err := clientconfig.NewServiceClient("<servicename>", &clientconfig.ClientOpts{
		Cloud: s.CloudName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create <servicename> client: %w", err)
	}
	return client, nil
}
```

### Step 5: Discovery Implementation

**File:** `pkg/discovery/services/<servicename>.go`

For each resource type:
- Create a discoverer struct
- Implement `Discoverer` interface
- Handle pagination and context cancellation

### Step 6: Auditor Implementation

**Directory:** `pkg/audit/<servicename>/`

For each resource type, create `<resource>.go`:
- Implement `Auditor` interface
- Implement `Check()` for policy evaluation
- Implement `Fix()` for remediation

### Step 7: Registration

In your service's `init()` function:
```go
func init() {
	MustRegister(&<ServiceName>Service{})
	RegisterResource("<servicename>", "resource1")
	RegisterResource("<servicename>", "resource2")
}
```

### Step 8: Import

Services are automatically registered when imported. Ensure the service packages are imported in `cmd/agent/main.go`:

```go
_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"        // Register services
_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services" // Register service implementations
_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services" // Register discoverers
```

This automatically imports all services via their `init()` functions.

### Step 9: Example Policy

Once implemented, you can use the service in policies:

```yaml
policies:
  - <servicename>:
    - name: my-rule
      description: Check resource1
      service: <servicename>
      resource: resource1
      check:
        status: active
      action: log
```

## Adding a Resource to Existing Service

### Example: Adding "backup" Resource to Cinder

**Step 1: Create Discoverer**

Add to `pkg/discovery/services/cinder.go`:

```go
type BlockStorageBackupDiscoverer struct{}

func (d *BlockStorageBackupDiscoverer) ResourceType() string {
	return "backup"
}

func (d *BlockStorageBackupDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error) {
	jobChan := make(chan Job, 100)
	
	go func() {
		defer close(jobChan)
		
		opts := backups.ListOpts{}
		if allTenants {
			opts.AllTenants = true
		}
		
		pager := backups.List(client, opts)
		err := pager.EachPage(func(page pagination.Page) (bool, error) {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			default:
			}
			
			backupList, err := backups.ExtractBackups(page)
			if err != nil {
				return false, err
			}
			
			for _, backup := range backupList {
				select {
				case <-ctx.Done():
					return false, ctx.Err()
				case jobChan <- Job{
					ResourceType: d.ResourceType(),
					ResourceID:   backup.ID,
					Resource:     backup,
					Service:      "cinder",
					ProjectID:    backup.ProjectID,
				}:
				}
			}
			return true, nil
		})
		
		if err != nil {
			log.Printf("Error discovering backups: %v", err)
		}
	}()
	
	return jobChan, nil
}
```

**Step 2: Create Auditor**

Create `pkg/audit/cinder/backup.go`:

```go
package blockstorage

import (
	"context"
	"fmt"
	"time"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/backups"
)

type BackupAuditor struct{}

func (a *BackupAuditor) ResourceType() string {
	return "backup"
}

func (a *BackupAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
	backup, ok := resource.(backups.Backup)
	if !ok {
		return nil, fmt.Errorf("expected backups.Backup, got %T", resource)
	}
	
	result := &audit.Result{
		RuleID:       rule.Name,
		ResourceID:   backup.ID,
		ResourceName: backup.Name,
		ProjectID:    backup.ProjectID,
		Compliant:    true,
		Rule:         rule,
		Status:       backup.Status,
	}
	
	check := rule.Check
	
	// Implement check logic
	if check.Status != "" && backup.Status != check.Status {
		return result, nil
	}
	
	// Add age check if needed
	if check.AgeGT != "" {
		ageThreshold, err := check.ParseAgeGT()
		if err != nil {
			return nil, fmt.Errorf("failed to parse age_gt: %w", err)
		}
		
		age := time.Since(backup.CreatedAt)
		if age > ageThreshold {
			result.Compliant = false
			result.Observation = fmt.Sprintf("Backup is %s old (threshold: %s)", age, ageThreshold)
		}
	}
	
	return result, nil
}

func (a *BackupAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	if rule.Action != "delete" {
		return nil
	}
	
	backup, ok := resource.(backups.Backup)
	if !ok {
		return fmt.Errorf("expected backups.Backup, got %T", resource)
	}
	
	serviceClient, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}
	
	return backups.Delete(serviceClient, backup.ID).ExtractErr()
}
```

**Step 3: Update Service**

Update `pkg/services/services/cinder.go`:

```go
// In GetResourceAuditor():
case "backup":
	return &blockstorage.BackupAuditor{}, nil

// In GetResourceDiscoverer():
case "backup":
	return &discovery.BlockStorageBackupDiscoverer{}, nil

// In init():
RegisterResource("cinder", "backup")
```

**Step 4: Write Tests**

Create `pkg/audit/cinder/backup_test.go`:

```go
package blockstorage

import (
	"context"
	"testing"
	"time"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/backups"
)

func TestBackupAuditor_ResourceType(t *testing.T) {
	auditor := &BackupAuditor{}
	if auditor.ResourceType() != "backup" {
		t.Errorf("expected 'backup', got %q", auditor.ResourceType())
	}
}

func TestBackupAuditor_Check_StatusMatch(t *testing.T) {
	auditor := &BackupAuditor{}
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
	
	result, err := auditor.Check(context.Background(), backup, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if result.ResourceID != "backup-1" {
		t.Errorf("expected ResourceID 'backup-1', got %q", result.ResourceID)
	}
	
	if !result.Compliant {
		t.Error("expected compliant result for matching status")
	}
}

func TestBackupAuditor_Check_AgeGT(t *testing.T) {
	auditor := &BackupAuditor{}
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
	
	result, err := auditor.Check(context.Background(), backup, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if result.Compliant {
		t.Error("expected non-compliant result for old backup")
	}
}
```

**Step 5: Test Integration**

```bash
# Run unit tests
go test ./pkg/audit/cinder/... -v

# Test with agent
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy test-backup-policy.yaml \
  --out test-findings.jsonl
```

## Testing

### Unit Tests

Run all unit tests:

```bash
go test ./...
```

Run tests for specific package:

```bash
go test ./pkg/audit/nova/... -v
```

Run specific test:

```bash
go test ./pkg/audit/nova/... -run TestInstanceAuditor_Check -v
```

### Integration Tests

Test service registration:

```bash
go test ./pkg/services/... -v
```

Test policy validation:

```bash
go test ./pkg/policy/... -v
```

### E2E Tests

**Prerequisites:**
- `OS_CLOUD` environment variable set
- Access to OpenStack cloud with existing resources

**Run all e2e tests:**

```bash
OS_CLOUD=mycloud go test -tags=e2e ./e2e/... -v
```

**Run specific service tests:**

```bash
OS_CLOUD=mycloud go test -tags=e2e ./e2e/compute_test.go -v
```

**Run with custom policy:**

```bash
OSPA_E2E_POLICY=./test-policy.yaml \
OS_CLOUD=mycloud \
go test -tags=e2e ./e2e/... -v
```

### Test Coverage

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Running the Agent

### Audit Mode (Safe, Default)

```bash
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy ./examples/policies.yaml \
  --out findings.jsonl
```

### Enforce Mode (Applies Remediations)

```bash
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy ./examples/policies.yaml \
  --out findings.jsonl \
  --apply
```

### All Tenants (Requires Admin)

```bash
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy ./examples/policies.yaml \
  --out findings.jsonl \
  --all-tenants
```

### Custom Worker Count

```bash
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy ./examples/policies.yaml \
  --out findings.jsonl \
  --workers 10
```

## Troubleshooting

### Service Not Found

**Error:** `service "glance" not found`

**Solution:**
1. Verify service is registered in `init()`:
   ```go
   func init() {
       MustRegister(&GlanceService{})
   }
   ```
2. Ensure service package is imported (automatic via `pkg/services` package)
3. Check service name matches exactly (case-sensitive)

### Resource Not Supported

**Error:** `unsupported resource type "image" for service "glance"`

**Solution:**
1. Verify resource is registered:
   ```go
   RegisterResource("glance", "image")
   ```
2. Check resource type matches exactly (case-sensitive)
3. Ensure discoverer and auditor are implemented

### Validation Fails

**Error:** Policy validation fails

**Solution:**
1. Check policy YAML syntax
2. Verify service and resource names match registered ones
3. Check required fields are present
4. Run validator tests: `go test ./pkg/policy/... -v`

### Client Creation Fails

**Error:** `failed to create glance client`

**Solution:**
1. Verify service type matches OpenStack service name
2. Check `clouds.yaml` has correct service endpoints
3. Verify authentication credentials
4. Test with: `openstack image list` (for Glance)

### Permission Errors

**Error:** `403 Forbidden` or `401 Unauthorized`

**Solution:**
1. Verify credentials in `clouds.yaml`
2. Check project has required permissions
3. For `--all-tenants`, ensure admin credentials
4. Test with OpenStack CLI first

### No Resources Found

**Warning:** No resources discovered

**Solution:**
1. Verify resources exist in OpenStack
2. Check `--all-tenants` if resources are in other projects
3. Verify discoverer implementation is correct
4. Check OpenStack API endpoints are accessible

### Test Failures

**Unit Tests Fail:**

1. Check test data matches actual OpenStack resource structure
2. Verify type assertions are correct
3. Check error handling logic

**E2E Tests Fail:**

1. Verify `OS_CLOUD` is set correctly
2. Check OpenStack connectivity
3. Ensure test project has resources
4. Review test logs for specific errors

## Development Best Practices

1. **Always use scaffold tool** for new services to ensure consistency
2. **Write tests first** (TDD approach) or immediately after implementation
3. **Test incrementally** - unit tests → integration → e2e
4. **Follow existing patterns** - look at compute/network/blockstorage examples
5. **Document edge cases** in code comments
6. **Handle errors gracefully** - return errors, don't panic
7. **Respect context cancellation** - always check `ctx.Done()`
8. **Use type assertions with error checking** - never assume types
9. **Register resources** in `init()` for automatic validation
10. **Test with real policies** before submitting changes

## Common Development Patterns

### Pattern 1: Status-Based Checks

```go
if check.Status != "" && resource.Status != check.Status {
	return result, nil // Not a match
}
result.Compliant = false
```

### Pattern 2: Age-Based Checks

```go
if check.AgeGT != "" {
	ageThreshold, err := check.ParseAgeGT()
	if err != nil {
		return nil, fmt.Errorf("failed to parse age_gt: %w", err)
	}
	
	age := time.Since(resource.CreatedAt)
	if age > ageThreshold {
		result.Compliant = false
		result.Observation = fmt.Sprintf("Resource is %s old", age)
	}
}
```

### Pattern 3: Delete Remediation

```go
func (a *ResourceAuditor) Fix(ctx context.Context, client interface{}, resource interface{}, rule *policy.Rule) error {
	if rule.Action != "delete" {
		return nil
	}
	
	res, ok := resource.(ResourceType)
	if !ok {
		return fmt.Errorf("expected ResourceType, got %T", resource)
	}
	
	serviceClient, ok := client.(*gophercloud.ServiceClient)
	if !ok {
		return fmt.Errorf("expected *gophercloud.ServiceClient, got %T", client)
	}
	
	return resources.Delete(serviceClient, res.ID).ExtractErr()
}
```

## Testing Checklist

Use this checklist when adding new services or resources:

- [ ] Code generated using scaffold tool
- [ ] Service file created and registered
- [ ] Discovery implementation complete
- [ ] Auditor implementation complete
- [ ] Resource registered in `init()`
- [ ] Unit tests written and passing
- [ ] Service registration verified
- [ ] Resource discovery verified
- [ ] Policy validation works
- [ ] Agent can load and use service
- [ ] E2E tests written and passing
- [ ] Documentation updated

## Reference Implementations

For detailed implementation examples, see the existing service implementations:
- Nova: `pkg/services/services/nova.go`, `pkg/discovery/services/nova.go`, `pkg/audit/nova/`
- Neutron: `pkg/services/services/neutron.go`, `pkg/discovery/services/neutron.go`, `pkg/audit/neutron/`
- Cinder: `pkg/services/services/cinder.go`, `pkg/discovery/services/cinder.go`, `pkg/audit/cinder/`

## Next Steps

- See [`ARCHITECTURE.md`](ARCHITECTURE.md) for detailed architecture explanation
- See [`scaffold-README.md`](scaffold-README.md) for scaffold tool details
- Review existing service implementations for examples
