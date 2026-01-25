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
      check:
        unused: true
        exempt_names:
          - default
          - "system-*"
      action: log

    - name: ssh-open-to-world
      description: Flag rules allowing SSH from anywhere
      resource: security_group_rule
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
      check:
        status: ERROR
      action: log

    - name: old-instances
      description: Find instances older than 1 year
      resource: instance
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

