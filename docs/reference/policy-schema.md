# Policy Schema Reference

This page documents the complete YAML schema for OSPA policy files.

## Schema Overview

```yaml
version: v1              # Required: schema version
defaults:                # Optional: global defaults
  workers: <int>
  output: <string>
policies:                # Required: list of service policies
  - <service>:           # Service name (neutron, nova, cinder, etc.)
    - name: <string>     # Rule name
      description: <string>
      resource: <string>
      check: <object>
      action: <string>
      tag_name: <string> # Required when action is "tag"
      severity: <string> # Optional: critical, high, medium, low
      category: <string> # Optional: security, compliance, cost, hygiene
      guide_ref: <string> # Optional: OpenStack Security Guide ref (e.g., Check-Block-09, OSSN-0011)
```

---

## Top-Level Fields

### version

**Required.** Schema version. Currently only `v1` is supported.

```yaml
version: v1
```

### defaults

**Optional.** Global defaults applied to all rules.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `workers` | int | 16 | Concurrent worker count |
| `output` | string | — | Default output file |

```yaml
defaults:
  workers: 50
  output: findings.json
```

### policies

**Required.** List of service policy blocks.

---

## Rule Fields

### name

**Required.** Unique identifier for the rule.

```yaml
name: unused-security-groups
```

### description

**Optional.** Human-readable description.

```yaml
description: Find security groups not attached to any ports
```

### resource

**Required.** Resource type to audit.

```yaml
resource: security_group
```

### check

**Required.** Conditions that trigger a violation.

```yaml
check:
  status: ERROR
  age_gt: 30d
  unused: true
```

### action

**Required.** Action to take on violation. One of: `log`, `tag`, `delete`.

```yaml
action: log
```

### tag_name

**Required when action is `tag`.** The tag value to apply.

```yaml
action: tag
tag_name: ospa-flagged
```

### severity

**Optional.** Classifies the severity of a finding. One of: `critical`, `high`, `medium`, `low`.

```yaml
severity: high
```

### category

**Optional.** Classifies the type of finding. One of: `security`, `compliance`, `cost`, `hygiene`.

```yaml
category: security
```

### guide_ref

**Optional.** Reference to an OpenStack Security Guide checklist item (e.g., `Check-Block-09`, `OSSN-0011`).

```yaml
guide_ref: Check-Block-09
```

---

## Check Conditions

### Universal Checks

Available for most resource types:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `status` | string | Match resource status | `status: ERROR` |
| `age_gt` | duration | Resource older than | `age_gt: 30d` |
| `unused` | bool | Resource not in use | `unused: true` |
| `exempt_names` | list | Skip matching names | `exempt_names: ["default", "system-*"]` |

### Duration Format

The `age_gt` field accepts durations in the format `<number><unit>`:

| Unit | Meaning | Example |
|------|---------|---------|
| `h` | Hours | `24h` |
| `d` | Days | `30d` |
| `w` | Weeks | `2w` |
| `m` | Months (30 days) | `6m` |
| `y` | Years (365 days) | `1y` |

### Security Group Rule Checks

Additional checks for `security_group_rule`:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `direction` | string | `ingress` or `egress` | `direction: ingress` |
| `ethertype` | string | `IPv4` or `IPv6` | `ethertype: IPv4` |
| `protocol` | string | Protocol name | `protocol: tcp` |
| `port` | int | Port number | `port: 22` |
| `remote_ip_prefix` | string | CIDR range | `remote_ip_prefix: "0.0.0.0/0"` |

### Neutron-Specific Checks

For various Neutron resources (ports, floating_ips, networks, etc.):

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `port_range_wide` | bool | Port range spans >100 ports | `port_range_wide: true` |
| `unassociated` | bool | Floating IP not attached to port | `unassociated: true` |
| `shared_network` | bool | Network shared across tenants | `shared_network: true` |
| `no_security_group` | bool | Port has no security groups | `no_security_group: true` |

### Nova-Specific Checks

For `instance` resource:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `image_name` | list | Deprecated/banned image names | `image_name: ["ubuntu-14*"]` |
| `no_keypair` | bool | No SSH keypair attached | `no_keypair: true` |

### Cinder-Specific Checks

For `volume` and `snapshot` resources:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `encrypted` | bool ptr | Volume/snapshot not encrypted | `encrypted: false` |
| `attached` | bool ptr | Volume not attached to instance | `attached: false` |
| `has_backup` | bool ptr | Volume has no backup | `has_backup: false` |

### Glance-Specific Checks

For `image` resource:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `visibility` | string | Image visibility level | `visibility: public` |

### Keystone-Specific Checks

For `user` resource:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `password_expired` | bool | User password has expired | `password_expired: true` |
| `mfa_enabled` | bool ptr | User MFA not enabled | `mfa_enabled: false` |
| `inactive_days` | int | Inactive for N days | `inactive_days: 90` |
| `has_admin_role` | bool | User has admin role | `has_admin_role: true` |
| `token_provider` | string | Token provider type | `token_provider: fernet` |

---

## Action Types

### log

Report the violation without making changes.

```yaml
action: log
```

### tag

Add a tag to the resource. Requires `tag_name`.

```yaml
action: tag
tag_name: ospa-reviewed
```

### delete

Delete the resource. **Destructive action.**

```yaml
action: delete
```

---

## Complete Example

```yaml
version: v1

defaults:
  workers: 50
  output: findings.json

policies:
  - neutron:
    - name: unused-security-groups
      description: Find security groups not attached to any ports
      resource: security_group
      severity: medium
      category: hygiene
      check:
        unused: true
        exempt_names:
          - default
          - "system-*"
      action: log

    - name: ssh-open-to-world
      description: Flag rules allowing SSH from anywhere
      resource: security_group_rule
      severity: high
      category: security
      guide_ref: OSSN-0011
      check:
        direction: ingress
        protocol: tcp
        port: 22
        remote_ip_prefix: "0.0.0.0/0"
      action: tag
      tag_name: ospa-insecure

  - nova:
    - name: error-instances
      description: Find instances in ERROR state
      resource: instance
      severity: critical
      check:
        status: ERROR
      action: log

    - name: old-instances
      description: Find instances older than 1 year
      resource: instance
      severity: low
      category: cost
      check:
        age_gt: 1y
      action: log
```

---

## Validation

Policies are validated at load time. Common validation errors:

| Error | Cause | Fix |
|-------|-------|-----|
| `unknown service` | Invalid service name | Check service name spelling |
| `unknown resource` | Invalid resource for service | Check resource name |
| `tag_name required` | Action is `tag` but no tag_name | Add `tag_name` field |
| `invalid check` | Unsupported check for resource | See resource-specific checks |

