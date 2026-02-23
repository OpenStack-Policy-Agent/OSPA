# Policy Guide: Neutron (neutron)

This guide explains how to write policies for Neutron resources in OSPA.

## Service Overview

**Service Name:** `neutron`
**Display Name:** Neutron
**OpenStack Service Type:** network

## Supported Resources


### Network

**Resource Type:** `network`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names, shared_network

#### Security & Domain Checks

| Check | Severity | Category | Type | Description |
|-------|----------|----------|------|-------------|
- **`shared_network`** | high | security | bool | Network is shared across tenants without RBAC


### SecurityGroup

**Resource Type:** `security_group`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names


### SecurityGroupRule

**Resource Type:** `security_group_rule`

**Allowed Actions:** log, delete
**Allowed Checks:** direction, ethertype, protocol, port, remote_ip_prefix, port_range_wide, exempt_names

#### Security & Domain Checks

| Check | Severity | Category | Type | Description |
|-------|----------|----------|------|-------------|
- **`direction`** | medium | security | string | Traffic direction (ingress/egress)
- **`ethertype`** | low | security | string | Ethernet type (IPv4/IPv6)
- **`protocol`** | medium | security | string | IP protocol (tcp/udp/icmp)
- **`port`** | high | security | int | Port number within port range
- **`remote_ip_prefix`** | critical | security | cidr | Source/destination CIDR - 0.0.0.0/0 means open to world _(Ref: OSSN-0011)_
- **`port_range_wide`** | high | security | bool | Port range spans more than 100 ports
- **`exempt_names`** | low | hygiene | string_list | Exempt by security group ID pattern


### FloatingIp

**Resource Type:** `floating_ip`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, unassociated, exempt_names

#### Security & Domain Checks

| Check | Severity | Category | Type | Description |
|-------|----------|----------|------|-------------|
- **`unassociated`** | medium | cost | bool | Floating IP not attached to any port


### Subnet

**Resource Type:** `subnet`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names


### Router

**Resource Type:** `router`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names



## OpenStack Security Guide Checklist

The following items from the OpenStack Security Guide apply to Neutron.
These are **configuration-level** checks that require manual verification on
the control plane (not API-auditable).

| ID | Description | Section | Manual |
|----|-------------|---------|--------|
- **Check-Neutron-01** | User/group ownership of config files set to root/neutron | networking/checklist | Yes
- **Check-Neutron-02** | Strict permissions (640) on configuration files | networking/checklist | Yes
- **Check-Neutron-03** | Keystone used for authentication | networking/checklist | Yes
- **Check-Neutron-04** | Secure protocol (TLS) for authentication | networking/checklist | Yes
- **Check-Neutron-05** | TLS enabled on Neutron API server | networking/checklist | Yes



## Policy Structure

All policies for Neutron follow this structure:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.json
policies:
  - neutron:
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
  description: Find inactive neutron resources
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
  description: Find unused neutron resources
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
  description: Audit neutron resources
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


### Network Examples

#### Security Check Example

```yaml
- name: security-check-network-shared_network
  description: "Network is shared across tenants without RBAC"
  resource: network
  severity: high
  category: security
  check:
    shared_network: true
  action: log
```


#### Find Inactive Network Resources

```yaml
- name: find-inactive-network
  description: Find inactive network resources
  resource: network
  check:
    status: inactive
  action: log
```

#### Find Old Network Resources

```yaml
- name: find-old-network
  description: Find network resources older than 30 days
  resource: network
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused Network Resources

```yaml
- name: cleanup-unused-network
  description: Delete unused network resources
  resource: network
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```


### SecurityGroup Examples


#### Find Inactive SecurityGroup Resources

```yaml
- name: find-inactive-security_group
  description: Find inactive security_group resources
  resource: security_group
  check:
    status: inactive
  action: log
```

#### Find Old SecurityGroup Resources

```yaml
- name: find-old-security_group
  description: Find security_group resources older than 30 days
  resource: security_group
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused SecurityGroup Resources

```yaml
- name: cleanup-unused-security_group
  description: Delete unused security_group resources
  resource: security_group
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```


