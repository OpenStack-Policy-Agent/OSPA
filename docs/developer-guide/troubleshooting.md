# Troubleshooting

This guide covers common issues and their solutions when developing with OSPA.

## Service Issues

### Service Not Found

**Error:**
```
service "glance" not found
```

**Cause:** Service is not registered.

**Solution:**

1. Verify service is registered in `init()`:
   ```go
   func init() {
       MustRegister(&GlanceService{})
   }
   ```

2. Ensure service package is imported in `cmd/agent/main.go`:
   ```go
   _ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services"
   ```

3. Check service name matches exactly (case-sensitive)

### Resource Not Supported

**Error:**
```
unsupported resource type "image" for service "glance"
```

**Cause:** Resource is not registered for the service.

**Solution:**

1. Verify resource is registered:
   ```go
   RegisterResource("glance", "image")
   ```

2. Check resource type matches exactly

3. Ensure discoverer and auditor are implemented and returned by service

## Client Issues

### Client Creation Fails

**Error:**
```
failed to create glance client: ...
```

**Cause:** OpenStack authentication or service endpoint issue.

**Solution:**

1. Verify service type in `GetClient()`:
   ```go
   clientconfig.NewServiceClient("image", &clientconfig.ClientOpts{...})
   ```

2. Check `clouds.yaml` has correct endpoints

3. Test with OpenStack CLI:
   ```bash
   openstack image list
   ```

### Permission Errors

**Error:**
```
403 Forbidden
```

**Cause:** User lacks required permissions.

**Solution:**

1. Verify credentials in `clouds.yaml`
2. Check project has required roles
3. For `--all-tenants`, ensure admin credentials
4. Test with OpenStack CLI first

## Policy Issues

### Validation Fails

**Error:**
```
policy validation failed: rule "my-rule": unknown resource type
```

**Cause:** Policy references unregistered resource.

**Solution:**

1. Check policy YAML syntax
2. Verify service and resource names match registered ones
3. Run validator tests:
   ```bash
   go test ./pkg/policy/... -v
   ```

### Tag Name Required

**Error:**
```
rule "my-rule": tag_name is required when action is 'tag'
```

**Cause:** Tag action without tag_name field.

**Solution:**

Add `tag_name` to the policy rule:
```yaml
action: tag
tag_name: ospa-audited
```

## Discovery Issues

### No Resources Found

**Warning:** No resources discovered.

**Cause:** Discovery returns empty results.

**Solution:**

1. Verify resources exist in OpenStack
2. Check `--all-tenants` if resources are in other projects
3. Review discoverer implementation:
   ```go
   // Check opts are correct
   opts := resources.ListOpts{}
   if allTenants {
       opts.AllTenants = true
   }
   ```
4. Add debug logging to discoverer
5. Check OpenStack API endpoints are accessible

### Discovery Hangs

**Symptom:** Agent never completes discovery.

**Cause:** Channel not closed or pagination issue.

**Solution:**

1. Ensure channel is closed:
   ```go
   go func() {
       defer close(jobChan)  // Must be present
       // ...
   }()
   ```

2. Check pagination terminates:
   ```go
   pager.EachPage(func(page pagination.Page) (bool, error) {
       // Return false to stop, true to continue
       return true, nil
   })
   ```

3. Verify context cancellation is handled:
   ```go
   select {
   case <-ctx.Done():
       return false, ctx.Err()
   default:
   }
   ```

## Auditor Issues

### Type Assertion Fails

**Error:**
```
expected images.Image, got map[string]interface{}
```

**Cause:** Resource wasn't properly typed during discovery.

**Solution:**

1. Ensure discoverer sends correct type:
   ```go
   jobChan <- discovery.Job{
       Resource: image,  // Must be images.Image, not raw response
   }
   ```

2. Use proper extraction in discoverer:
   ```go
   imageList, err := images.ExtractImages(page)
   ```

### Check Always Returns Compliant

**Symptom:** No violations found when expected.

**Cause:** Check logic inverted or incomplete.

**Solution:**

Review check logic:
```go
// WRONG: Returns compliant when status matches
if check.Status != "" && resource.Status != check.Status {
    result.Compliant = false
}

// CORRECT: Returns non-compliant when status matches
if check.Status != "" && resource.Status == check.Status {
    result.Compliant = false
}
```

## Test Issues

### Unit Tests Fail

**Cause:** Test data doesn't match resource structure.

**Solution:**

1. Check struct field names match OpenStack types
2. Use correct test data:
   ```go
   sg := groups.SecGroup{
       ID:       "sg-1",
       Name:     "test",
       TenantID: "project-1",  // Not ProjectID
   }
   ```

3. Verify time format matches:
   ```go
   UpdatedAt: time.Now(),  // Not string
   ```

### E2E Tests Fail

**Cause:** OpenStack connectivity or resource issues.

**Solution:**

1. Verify `OS_CLOUD` is set:
   ```bash
   echo $OS_CLOUD
   ```

2. Check OpenStack connectivity:
   ```bash
   openstack token issue
   ```

3. Ensure test project has resources

4. Review test logs:
   ```bash
   go test -tags=e2e ./e2e/... -v 2>&1 | tee test.log
   ```

5. Clean up orphaned resources:
   ```bash
   go test -tags=e2e ./e2e/... -run TestCleanup -v
   ```

## Build Issues

### Import Errors

**Error:**
```
cannot find package "github.com/gophercloud/gophercloud/openstack/..."
```

**Solution:**

```bash
go mod tidy
go mod download
```

### Circular Import

**Error:**
```
import cycle not allowed
```

**Solution:**

1. Move shared types to a common package
2. Use interfaces instead of concrete types
3. Reorganize package structure

## Performance Issues

### High Memory Usage

**Symptom:** Agent uses excessive memory.

**Solution:**

1. Reduce worker count:
   ```bash
   --workers 10
   ```

2. Use buffered channels appropriately:
   ```go
   jobChan := make(chan Job, 100)
   ```

3. Process results as they come instead of collecting all

### Slow Discovery

**Symptom:** Discovery takes very long.

**Solution:**

1. Check pagination efficiency
2. Reduce page size if too large
3. Add parallelism where appropriate
4. Check network latency to OpenStack

## Debugging

### Enable Debug Logging

```bash
go run ./cmd/agent \
  --log-level debug \
  --log-format json \
  ...
```

### Use Delve

```bash
# Debug agent
dlv debug ./cmd/agent -- --cloud mycloud --policy policy.yaml

# Set breakpoint
(dlv) break pkg/audit/neutron/security_group.go:50
(dlv) continue
```

### Add Temporary Logging

```go
import "log"

func (a *Auditor) Check(...) (*Result, error) {
    log.Printf("DEBUG: Checking resource %s", resource.ID)
    // ...
}
```

## Getting Help

If you're still stuck:

1. Search existing [GitHub Issues](https://github.com/OpenStack-Policy-Agent/OSPA/issues)
2. Check the [architecture documentation](architecture.md)
3. Open a new issue with:
   - OSPA version/commit
   - Go version
   - Full error message
   - Steps to reproduce
   - Relevant code snippets

