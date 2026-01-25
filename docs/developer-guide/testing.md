# Testing

This guide covers OSPA's testing strategy, including unit tests, integration tests, and end-to-end tests.

## Test Types

| Type | Location | Purpose | Requires OpenStack |
|------|----------|---------|-------------------|
| Unit | `pkg/*_test.go` | Test individual functions | No |
| Integration | `pkg/*_test.go` (tagged) | Test component integration | Optional |
| E2E | `e2e/` | Test full workflow | Yes |

## Running Tests

### All Unit Tests

```bash
go test ./...
```

### Package-Specific Tests

```bash
# Specific package
go test ./pkg/audit/neutron/... -v

# Specific test
go test ./pkg/audit/neutron/... -run TestSecurityGroup -v
```

### With Race Detector

```bash
go test -race ./...
```

### With Coverage

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Coverage summary
go tool cover -func=coverage.out
```

## Unit Testing

### Auditor Tests

Test each auditor's Check() and Fix() methods:

```go
package neutron

import (
    "context"
    "testing"

    "github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
    "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
)

func TestSecurityGroupAuditor_ResourceType(t *testing.T) {
    a := &SecurityGroupAuditor{}
    if got := a.ResourceType(); got != "security_group" {
        t.Errorf("ResourceType() = %q, want %q", got, "security_group")
    }
}

func TestSecurityGroupAuditor_Check_StatusMatch(t *testing.T) {
    a := &SecurityGroupAuditor{}

    sg := groups.SecGroup{
        ID:   "sg-1",
        Name: "test-sg",
    }

    rule := &policy.Rule{
        Name:   "test-rule",
        Action: "log",
        Check: policy.CheckConditions{
            Status: "ACTIVE",
        },
    }

    result, err := a.Check(context.Background(), sg, rule)
    if err != nil {
        t.Fatalf("Check() error = %v", err)
    }

    if result.Compliant {
        t.Error("expected non-compliant for matching status")
    }
}
```

### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestSecurityGroupAuditor_Check(t *testing.T) {
    tests := []struct {
        name      string
        sg        groups.SecGroup
        rule      *policy.Rule
        wantCompl bool
        wantErr   bool
    }{
        {
            name: "status match",
            sg: groups.SecGroup{
                ID:   "sg-1",
                Name: "test",
            },
            rule: &policy.Rule{
                Check: policy.CheckConditions{Status: "ACTIVE"},
            },
            wantCompl: false,
        },
        {
            name: "exempt by name",
            sg: groups.SecGroup{
                ID:   "sg-1",
                Name: "default",
            },
            rule: &policy.Rule{
                Check: policy.CheckConditions{
                    Status:      "ACTIVE",
                    ExemptNames: []string{"default"},
                },
            },
            wantCompl: true,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            a := &SecurityGroupAuditor{}
            got, err := a.Check(context.Background(), tt.sg, tt.rule)

            if (err != nil) != tt.wantErr {
                t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if got.Compliant != tt.wantCompl {
                t.Errorf("Check() compliant = %v, want %v", got.Compliant, tt.wantCompl)
            }
        })
    }
}
```

### Testing Discoverers

Use mock clients or test against actual behavior:

```go
func TestDiscoverer_Discover(t *testing.T) {
    // Skip if no OpenStack access
    if os.Getenv("OS_CLOUD") == "" {
        t.Skip("OS_CLOUD not set, skipping integration test")
    }

    // ... test with real client
}
```

## E2E Testing

### Location

E2E tests are in `e2e/<service>/`:

```
e2e/
├── engine.go               # Test engine
├── neutron/
│   ├── resource_creator.go # Create test resources
│   ├── network_test.go
│   ├── security_group_test.go
│   └── ...
├── nova/
│   ├── resource_creator.go
│   └── instance_test.go
└── cinder/
    └── ...
```

### Test Engine

The test engine provides common functionality:

```go
func TestNeutron_SecurityGroup_StatusCheck(t *testing.T) {
    engine := e2e.NewTestEngine(t)
    client := engine.GetNetworkClient(t)

    // Create test resource
    resourceID, cleanup := CreateSecurityGroup(t, client)
    defer cleanup()

    // Define policy
    policyYAML := `
version: v1
policies:
  - neutron:
    - name: test-sg-status
      service: neutron
      resource: security_group
      check:
        status: ACTIVE
      action: log
