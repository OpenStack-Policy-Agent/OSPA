# OSPA Architecture Guide

## Overview

OSPA (OpenStack Policy Agent) follows a plugin-based architecture that makes it easy to add support for new OpenStack services and resource types. This document explains the architecture and how to extend it.

## Architecture Components

### 1. Service Layer (`pkg/services/`)

Services are the top-level abstraction for OpenStack services (Nova, Neutron, Cinder, etc.). Each service:
- Manages authentication and client creation
- Provides access to resource auditors and discoverers
- Registers itself automatically via `init()` functions

**Key Files:**
- `interface.go` - Service interface definition
- `registry.go` - Service registration and lookup
- `resource_registry.go` - Resource type registration
- `services/<servicename>.go` - Service implementations

**To add a new service:**
1. Create `pkg/services/services/<servicename>.go`
2. Implement the `Service` interface
3. Register in `init()` using `MustRegister()`
4. Register supported resources using `RegisterResource()`

See [`DEVELOPMENT.md`](DEVELOPMENT.md) for detailed instructions on adding new services.

### 2. Discovery Layer (`pkg/discovery/`)

Discoverers are responsible for finding resources in OpenStack. Each resource type has a discoverer that:
- Lists resources from the OpenStack API
- Converts them to generic `Job` structures
- Handles pagination and context cancellation

**Key Files:**
- `interface.go` - Discoverer interface
- `job.go` - Generic job structure
- `helpers.go` - Common discovery helper functions
- `services/<servicename>.go` - Service-specific discoverers

**To add discovery for a new resource:**
1. Create a discoverer struct implementing `Discoverer`
2. Implement `Discover()` method that returns a channel of `Job`
3. Register in the service's `GetResourceDiscoverer()` method

### 3. Audit Layer (`pkg/audit/`)

Auditors evaluate resources against policy rules. Each resource type has an auditor that:
- Implements `Check()` to evaluate compliance
- Implements `Fix()` to apply remediation
- Returns structured `Result` objects

**Key Files:**
- `interface.go` - Auditor interface
- `result.go` - Result structure
- `registry.go` - Auditor registry (optional)
- `<servicename>/<resource>.go` - Resource-specific auditors

**To add auditing for a new resource:**
1. Create an auditor struct implementing `Auditor`
2. Implement `Check()` method for policy evaluation
3. Implement `Fix()` method for remediation (if applicable)
4. Register in the service's `GetResourceAuditor()` method

### 4. Policy Layer (`pkg/policy/`)

Policies define rules for auditing resources. The policy system:
- Loads YAML policy files
- Validates policy structure
- Provides rule evaluation context

**Key Files:**
- `policy.go` - Policy structures
- `loader.go` - YAML loading and parsing
- `validator.go` - Main policy validation orchestrator
- `validation/` - Service-specific validation modules
  - `interface.go` - Validator interface
  - `registry.go` - Validator registry
  - `<servicename>.go` - Service-specific validators (e.g., `nova.go`, `neutron.go`)

**Validation Architecture:**
- Each OpenStack service has its own validator in `pkg/policy/validation/<servicename>.go`
- Validators register themselves automatically via `init()` functions
- The main validator delegates resource-specific validation to service validators
- This modular approach allows validation rules to be co-located with service implementations

**Policy Structure:**
```yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - <service>:
    - name: rule-name
      description: Rule description
      service: <service>
      resource: <resource_type>
      check:
        # Check conditions
      action: log|delete|tag
```

### 5. Orchestrator (`pkg/orchestrator/`)

The orchestrator coordinates the entire audit process:
- Loads and validates policies
- Manages worker pools
- Coordinates discovery, audit, and remediation
- Handles graceful shutdown

**Key Files:**
- `orchestrator.go` - Main orchestration logic

### 6. Remediation (`pkg/remediate/`)

The remediation system provides action handlers:
- `log` - No-op, just logs violations
- `delete` - Deletes resources
- `tag` - Tags resources with metadata

**Key Files:**
- `interface.go` - Remediator interface
- `registry.go` - Action registry
- `actions.go` - Action implementations

## Extension Points

### Adding a New Service

1. **Create Service Implementation** (`pkg/services/services/<servicename>.go`)
   ```go
   type <Service>Service struct{}
   func init() {
       MustRegister(&<Service>Service{})
       RegisterResource("<servicename>", "resource1")
   }
   ```

2. **Add Client Method** (`pkg/auth/auth.go`)
   ```go
   func (s *Session) Get<Service>Client() (*gophercloud.ServiceClient, error)
   ```

3. **Create Discoverers** (`pkg/discovery/services/<servicename>.go`)
   - One discoverer per resource type

4. **Create Auditors** (`pkg/audit/<servicename>/`)
   - One auditor per resource type

5. **Import Service** (automatic via `init()`)
   - Services register themselves when imported
   - Import in `cmd/agent/main.go` or via blank import

### Adding a New Resource Type to Existing Service

1. **Create Discoverer** in `pkg/discovery/services/<servicename>.go`
2. **Create Auditor** in `pkg/audit/<servicename>/<resource>.go`
3. **Update Service** to return new discoverer/auditor
4. **Register Resource** in service's `init()` function
5. **Update Validator** in `pkg/policy/validation/<servicename>.go` to add validation rules for the new resource

## Data Flow

```
Policy File (YAML)
    ↓
Policy Loader
    ↓
Orchestrator
    ↓
    ├─→ Service Registry → Get Service
    │                        ↓
    │                   Get Discoverer → Discover Resources → Jobs Channel
    │                        ↓
    │                   Get Auditor → Check Resources → Results Channel
    │                        ↓
    │                   Remediation (if needed)
    ↓
Results → Reporter → JSONL Output
```

## Concurrency Model

- **Discovery**: Each resource type has its own goroutine
- **Workers**: Configurable worker pool processes jobs
- **Results**: Buffered channel (100 items) for result aggregation
- **Shutdown**: Context-based cancellation with graceful shutdown

## Best Practices

1. **Service Registration**: Always use `MustRegister()` in `init()`
2. **Resource Registration**: Register all resources in service's `init()`
3. **Error Handling**: Return errors, don't panic (except in `init()`)
4. **Context Handling**: Always respect context cancellation
5. **Type Safety**: Use type assertions with error checking
6. **Documentation**: Document supported resources in service comments

## Testing

- **Unit Tests**: Test each auditor and discoverer independently
- **Integration Tests**: Test service registration and lookup
- **E2E Tests**: Test full workflow with real OpenStack (tagged with `e2e`)

## Examples

See existing implementations:
- Nova: `pkg/services/services/nova.go`, `pkg/discovery/services/nova.go`, `pkg/audit/nova/`
- Neutron: `pkg/services/services/neutron.go`, `pkg/discovery/services/neutron.go`, `pkg/audit/neutron/`
- Cinder: `pkg/services/services/cinder.go`, `pkg/discovery/services/cinder.go`, `pkg/audit/cinder/`