### SecurityGroupRule Examples

#### Security Check Example

```yaml
- name: security-check-security_group_rule-direction
  description: "Traffic direction (ingress/egress)"
  resource: security_group_rule
  severity: medium
  category: security
  check:
    direction: "value"
  action: log
```


#### Find Inactive SecurityGroupRule Resources

```yaml
- name: find-inactive-security_group_rule
  description: Find inactive security_group_rule resources
  resource: security_group_rule
  check:
    status: inactive
  action: log
```

#### Find Old SecurityGroupRule Resources

```yaml
- name: find-old-security_group_rule
  description: Find security_group_rule resources older than 30 days
  resource: security_group_rule
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused SecurityGroupRule Resources

```yaml
- name: cleanup-unused-security_group_rule
  description: Delete unused security_group_rule resources
  resource: security_group_rule
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```


### FloatingIp Examples

#### Security Check Example

```yaml
- name: security-check-floating_ip-unassociated
  description: "Floating IP not attached to any port"
  resource: floating_ip
  severity: medium
  category: cost
  check:
    unassociated: true
  action: log
```


#### Find Inactive FloatingIp Resources

```yaml
- name: find-inactive-floating_ip
  description: Find inactive floating_ip resources
  resource: floating_ip
  check:
    status: inactive
  action: log
```

#### Find Old FloatingIp Resources

```yaml
- name: find-old-floating_ip
  description: Find floating_ip resources older than 30 days
  resource: floating_ip
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused FloatingIp Resources

```yaml
- name: cleanup-unused-floating_ip
  description: Delete unused floating_ip resources
  resource: floating_ip
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```


### Subnet Examples


#### Find Inactive Subnet Resources

```yaml
- name: find-inactive-subnet
  description: Find inactive subnet resources
  resource: subnet
  check:
    status: inactive
  action: log
```

#### Find Old Subnet Resources

```yaml
- name: find-old-subnet
  description: Find subnet resources older than 30 days
  resource: subnet
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused Subnet Resources

```yaml
- name: cleanup-unused-subnet
  description: Delete unused subnet resources
  resource: subnet
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```



### Router Examples


#### Find Inactive Router Resources

```yaml
- name: find-inactive-router
  description: Find inactive router resources
  resource: router
  check:
    status: inactive
  action: log
```

#### Find Old Router Resources

```yaml
- name: find-old-router
  description: Find router resources older than 30 days
  resource: router
  check:
    age_gt: 30d
  action: log
```

#### Cleanup Unused Router Resources

```yaml
- name: cleanup-unused-router
  description: Delete unused router resources
  resource: router
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```




## Complete Policy Example

Here's a complete policy file example for Neutron:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.json
policies:
  - neutron:
    - name: audit-network
      description: Audit network resources
      resource: network
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-network
      description: Find network resources older than 90 days
      resource: network
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-security_group
      description: Audit security_group resources
      resource: security_group
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-security_group
      description: Find security_group resources older than 90 days
      resource: security_group
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-security_group_rule
      description: Audit security_group_rule resources
      resource: security_group_rule
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-security_group_rule
      description: Find security_group_rule resources older than 90 days
      resource: security_group_rule
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-floating_ip
      description: Audit floating_ip resources
      resource: floating_ip
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-floating_ip
      description: Find floating_ip resources older than 90 days
      resource: floating_ip
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-subnet
      description: Audit subnet resources
      resource: subnet
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-subnet
      description: Find subnet resources older than 90 days
      resource: subnet
      severity: low
      category: cost
      check:
        age_gt: 90d
        exempt_names:
          - default
      action: log
    - name: audit-router
      description: Audit router resources
      resource: router
      severity: medium
      category: hygiene
      check:
        status: active
      action: log
    - name: cleanup-old-router
      description: Find router resources older than 90 days
      resource: router
      severity: low
      category: cost
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
- Ensure service name matches exactly: `neutron`
- Verify resource type is supported: `network`, `security_group`, `security_group_rule`, `floating_ip`, `subnet`, `router`
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
