# Writing Policies

Policies define what OSPA audits and how it responds to violations. This guide covers the policy format and best practices.

## Policy Structure

A policy file is a YAML document with the following structure:

```yaml
version: v1
defaults:
  workers: <number-of-concurrent-workers>
  output: findings.json
policies:
  - <service>:
    - name: <rule-name>
      description: <description>
      service: <service>
      resource: <resource-type>
      check:
        <condition>: <value>
      action: <action>
```

## Top-Level Fields

### version

Required. Currently only `v1` is supported.

```yaml
version: v1
```

### defaults

Optional. Global defaults for all rules.

| Field | Type | Description |
|-------|------|-------------|
| `workers` | int | Number of concurrent workers (default: 16) |
| `output` | string | Default output file path |

```yaml
defaults:
  workers: 50
  output: findings.json
```

### policies

Required. List of service policy groups.

```yaml
policies:
  - neutron:
    - name: rule-1
      # ...
  - nova:
    - name: rule-2
      # ...
```

## Rule Definition

Each rule has these fields:

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier for the rule |
| `description` | No | Human-readable description |
| `service` | Yes | OpenStack service (nova, neutron, cinder) |
| `resource` | Yes | Resource type to audit |
| `check` | Yes | Conditions to match |
| `action` | Yes | Action on violation (log, delete, tag) |
| `tag_name` | Conditional | Tag name when action is `tag` |

### Example Rule

```yaml
- name: critical-ssh-open-to-world
  description: Find SSH rules open to the internet
  service: neutron
  resource: security_group_rule
  check:
    direction: ingress
    protocol: tcp
    port: 22
    remote_ip_prefix: 0.0.0.0/0
  action: log
```

## Check Conditions

Check conditions filter which resources are flagged. See [Policy Schema](../reference/policy-schema.md) for the complete reference.

### Common Checks

```yaml
# Status-based
check:
  status: ACTIVE

# Age-based
check:
  age_gt: 30d

# Unused resources
check:
  unused: true

# Exempt by name pattern
check:
  status: ACTIVE
  exempt_names:
    - default
    - system-*
```

### Combining Checks

All conditions must match (AND logic):

```yaml
check:
  status: available
  age_gt: 7d
  # Both conditions must be true
```

## Actions

Actions define what happens when a resource matches:

| Action | Description |
|--------|-------------|
| `log` | Record the violation (no changes) |
| `delete` | Delete the resource (requires `--fix`) |
| `tag` | Tag the resource (requires `--fix`) |

### Log Action

Safe, read-only. Records violations without changes:

```yaml
action: log
```

### Delete Action

Removes matching resources. **Requires `--fix` flag**:

```yaml
action: delete
```

### Tag Action

Adds a tag to matching resources. **Requires `--fix` flag**:

```yaml
action: tag
tag_name: audit-flagged
```

## Complete Examples

### Security Audit Policy

```yaml
version: v1
defaults:
  workers: 50
policies:
  - neutron:
    - name: critical-ssh-open-to-world
      description: SSH port 22 open to 0.0.0.0/0
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: log

    - name: critical-rdp-open-to-world
      description: RDP port 3389 open to 0.0.0.0/0
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 3389
        remote_ip_prefix: 0.0.0.0/0
      action: log
```

### Cost Optimization Policy

```yaml
version: v1
defaults:
  workers: 50
policies:
  - neutron:
    - name: unused-floating-ips
      description: Unattached floating IPs
      service: neutron
      resource: floating_ip
      check:
        status: DOWN
      action: log

  - cinder:
    - name: unattached-volumes
      description: Available volumes older than 7 days
      service: cinder
      resource: volume
      check:
        status: available
        age_gt: 7d
      action: log

    - name: old-snapshots
      description: Snapshots older than 90 days
      service: cinder
      resource: snapshot
      check:
        age_gt: 90d
      action: delete
```

### Compliance Tagging Policy

```yaml
version: v1
policies:
  - nova:
    - name: stale-instances
      description: Instances running over 30 days
      service: nova
      resource: instance
      check:
        age_gt: 30d
        exempt_metadata:
          key: lifecycle
          value: permanent
      action: tag
      tag_name: audit-stale-instance
```

## Best Practices

### 1. Start with Log Actions

Always test policies with `action: log` first:

```yaml
action: log  # Safe, no changes
```

### 2. Use Descriptive Names

Make rule names meaningful:

```yaml
# Good
name: critical-ssh-open-to-world

# Avoid
name: rule-1
```

### 3. Add Descriptions

Document what each rule does:

```yaml
description: |
  Finds security group rules that allow SSH access from any IP.
  These should be restricted to specific trusted networks.
```

### 4. Use Exemptions

Protect known-good resources:

```yaml
check:
  unused: true
  exempt_names:
    - default
    - system-managed-*
```

### 5. Organize by Severity

Group rules by severity level:

```yaml
policies:
  - neutron:
    # Critical rules first
    - name: critical-ssh-open-to-world
      # ...
    
    # Then warnings
    - name: warning-icmp-open-to-world
      # ...
    
    # Then info
    - name: info-unused-security-groups
      # ...
```

## Validation

OSPA validates policies before running:

```bash
# The agent validates on startup
go run ./cmd/agent --policy policy.yaml --out /dev/null

# Check for errors
# 2024/01/15 10:30:00 Error: policy validation failed: ...
```

Common validation errors:

- Missing required fields
- Unknown service or resource type
- Invalid check conditions
- Missing `tag_name` for tag action

