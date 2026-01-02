# Policy Guide: Cinder (cinder)

This guide explains how to write policies for Cinder resources in OSPA.

## Service Overview

**Service Name:** `cinder`
**Display Name:** Cinder
**OpenStack Service Type:** volumev3

## Supported Resources


### Volume

**Resource Type:** `volume`


### Snapshot

**Resource Type:** `snapshot`



## Policy Structure

All policies for Cinder follow this structure:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - cinder:
    - name: rule-name
      description: Rule description
      service: cinder
      resource: <resource_type>
      check:
        # Check conditions (see below)
      action: log|delete|tag
```

## Check Conditions

### Common Check Conditions

The following check conditions are available for most resources:

#### Status Check

Check resources by their status:

```yaml
check:
  status: active|inactive|available|unavailable|DOWN|UP
```

**Example:**
```yaml
- name: find-inactive-resources
  description: Find inactive cinder resources
  service: cinder
  resource: <resource_type>
  check:
    status: inactive
  action: log
```

#### Age Check

Find resources older than a specified age:

```yaml
check:
  age_gt: 30d  # Options: 7d, 30d, 90d, 1h, 24h, etc.
```

**Supported units:**
- `d` or `day` or `days` - Days
- `h` or `hour` or `hours` - Hours
- `m` or `min` or `minute` or `minutes` - Minutes

**Example:**
```yaml
- name: find-old-resources
  description: Find resources older than 30 days
  service: cinder
  resource: <resource_type>
  check:
    age_gt: 30d
  action: log
```

#### Unused Check

Find resources that are not being used:

```yaml
check:
  unused: true
```

**Example:**
```yaml
- name: find-unused-resources
  description: Find unused cinder resources
  service: cinder
  resource: <resource_type>
  check:
    unused: true
  action: log
```

#### Exemptions

Exclude specific resources from checks:

```yaml
check:
  status: active
  exempt_names:
    - default
    - system-resource
```

**Example:**
```yaml
- name: find-active-except-default
  description: Find active resources except default ones
  service: cinder
  resource: <resource_type>
  check:
    status: active
    exempt_names:
      - default
  action: log
```

## Actions

### Log Action

Log violations without taking any action:

```yaml
action: log
```

**Example:**
```yaml
- name: audit-resources
  description: Audit cinder resources
  service: cinder
  resource: <resource_type>
  check:
    status: inactive
  action: log
```

### Delete Action

Delete non-compliant resources (use with caution):

```yaml
action: delete
```

**Example:**
```yaml
- name: cleanup-old-resources
  description: Delete resources older than 90 days
  service: cinder
  resource: <resource_type>
  check:
    age_gt: 90d
  action: delete
```

**Note:** The `--apply` flag must be set when running the agent for delete actions to take effect.

### Tag Action

Tag non-compliant resources with metadata:

```yaml
action: tag
tag_name: audit-tag-name
action_tag_name: "Display Name for Tag"
```

**Example:**
```yaml
- name: tag-old-resources
  description: Tag resources older than 30 days
  service: cinder
  resource: <resource_type>
  check:
    age_gt: 30d
  action: tag
  tag_name: audit-old-resource
  action_tag_name: "Old Resource"
```

## Resource-Specific Examples


### Volume Examples

#### Example 1: Find Inactive Volume Resources

```yaml
- name: find-inactive-volume
  description: Find inactive volume resources
  service: cinder
  resource: volume
  check:
    status: inactive
  action: log
```

#### Example 2: Find Old Volume Resources

```yaml
- name: find-old-volume
  description: Find volume resources older than 30 days
  service: cinder
  resource: volume
  check:
    age_gt: 30d
  action: log
```

#### Example 3: Cleanup Unused Volume Resources

```yaml
- name: cleanup-unused-volume
  description: Delete unused volume resources
  service: cinder
  resource: volume
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```

#### Example 4: Tag Old Volume Resources

```yaml
- name: tag-old-volume
  description: Tag volume resources older than 7 days
  service: cinder
  resource: volume
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-volume
  action_tag_name: "Old Volume"
```


### Snapshot Examples

#### Example 1: Find Inactive Snapshot Resources

```yaml
- name: find-inactive-snapshot
  description: Find inactive snapshot resources
  service: cinder
  resource: snapshot
  check:
    status: inactive
  action: log
```

#### Example 2: Find Old Snapshot Resources

```yaml
- name: find-old-snapshot
  description: Find snapshot resources older than 30 days
  service: cinder
  resource: snapshot
  check:
    age_gt: 30d
  action: log
```

#### Example 3: Cleanup Unused Snapshot Resources

```yaml
- name: cleanup-unused-snapshot
  description: Delete unused snapshot resources
  service: cinder
  resource: snapshot
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```

#### Example 4: Tag Old Snapshot Resources

```yaml
- name: tag-old-snapshot
  description: Tag snapshot resources older than 7 days
  service: cinder
  resource: snapshot
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-snapshot
  action_tag_name: "Old Snapshot"
```



## Complete Policy Example

Here's a complete policy file example for Cinder:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - cinder:
    - name: audit-volume
      description: Audit volume resources
      service: cinder
      resource: volume
      check:
        status: active
      action: log
    - name: cleanup-old-volume
      description: Find volume resources older than 90 days
      service: cinder
      resource: volume
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-snapshot
      description: Audit snapshot resources
      service: cinder
      resource: snapshot
      check:
        status: active
      action: log
    - name: cleanup-old-snapshot
      description: Find snapshot resources older than 90 days
      service: cinder
      resource: snapshot
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
```

## OpenStack Documentation References

For more information about Cinder resources and their properties:

- **OpenStack Cinder API Documentation:** https://docs.openstack.org/api-ref/cinder/
- **Cinder Service Guide:** https://docs.openstack.org/cinder/latest/

## Testing Your Policy

1. **Validate the policy:**
   ```bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out /dev/null
   ```

2. **Run in audit mode (safe):**
   ```bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.jsonl
   ```

3. **Apply remediations (use with caution):**
   ```bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.jsonl --apply
   ```

## Notes

- All check conditions are optional, but at least one should be specified
- Multiple check conditions are combined with AND logic (all must match)
- The `exempt_names` list allows you to exclude specific resources by name
- Age checks use the resource's `UpdatedAt` timestamp, falling back to `CreatedAt` if not available
- Status values are case-sensitive and should match OpenStack API responses exactly

## Troubleshooting

**Policy validation fails:**
- Ensure service name matches exactly: `cinder`
- Verify resource type is supported: `volume`, `snapshot`
- Check YAML syntax is correct

**No resources found:**
- Verify resources exist in your OpenStack project
- Use `--all-tenants` flag if resources are in other projects (requires admin)
- Check OpenStack API endpoints are accessible

**Actions not working:**
- Ensure `--apply` flag is set for delete/tag actions
- Verify you have permissions to modify resources
- Check action-specific requirements (e.g., `tag_name` for tag action)

## See Also

- [OSPA Development Guide](../../docs/DEVELOPMENT.md)
- [OSPA Architecture Guide](../../docs/ARCHITECTURE.md)
- [Example Policies](../policies.yaml)