`

    // Run audit
    policy := engine.LoadPolicyFromYAML(t, policyYAML)
    results := engine.RunAudit(t, policy)

    // Verify results
    resourceResults := results.FilterByResourceID(resourceID)
    if resourceResults.Violations == 0 {
        t.Error("expected violation")
    }
}
```

### Resource Creators

Create helpers for test resources:

```go
// e2e/neutron/resource_creator.go
func CreateSecurityGroup(t *testing.T, client *gophercloud.ServiceClient) (string, func()) {
    t.Helper()

    name := fmt.Sprintf("ospa-e2e-sg-%d", time.Now().UnixNano())
    
    sg, err := groups.Create(client, groups.CreateOpts{
        Name:        name,
        Description: "OSPA E2E test security group",
    }).Extract()
    if err != nil {
        t.Fatalf("Failed to create security group: %v", err)
    }

    cleanup := func() {
        t.Logf("Cleaning up security group: %s", sg.ID)
        groups.Delete(client, sg.ID)
    }

    return sg.ID, cleanup
}
```

### Running E2E Tests

```bash
# Set credentials
export OS_CLIENT_CONFIG_FILE=/path/to/clouds.yaml
export OS_CLOUD=mycloud

# Run all E2E tests
go test -tags=e2e ./e2e/... -v

# Run specific service
go test -tags=e2e ./e2e/neutron/... -v

# Run specific test
go test -tags=e2e ./e2e/neutron/... -run TestSecurityGroup -v
```

### E2E Test Guidelines

1. **Clean up resources** - Always use defer for cleanup
2. **Unique names** - Use timestamps in resource names
3. **Isolation** - Each test should be independent
4. **Timeout handling** - Set appropriate timeouts
5. **Error context** - Provide helpful error messages

### Orphan Cleanup

Handle orphaned test resources:

```go
func TestCleanup_SecurityGroup(t *testing.T) {
    engine := e2e.NewTestEngine(t)
    client := engine.GetNetworkClient(t)

    CleanupOrphans(t, client)
}

func CleanupOrphans(t *testing.T, client *gophercloud.ServiceClient) {
    pages, _ := groups.List(client, groups.ListOpts{}).AllPages()
    sgs, _ := groups.ExtractGroups(pages)

    for _, sg := range sgs {
        if strings.HasPrefix(sg.Name, "ospa-e2e-") {
            t.Logf("Cleaning orphan: %s", sg.Name)
            groups.Delete(client, sg.ID)
        }
    }
}
```

## Test Coverage

### Coverage Report

```bash
# Generate coverage for pkg only
go test -coverprofile=pkg.cover.out ./pkg/...

# View HTML report
go tool cover -html=pkg.cover.out -o coverage.html

# Summary
go tool cover -func=pkg.cover.out | tail -1
```

### Coverage Script

```bash
#!/bin/bash
# scripts/test-pkg.sh

set -e

echo "Running pkg unit tests with coverage..."
go test ./pkg/... -count=1 -race -coverprofile=pkg.cover.out

echo ""
echo "Coverage summary:"
go tool cover -func=pkg.cover.out | tail -1

echo ""
echo "Generating HTML report..."
go tool cover -html=pkg.cover.out -o pkg.cover.html
echo "Report: pkg.cover.html"
```

## CI Configuration

### GitHub Actions

```yaml
name: Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: go test -race -coverprofile=coverage.out ./pkg/...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: coverage.out
```

## Mocking

### Interface-Based Mocking

Design for testability with interfaces:

```go
type Discoverer interface {
    Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error)
}

// In tests
type MockDiscoverer struct {
    jobs []Job
}

func (m *MockDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error) {
    ch := make(chan Job)
    go func() {
        defer close(ch)
        for _, job := range m.jobs {
            ch <- job
        }
    }()
    return ch, nil
}
```

### Test Helpers

```go
// testutil/helpers.go
func MakeTestRule(name string, check policy.CheckConditions) *policy.Rule {
    return &policy.Rule{
        Name:   name,
        Action: "log",
        Check:  check,
    }
}

func MakeTestSecurityGroup(id, name string) groups.SecGroup {
    return groups.SecGroup{
        ID:   id,
        Name: name,
    }
}
```

## Best Practices

1. **Test behavior, not implementation** - Focus on what, not how
2. **One assertion per test** - Keep tests focused
3. **Descriptive names** - Test names should explain the scenario
4. **Independent tests** - Tests should not depend on each other
5. **Fast tests** - Unit tests should be quick
6. **Meaningful coverage** - Cover edge cases, not just lines

