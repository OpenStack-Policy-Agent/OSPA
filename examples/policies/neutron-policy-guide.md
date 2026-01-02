# Policy Guide: Neutron (neutron)

This guide explains how to write policies for Neutron resources in OSPA.

## Service Overview

**Service Name:** `neutron`
**Display Name:** Neutron
**OpenStack Service Type:** network

## Supported Resources


### Security_group_rule

**Resource Type:** `security_group_rule`


### Floating_ip

**Resource Type:** `floating_ip`


### Security_group

**Resource Type:** `security_group`



## Policy Structure

All policies for Neutron follow this structure:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - neutron:
    - name: rule-name
      description: Rule description
      service: neutron
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
  description: Find inactive neutron resources
  service: neutron
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
  service: neutron
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
  description: Find unused neutron resources
  service: neutron
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
  service: neutron
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
  description: Audit neutron resources
  service: neutron
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
  service: neutron
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
  service: neutron
  resource: <resource_type>
  check:
    age_gt: 30d
  action: tag
  tag_name: audit-old-resource
  action_tag_name: "Old Resource"
```

## Resource-Specific Examples


### Security_group_rule Examples

#### Example 1: Find Inactive Security_group_rule Resources

```yaml
- name: find-inactive-security_group_rule
  description: Find inactive security_group_rule resources
  service: neutron
  resource: security_group_rule
  check:
    status: inactive
  action: log
```

#### Example 2: Find Old Security_group_rule Resources

```yaml
- name: find-old-security_group_rule
  description: Find security_group_rule resources older than 30 days
  service: neutron
  resource: security_group_rule
  check:
    age_gt: 30d
  action: log
```

#### Example 3: Cleanup Unused Security_group_rule Resources

```yaml
- name: cleanup-unused-security_group_rule
  description: Delete unused security_group_rule resources
  service: neutron
  resource: security_group_rule
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```

#### Example 4: Tag Old Security_group_rule Resources

```yaml
- name: tag-old-security_group_rule
  description: Tag security_group_rule resources older than 7 days
  service: neutron
  resource: security_group_rule
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-security_group_rule
  action_tag_name: "Old Security_group_rule"
```


### Floating_ip Examples

#### Example 1: Find Inactive Floating_ip Resources

```yaml
- name: find-inactive-floating_ip
  description: Find inactive floating_ip resources
  service: neutron
  resource: floating_ip
  check:
    status: inactive
  action: log
```

#### Example 2: Find Old Floating_ip Resources

```yaml
- name: find-old-floating_ip
  description: Find floating_ip resources older than 30 days
  service: neutron
  resource: floating_ip
  check:
    age_gt: 30d
  action: log
```

#### Example 3: Cleanup Unused Floating_ip Resources

```yaml
- name: cleanup-unused-floating_ip
  description: Delete unused floating_ip resources
  service: neutron
  resource: floating_ip
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```

#### Example 4: Tag Old Floating_ip Resources

```yaml
- name: tag-old-floating_ip
  description: Tag floating_ip resources older than 7 days
  service: neutron
  resource: floating_ip
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-floating_ip
  action_tag_name: "Old Floating_ip"
```


### Security_group Examples

#### Example 1: Find Inactive Security_group Resources

```yaml
- name: find-inactive-security_group
  description: Find inactive security_group resources
  service: neutron
  resource: security_group
  check:
    status: inactive
  action: log
```

#### Example 2: Find Old Security_group Resources

```yaml
- name: find-old-security_group
  description: Find security_group resources older than 30 days
  service: neutron
  resource: security_group
  check:
    age_gt: 30d
  action: log
```

#### Example 3: Cleanup Unused Security_group Resources

```yaml
- name: cleanup-unused-security_group
  description: Delete unused security_group resources
  service: neutron
  resource: security_group
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```

#### Example 4: Tag Old Security_group Resources

```yaml
- name: tag-old-security_group
  description: Tag security_group resources older than 7 days
  service: neutron
  resource: security_group
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-security_group
  action_tag_name: "Old Security_group"
```



## Complete Policy Example

Here's a complete policy file example for Neutron:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - neutron:
    - name: audit-security_group_rule
      description: Audit security_group_rule resources
      service: neutron
      resource: security_group_rule
      check:
        status: active
      action: log
    - name: cleanup-old-security_group_rule
      description: Find security_group_rule resources older than 90 days
      service: neutron
      resource: security_group_rule
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-floating_ip
      description: Audit floating_ip resources
      service: neutron
      resource: floating_ip
      check:
        status: active
      action: log
    - name: cleanup-old-floating_ip
      description: Find floating_ip resources older than 90 days
      service: neutron
      resource: floating_ip
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-security_group
      description: Audit security_group resources
      service: neutron
      resource: security_group
      check:
        status: active
      action: log
    - name: cleanup-old-security_group
      description: Find security_group resources older than 90 days
      service: neutron
      resource: security_group
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
```

## OpenStack Documentation References

For more information about Neutron resources and their properties:

- **OpenStack Neutron API Documentation:** https://docs.openstack.org/api-ref/neutron/
- **Neutron Service Guide:** https://docs.openstack.org/neutron/latest/

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
   go run ./cmd/agent --cloud "$OS_CLOUD" --policy your-policy.yaml --out findings.jsonl --fix
   ```

## Notes

- All check conditions are optional, but at least one should be specified
- Multiple check conditions are combined with AND logic (all must match)
- The `exempt_names` list allows you to exclude specific resources by name
- Age checks use the resource's `UpdatedAt` timestamp, falling back to `CreatedAt` if not available
- Status values are case-sensitive and should match OpenStack API responses exactly

## Troubleshooting

**Policy validation fails:**
- Ensure service name matches exactly: `neutron`
- Verify resource type is supported: `security_group_rule`, `floating_ip`, `security_group`
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

- [OSPA Development Guide](../../docs/DEVELOPMENT.md)
- [OSPA Architecture Guide](../../docs/ARCHITECTURE.md)
- [Example Policies](../policies.yaml)
