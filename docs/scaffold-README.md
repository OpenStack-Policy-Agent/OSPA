# OSPA Scaffold Tool

The scaffold tool generates boilerplate code for adding new OpenStack services to OSPA.

## Usage

**Using Make (recommended):**
```bash
make scaffold SERVICE=<name> RESOURCES=<list> [DISPLAY_NAME=<name>] [TYPE=<type>]
```

**Or directly:**
```bash
go run ./cmd/scaffold --service <name> [options]
```

**Or build and install:**
```bash
go build -o ospa-scaffold ./cmd/scaffold
./ospa-scaffold --service <name> [options]
```

## Options

- `--service` (required): Service name (e.g., `glance`, `keystone`)
- `--display-name`: Display name for the service (defaults to service display name from registry)
- `--resources`: Comma-separated list of resource types (e.g., `image,member`)
- `--type`: OpenStack service type for client creation (defaults to service type from registry)
- `--force`: Overwrite existing files
- `--list`: List all available OpenStack services and their resources

## Examples

### Generate Glance (Image Service) support

```bash
go run ./cmd/scaffold \
  --service glance \
  --display-name Glance \
  --resources image,member \
  --type image
```

This generates:
- `pkg/services/services/glance.go`
- `pkg/discovery/services/glance.go`
- `pkg/audit/glance/image.go`
- `pkg/audit/glance/member.go`
- `pkg/auth/auth.go` (client method appended)
- `pkg/policy/validation/glance.go` (validation file)
- `pkg/audit/glance/image_test.go`
- `pkg/audit/glance/member_test.go`
- `e2e/glance_test.go`
- `examples/policies/glance-policy-guide.md`

### Generate Keystone (Identity Service) support

```bash
go run ./cmd/scaffold \
  --service keystone \
  --display-name Keystone \
  --resources user,role,project \
  --type identity
```

### List Available Services

To see all available OpenStack services and their resources:

```bash
go run ./cmd/scaffold --list
```

This will display all supported OpenStack services (Nova, Neutron, Cinder, Glance, Keystone, etc.) along with their available resource types.

## Validation

The scaffold tool validates that:
- The service name exists in OpenStack (validated against known OpenStack services)
- All specified resources exist for the given service
- Service type matches OpenStack conventions

If validation fails, the tool will:
- Show an error message with available options
- Suggest using `--list` to see all available services
- Display available resources for the specified service if the service is valid but resources are not

## Generated Files

The tool generates:

1. **Service file** (`pkg/services/services/<servicename>.go`)
   - Service implementation
   - Resource registration
   - Client, auditor, and discoverer methods

2. **Discovery file** (`pkg/discovery/services/<servicename>.go`)
   - Discoverer implementations for each resource type
   - Handles pagination and context cancellation

3. **Auditor files** (`pkg/audit/<servicename>/<resource>.go`)
   - Auditor implementation for each resource type
   - Check() and Fix() methods

4. **Auth client method** (appended to `pkg/auth/auth.go`)
   - `Get<DisplayName>Client()` method for service authentication

5. **Unit test files** (`pkg/audit/<servicename>/<resource>_test.go`)
   - Basic unit tests for each auditor
   - Tests for ResourceType(), Check(), and Fix() methods

6. **E2E test file** (`e2e/<servicename>_test.go`)
   - End-to-end tests using the e2e engine
   - Test functions for each resource type

7. **Validation file** (`pkg/policy/validation/<servicename>.go`)
   - Service-specific policy validator
   - Validates check conditions for each resource type
   - Auto-registers with the validation registry
   - Automatically imported in `validator.go`

8. **Policy guide** (`examples/policies/<servicename>-policy-guide.md`)
   - Comprehensive guide for writing policies for the new service
   - Examples for each resource type
   - Check conditions documentation
   - OpenStack API references

## Next Steps

After generating files:

1. **Review generated code**:
   - Adjust import paths for OpenStack client libraries
   - Update resource struct names and field names
   - Customize check conditions based on resource type
   - Implement validation rules in `pkg/policy/validation/<servicename>.go` (replace TODO comments)
   - Implement tagging logic if needed

2. **Test**:
   ```bash
   go test ./pkg/services/...
   go test ./pkg/audit/<servicename>/...
   ```

3. **Review policy guide**:
   - Check `examples/policies/<servicename>-policy-guide.md`
   - Follow the examples to create your first policy
   - Reference OpenStack API documentation links provided

4. **Create policy**:
   ```yaml
   policies:
     - <servicename>:
       - name: my-rule
         description: Check resource
         service: <servicename>
         resource: <resource>
         check:
           status: active
         action: log
   ```

## Architecture

The scaffold tool is organized into a modular structure:

- **`cmd/scaffold/main.go`** - CLI entry point with validation
- **`cmd/scaffold/internal/registry/`** - OpenStack service/resource registry and validation
  - `registry.go`
- **`cmd/scaffold/internal/generators/`** - Generation functions organized by step:
  - `orchestrator.go` - Main orchestration function
  - `service.go` - Service file generation
  - `service_updater.go` - Add new resources to existing service files
  - `discovery.go` - Discovery file generation
  - `discovery_updater.go` - Add new discoverers to existing discovery files
  - `auditor.go` - Auditor files generation
  - `auth.go` - Auth method generation
  - `validation.go` - Validation file generation
  - `tests.go` - Unit tests generation
  - `e2e.go` - E2E tests generation
  - `policy_guide.go` - Policy guide generation
  - `utils.go` - Utility functions

## Notes

- Generated code uses common patterns but may need customization
- Import paths for OpenStack client libraries may need adjustment
- Resource struct field names (e.g., `TenantID`, `Name`) may vary
- Some resources may not support all check conditions
- Tagging implementation depends on resource type capabilities
- The tool validates services and resources against a registry of known OpenStack services
- Use `--list` to discover available services and resources before generating code

