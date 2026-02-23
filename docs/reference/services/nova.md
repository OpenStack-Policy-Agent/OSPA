# Policy Guide: Nova (nova)

This guide explains how to write policies for Nova resources in OSPA.

## Service Overview

**Service Name:** `nova`
**Display Name:** Nova
**OpenStack Service Type:** compute

## Supported Resources


### Instance

**Resource Type:** `instance`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names, image_name, no_keypair

#### Security & Domain Checks

| Check | Severity | Category | Type | Description |
|-------|----------|----------|------|-------------|
- **`image_name`** | medium | compliance | string_list | Instance uses a deprecated or banned image
- **`no_keypair`** | medium | security | bool | Instance has no SSH keypair attached


### Keypair

**Resource Type:** `keypair`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** age_gt, unused, exempt_names



## OpenStack Security Guide Checklist

The following items from the OpenStack Security Guide apply to Nova.
These are **configuration-level** checks that require manual verification on
the control plane (not API-auditable).

| ID | Description | Section | Manual |
|----|-------------|---------|--------|
- **Check-Compute-01** | User/group ownership of config files set to root/nova | compute/checklist | Yes
- **Check-Compute-02** | Strict permissions (640) on configuration files | compute/checklist | Yes
- **Check-Compute-03** | Keystone used for authentication | compute/checklist | Yes
- **Check-Compute-04** | Secure protocol (TLS) for authentication | compute/checklist | Yes
- **Check-Compute-05** | Nova communicates with Glance over TLS | compute/checklist | Yes



## Policy Structure

All policies for Nova follow this structure:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.json
policies:
  - nova:
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
  description: Find inactive nova resources
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
  description: Find unused nova resources
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
  description: Audit nova resources
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


### Instance Examples

#### Security Check Example

```yaml
- name: security-check-instance-image_name
  description: "Instance uses a deprecated or banned image"
  resource: instance
  severity: medium
  category: compliance
  check:
    image_name: true
  action: log
```


#### Find Inactive Instance Resources

```yaml
- name: find-inactive-instance
  description: Find inactive instance resources
  resource: instance
  check:
    status: inactive
  action: log
```

#### Find Old Instance Resources

```yaml
- name: find-old-instance
  description: Find instance resources older than 30 days
  resource: instance
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused Instance Resources

```yaml
- name: cleanup-unused-instance
  description: Delete unused instance resources
  resource: instance
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```


### Keypair Examples


#### Find Inactive Keypair Resources

```yaml
- name: find-inactive-keypair
  description: Find inactive keypair resources
  resource: keypair
  check:
    status: inactive
  action: log
```

#### Find Old Keypair Resources

```yaml
- name: find-old-keypair
  description: Find keypair resources older than 30 days
  resource: keypair
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused Keypair Resources

```yaml
- name: cleanup-unused-keypair
  description: Delete unused keypair resources
  resource: keypair
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```



## Complete Policy Example

Here's a complete policy file example for Nova:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.json
policies:
  - nova:
    - name: audit-instance
      description: Audit instance resources
      resource: instance
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-instance
      description: Find instance resources older than 90 days
      resource: instance
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-keypair
      description: Audit keypair resources
      resource: keypair
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-keypair
      description: Find keypair resources older than 90 days
      resource: keypair
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
```

## OpenStack Documentation References

For more information about Nova resources and their properties:

- **OpenStack Nova API Documentation:** https://docs.openstack.org/api-ref/nova/
- **Nova Service Guide:** https://docs.openstack.org/nova/latest/
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
- Ensure service name matches exactly: `nova`
- Verify resource type is supported: `{instance Server instances [status age_gt unused exempt_names image_name no_keypair] [{image_name string_list Instance uses a deprecated or banned image compliance medium } {no_keypair bool Instance has no SSH keypair attached security medium }] [log delete tag] {false false false}}`, `{keypair SSH keypairs [age_gt unused exempt_names] [] [log delete tag] {false false false}}`
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
