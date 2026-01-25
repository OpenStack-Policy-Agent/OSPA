# OSPA Scaffold Tool

The scaffold tool generates boilerplate code for adding new OpenStack services to OSPA.

## Usage

```bash
go run ./cmd/scaffold --service <name> --resources <list>
```

Or build and install:

```bash
go build -o ospa-scaffold ./cmd/scaffold
./ospa-scaffold --service nova --resources instance,keypair
```

## Options

- `--service` (required): Service name (e.g., `nova`, `neutron`, `cinder`)
- `--resources` (required): Comma-separated list of resource types (e.g., `instance,keypair`)
- `--list`: List all available OpenStack services and resources

## Examples

### Generate Nova (Compute) support

```bash
go run ./cmd/scaffold --service nova --resources instance,keypair
```

### Generate Neutron (Network) support

```bash
go run ./cmd/scaffold --service neutron --resources security_group,security_group_rule,floating_ip
```

### List available services

```bash
go run ./cmd/scaffold --list
```

## Generated Files

The tool generates:

1. **Service file** (`pkg/services/services/<servicename>.go`)
   - Service implementation with resource registration
   - Client, auditor, and discoverer wiring

2. **Discovery file** (`pkg/discovery/services/<servicename>.go`)
   - Discoverer stubs for each resource type
   - TODO comments with gophercloud references

3. **Auditor files** (`pkg/audit/<servicename>/<resource>.go`)
   - Auditor implementation for each resource type
   - TODO comments for Check() and Fix() implementation

4. **Auth client method** (appended to `pkg/auth/auth.go`)
   - `Get<DisplayName>Client()` method for service authentication

5. **Unit test files** (`pkg/audit/<servicename>/<resource>_test.go`)
   - Basic unit tests for each auditor

6. **E2E test directory** (`e2e/<servicename>/`)
   - `resource_creator.go` - Helper functions to create test resources and dependencies
   - `<resource>_test.go` - Individual test files for each resource with coverage checklist
   - Also adds `Get<Service>Client()` method to `e2e/engine.go` if not already present

7. **Validation file** (`pkg/policy/validation/<servicename>.go`)
   - Service-specific policy validator

8. **Policy guide** (`examples/policies/<servicename>-policy-guide.md`)
   - Documentation for writing policies for the new service

## Next Steps After Generating

### 1. Update the registry config with accurate checks

**Important:** The default checks in the registry config may not match what the OpenStack API actually supports for each resource.

Edit `cmd/scaffold/internal/registry/config/<servicename>.yaml` and update the `checks` field for each resource:

```yaml
resources:
  instance:
    description: Server instances
    checks:
      - status      # Most resources have status
      - age_gt      # If the resource has updated_at/created_at
      - unused      # Only if "unused" detection makes sense for this resource
      - exempt_names
    actions:
      - log
      - delete
      - tag         # Only if the OpenStack API supports tagging this resource
```

**How to determine valid checks:**

- **status**: Check the OpenStack API docs for the resource's status field values
  - Nova instances: `ACTIVE`, `SHUTOFF`, `ERROR`, etc.
  - Neutron resources: `ACTIVE`, `DOWN`, `BUILD`, etc.
- **age_gt**: Only valid if the resource has `updated_at` or `created_at` timestamps
- **unused**: Only valid if there's a logical way to determine "unused" for this resource type
  - Floating IPs: unused if not attached to a port
  - Security groups: unused if not attached to any ports
  - Keypairs: unused if not referenced by any instances
- **exempt_names**: Always valid (filters by resource name)

**OpenStack API References:**
- Nova: https://docs.openstack.org/api-ref/compute/
- Neutron: https://docs.openstack.org/api-ref/network/
- Cinder: https://docs.openstack.org/api-ref/block-storage/
- Glance: https://docs.openstack.org/api-ref/image/
- Keystone: https://docs.openstack.org/api-ref/identity/

### 2. Implement the discoverers

Edit `pkg/discovery/services/<servicename>.go`:

