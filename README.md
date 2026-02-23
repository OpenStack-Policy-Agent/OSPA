# OSPA (OpenStack Policy Agent)

A policy-driven audit and remediation agent for OpenStack clouds.

**Define** policies in YAML → **Discover** resources → **Audit** against rules → **Remediate** violations.

## Features

- **Declarative policies** - Write audit rules in simple YAML
- **Multi-service support** - Audit resources across OpenStack services
- **Safe by default** - Read-only mode unless explicitly enabled
- **Extensible** - Add new services and resources with scaffolding tools
- **Concurrent** - Parallel discovery and audit for large clouds

## Quick Start

```bash
# Set up credentials
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml

# Audit only (safe, no changes)
go run ./cmd/agent --cloud mycloud --policy ./examples/policies.yaml --out findings.json

# Apply remediation (with --fix flag)
go run ./cmd/agent --cloud mycloud --policy ./examples/policies.yaml --out findings.json --fix
```

## Supported Services

| Service | Description | Status |
|---------|-------------|--------|
| **Neutron** | Networking | ✔ Implemented |
| **Nova** | Compute | ◐ Partial |
| **Cinder** | Block Storage | ◐ Partial |
| **Glance** | Image | — |
| **Keystone** | Identity | — |
| **Heat** | Orchestration | — |
| **Swift** | Object Storage | — |
| **Octavia** | Load Balancing | — |
| **Barbican** | Key Manager | — |
| **Manila** | Shared File Systems | — |
| **Trove** | Database | — |
| **Magnum** | Container Infrastructure | — |
| **Ironic** | Bare Metal | — |
| **Designate** | DNS | — |
| **Senlin** | Clustering | — |
| **Zaqar** | Messaging | — |

**Legend:** ✔ Fully implemented | ◐ Partial (some resources) | — Not yet

See the [full resource catalog](https://openstack-policy-agent.github.io/OSPA/reference/catalog/) for detailed resource support.

## Example Policy

```yaml
version: v1
policies:
  - neutron:
    - name: unused-security-groups
      description: Find security groups not attached to any ports
      resource: security_group
      check:
        unused: true
      action: log

    - name: dangerous-ingress-rules
      description: Flag rules allowing SSH from anywhere
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 22
        remote_ip_prefix: "0.0.0.0/0"
      action: log
```

## Documentation

- [Getting Started](https://openstack-policy-agent.github.io/OSPA/getting-started/)
- [User Guide](https://openstack-policy-agent.github.io/OSPA/user-guide/)
- [Developer Guide](https://openstack-policy-agent.github.io/OSPA/developer-guide/)
- [Reference](https://openstack-policy-agent.github.io/OSPA/reference/)

## Tests

```bash
# Unit tests
go test ./... -count=1

# E2E tests (requires OpenStack cloud)
export OS_CLOUD=mycloud
go test -tags=e2e ./e2e/... -count=1
```

## Extending OSPA

```bash
# List available services and resources
go run ./cmd/scaffold --list

# Generate scaffolding for a new service
go run ./cmd/scaffold --service glance --resources image,member
```

See the [Developer Guide](https://openstack-policy-agent.github.io/OSPA/developer-guide/) for details.

## License

Apache 2.0
