# Policy Guide: Cinder (cinder)

This guide explains how to write policies for Cinder resources in OSPA.

## Service Overview

**Service Name:** `cinder`
**Display Name:** Cinder
**OpenStack Service Type:** volumev3

## Supported Resources


### Volume

**Resource Type:** `volume`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names, encrypted, attached, has_backup

#### Security & Domain Checks

| Check | Severity | Category | Type | Description |
|-------|----------|----------|------|-------------|
- **`encrypted`** | high | security | bool | Volume is not encrypted _(Ref: Check-Block-09)_
- **`attached`** | medium | cost | bool | Volume is not attached to any instance
- **`has_backup`** | medium | compliance | bool | Volume has no backup


### Snapshot

**Resource Type:** `snapshot`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names, encrypted

#### Security & Domain Checks

| Check | Severity | Category | Type | Description |
|-------|----------|----------|------|-------------|
- **`encrypted`** | high | security | bool | Snapshot is not encrypted



## OpenStack Security Guide Checklist

The following items from the OpenStack Security Guide apply to Cinder.
These are **configuration-level** checks that require manual verification on
the control plane (not API-auditable).

| ID | Description | Section | Manual |
|----|-------------|---------|--------|
- **Check-Block-01** | User/group ownership of config files set to root/cinder | block-storage/checklist | Yes
- **Check-Block-02** | Strict permissions (640) on configuration files | block-storage/checklist | Yes
- **Check-Block-03** | Keystone used for authentication | block-storage/checklist | Yes
- **Check-Block-04** | TLS enabled for authentication | block-storage/checklist | Yes
- **Check-Block-05** | Cinder communicates with Nova over TLS | block-storage/checklist | Yes
- **Check-Block-06** | Cinder communicates with Glance over TLS | block-storage/checklist | Yes
- **Check-Block-07** | NAS operating in a secure environment | block-storage/checklist | Yes
- **Check-Block-08** | Max request body size set to default (114688) | block-storage/checklist | Yes
- **Check-Block-09** | Volume encryption feature enabled | block-storage/checklist | Yes



## Policy Structure

All policies for Cinder follow this structure:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.json
policies:
  - cinder:
    - name: rule-name
      description: Rule description
      resource: <resource_type>
      severity: critical|high|medium|low
      category: security|compliance|cost|hygiene
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
  resource: <resource_type>
  check:
    age_gt: 90d
  action: delete
```

**Note:** The `--fix` flag must be set when running the agent for delete actions to take effect.

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
  resource: <resource_type>
  check:
    age_gt: 30d
  action: tag
  tag_name: audit-old-resource
  action_tag_name: "Old Resource"
```

## Resource-Specific Examples


### Volume Examples

#### Security Check Example

```yaml
- name: security-check-volume-encrypted
  description: "Volume is not encrypted"
  resource: volume
  severity: high
  category: security
  guide_ref: "Check-Block-09"
  check:
    encrypted: true
  action: log
```


#### Find Inactive Volume Resources

```yaml
- name: find-inactive-volume
  description: Find inactive volume resources
  resource: volume
  check:
    status: inactive
  action: log
```

#### Find Old Volume Resources

```yaml
- name: find-old-volume
  description: Find volume resources older than 30 days
  resource: volume
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused Volume Resources

```yaml
- name: cleanup-unused-volume
  description: Delete unused volume resources
  resource: volume
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```


### Snapshot Examples

#### Security Check Example

```yaml
- name: security-check-snapshot-encrypted
  description: "Snapshot is not encrypted"
  resource: snapshot
  severity: high
  category: security
  check:
    encrypted: true
  action: log
```


#### Find Inactive Snapshot Resources

```yaml
- name: find-inactive-snapshot
  description: Find inactive snapshot resources
  resource: snapshot
  check:
    status: inactive
  action: log
```

#### Find Old Snapshot Resources

```yaml
- name: find-old-snapshot
  description: Find snapshot resources older than 30 days
  resource: snapshot
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused Snapshot Resources

```yaml
- name: cleanup-unused-snapshot
  description: Delete unused snapshot resources
  resource: snapshot
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```



## Complete Policy Example

Here's a complete policy file example for Cinder:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.json
policies:
  - cinder:
    - name: audit-volume
      description: Audit volume resources
      resource: volume
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-volume
      description: Find volume resources older than 90 days
      resource: volume
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-snapshot
      description: Audit snapshot resources
      resource: snapshot
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-snapshot
      description: Find snapshot resources older than 90 days
      resource: snapshot
      severity: low
      category: cost
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
- **OpenStack Security Guide:** https://docs.openstack.org/security-guide/

## Testing Your Policy

1. **Validate the policy:**
   ```bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out /dev/null
   ```

2. **Run in audit mode (safe):**
   ```bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.json
   ```

3. **Apply remediations (use with caution):**
   ```bash
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.json --fix
   ```

## Notes

- All check conditions are optional, but at least one should be specified
- Multiple check conditions are combined with AND logic (all must match)
- The `exempt_names` list allows you to exclude specific resources by name
- Age checks use the resource's `UpdatedAt` timestamp, falling back to `CreatedAt` if not available
- Status values are case-sensitive and should match OpenStack API responses exactly
- Use `severity` and `category` to classify findings for prioritization

## Troubleshooting

**Policy validation fails:**
- Ensure service name matches exactly: `cinder`
- Verify resource type is supported: `{volume Block storage volumes [status age_gt unused exempt_names encrypted attached has_backup] [{encrypted bool Volume is not encrypted security high Check-Block-09} {attached bool Volume is not attached to any instance cost medium } {has_backup bool Volume has no backup compliance medium }] [log delete tag] {false false false}}`, `{snapshot Volume snapshots [status age_gt unused exempt_names encrypted] [{encrypted bool Snapshot is not encrypted security high }] [log delete tag] {false false false}}`
- Check YAML syntax is correct

**No resources found:**
- Verify resources exist in your OpenStack project
- Use `--all-tenants` flag if resources are in other projects (requires admin)
- Check OpenStack API endpoints are accessible

**Actions not working:**
- Ensure `--fix` flag is set for delete/tag actions
- Verify you have permissions to modify resources
- Check action-specific requirements (e.g., `tag_name` for tag action)

## See Also

- [OSPA Development Guide](../../developer-guide/index.md)
- [OSPA Architecture Guide](../../developer-guide/architecture.md)
- [Example Policies](https://github.com/OpenStack-Policy-Agent/OSPA/blob/main/examples/policies.yaml)