```go
func (d *NovaInstanceDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
    ch := make(chan discovery.Job)

    go func() {
        defer close(ch)

        // Use gophercloud to list resources
        opts := servers.ListOpts{AllTenants: allTenants}
        pages, err := servers.List(client, opts).AllPages()
        if err != nil {
            return
        }
        instances, _ := servers.ExtractServers(pages)

        for _, instance := range instances {
            select {
            case <-ctx.Done():
                return
            case ch <- discovery.Job{
                Service:      "nova",
                ResourceType: "instance",
                ResourceID:   instance.ID,
                ProjectID:    instance.TenantID,
                Resource:     instance,
            }:
            }
        }
    }()

    return ch, nil
}
```

**Gophercloud docs:** https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack

### 3. Implement the auditors

Edit `pkg/audit/<servicename>/<resource>.go`:

```go
func (a *InstanceAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
    // Cast to the correct type
    instance := resource.(servers.Server)

    result := &audit.Result{
        RuleID:       rule.Name,
        ResourceID:   instance.ID,
        ResourceName: instance.Name,
        ProjectID:    instance.TenantID,
        Status:       instance.Status,
        UpdatedAt:    instance.Updated,
        Compliant:    true,
        Rule:         rule,
    }

    // Implement checks
    if rule.Check.Status != "" && instance.Status == rule.Check.Status {
        result.Compliant = false
        result.Observation = fmt.Sprintf("status is %s", instance.Status)
    }

    // Age check
    if rule.Check.AgeGT != "" {
        age, _ := time.ParseDuration(rule.Check.AgeGT)
        if time.Since(instance.Updated) > age {
            result.Compliant = false
            result.Observation = fmt.Sprintf("resource is older than %s", rule.Check.AgeGT)
        }
    }

    return result, nil
}
```

### 4. Implement E2E tests

E2E tests are organized per service in `e2e/<servicename>/`:

```
e2e/neutron/
├── resource_creator.go    # Implement Create<Resource>() functions here
├── network_test.go        # Tests for network resource
├── port_test.go           # Tests for port resource (uses network as dependency)
└── ...
```

**Step 1: Implement resource creators in `resource_creator.go`**

Each `Create<Resource>()` function should:
1. Create any dependencies (e.g., port needs network+subnet)
2. Create the test resource
3. Return the resource ID and a cleanup function

```go
func CreatePort(t *testing.T, client *gophercloud.ServiceClient) (resourceID string, cleanup func()) {
    // Create network first (dependency)
    networkID, networkCleanup := CreateNetwork(t, client)
    
    // Create port
    port, err := ports.Create(client, ports.CreateOpts{
        Name:      fmt.Sprintf("ospa-e2e-port-%d", time.Now().UnixNano()),
        NetworkID: networkID,
    }).Extract()
    if err != nil {
        networkCleanup()
        t.Fatalf("Failed to create port: %v", err)
    }
    
    cleanup = func() {
        ports.Delete(client, port.ID)
        networkCleanup() // Clean up network after port
    }
    
    return port.ID, cleanup
}
```

**Step 2: Use the creators in test files**

The generated test files call your `Create<Resource>()` function:

```go
func TestNeutron_Port_StatusCheck(t *testing.T) {
    engine := e2e.NewTestEngine(t)
    client := engine.GetNetworkClient(t)
    
    resourceID, cleanup := CreatePort(t, client)
    defer cleanup()
    
    // Run audit and verify results...
}
```

### 5. Run tests

```bash
# Unit tests
go test ./pkg/audit/<servicename>/...

# E2E tests (requires OpenStack)
OS_CLOUD=mycloud go test -tags=e2e ./e2e/<servicename>/... -v
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
        ├── policy_guide.go     # Policy guide generation
        ├── metadata.go         # ResourceSpec helpers
        └── utils.go            # Utilities
```

## Notes

- Generated code uses placeholder implementations with TODO comments
- The registry config files (`cmd/scaffold/internal/registry/config/*.yaml`) define valid services, resources, checks, and actions
- Update the registry config to match OpenStack API capabilities before generating
- Regenerating overwrites existing files
