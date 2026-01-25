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
**Allowed Checks:** status, age_gt, unused, exempt_names


### SecurityGroup

**Resource Type:** `security_group`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names

**Implementation Status:** ✅ Fully implemented

Security groups can be audited for:
- Unused security groups (not attached to any ports)
- Age-based checks
- Name-based exemptions (e.g., exempt the `default` security group)


### SecurityGroupRule

**Resource Type:** `security_group_rule`

**Allowed Actions:** log, delete
**Allowed Checks:** direction, ethertype, protocol, port, remote_ip_prefix, exempt_names

**Implementation Status:** ✅ Fully implemented

Security group rules are primarily audited for **dangerous configurations** such as:
- SSH (port 22) open to the world (0.0.0.0/0)
- RDP (port 3389) open to the world
- All ports open to the world
- Insecure protocols exposed publicly


### FloatingIp

**Resource Type:** `floating_ip`

**Allowed Actions:** log, delete, tag
**Allowed Checks:** status, age_gt, unused, exempt_names



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

### Security Group Rule Specific Checks

Security group rules support specialized checks for detecting dangerous firewall configurations:

#### Direction Check

Filter rules by traffic direction:

```yaml
check:
  direction: ingress  # or egress
```

#### Ethertype Check

Filter rules by IP version:

```yaml
check:
  ethertype: IPv4  # or IPv6
```

#### Protocol Check

Filter rules by protocol:

```yaml
check:
  protocol: tcp  # or udp, icmp, etc.
```

#### Port Check

Find rules that allow a specific port (matches if the port falls within the rule's port range):

```yaml
check:
  port: 22  # SSH port
```

#### Remote IP Prefix Check

Find rules open to specific CIDR ranges (commonly used to detect "open to world" rules):

```yaml
check:
  remote_ip_prefix: 0.0.0.0/0  # Open to the entire internet
```

#### Combined Security Group Rule Example

Find SSH rules open to the world:

```yaml
- name: critical-ssh-open-to-world
  description: Find SSH (port 22) ingress rules open to 0.0.0.0/0
  service: neutron
  resource: security_group_rule
  check:
    direction: ingress
    ethertype: IPv4
    protocol: tcp
    port: 22
    remote_ip_prefix: 0.0.0.0/0
  action: log
```

**Note:** All specified check conditions must match for a rule to be flagged. This allows precise targeting of dangerous configurations.

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


### Network Examples

#### Example 1: Find Inactive Network Resources

```yaml
- name: find-inactive-network
  description: Find inactive network resources
  service: neutron
  resource: network
  check:
    status: inactive
  action: log
```

#### Example 2: Find Old Network Resources

```yaml
- name: find-old-network
  description: Find network resources older than 30 days
  service: neutron
  resource: network
  check:
    age_gt: 30d
  action: log
```

#### Example 3: Cleanup Unused Network Resources

```yaml
- name: cleanup-unused-network
  description: Delete unused network resources
  service: neutron
  resource: network
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```

#### Example 4: Tag Old Network Resources

```yaml
- name: tag-old-network
  description: Tag network resources older than 7 days
  service: neutron
  resource: network
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-network
  action_tag_name: "Old Network"
```


### SecurityGroup Examples

#### Example 1: Find Unused Security Groups

A security group is considered unused if it's not attached to any ports:

```yaml
- name: cleanup-unused-security-groups
  description: Find security groups not attached to any ports
  service: neutron
  resource: security_group
  check:
    unused: true
    exempt_names:
      - default
  action: log
```

#### Example 2: Find Old Security Groups

```yaml
- name: find-old-security-groups
  description: Find security groups older than 90 days
  service: neutron
  resource: security_group
  check:
    age_gt: 90d
    exempt_names:
      - default
  action: log
```

#### Example 3: Tag Unused Security Groups

```yaml
- name: tag-unused-security-groups
  description: Tag security groups not attached to any ports
  service: neutron
  resource: security_group
  check:
    unused: true
    exempt_names:
      - default
  action: tag
  tag_name: ospa-unused
```

#### Example 4: Delete Unused Security Groups (Remediation)

```yaml
- name: delete-unused-security-groups
  description: Delete security groups not attached to any ports
  service: neutron
  resource: security_group
  check:
    unused: true
    exempt_names:
      - default
  action: delete
```

**Warning:** The delete action will remove matching security groups. The auditor checks if a security group is in use before deletion.


### SecurityGroupRule Examples

Security group rules are primarily audited for dangerous configurations. Here are common security-focused examples:

#### Example 1: Find SSH Open to World (Critical)

```yaml
- name: critical-ssh-open-to-world
  description: Find SSH (port 22) rules open to 0.0.0.0/0
  service: neutron
  resource: security_group_rule
  check:
    direction: ingress
    ethertype: IPv4
    protocol: tcp
    port: 22
    remote_ip_prefix: 0.0.0.0/0
  action: log
```

#### Example 2: Find RDP Open to World (Critical)

```yaml
- name: critical-rdp-open-to-world
  description: Find RDP (port 3389) rules open to 0.0.0.0/0
  service: neutron
  resource: security_group_rule
  check:
    direction: ingress
    ethertype: IPv4
    protocol: tcp
    port: 3389
    remote_ip_prefix: 0.0.0.0/0
  action: log
```

#### Example 3: Find ICMP Open to World (Warning)

```yaml
- name: warning-icmp-open-to-world
  description: Find ICMP (ping) rules open to 0.0.0.0/0
  service: neutron
  resource: security_group_rule
  check:
    direction: ingress
    ethertype: IPv4
    protocol: icmp
    remote_ip_prefix: 0.0.0.0/0
  action: log
```

#### Example 4: Find All Ingress Open to World

```yaml
- name: warning-ingress-open-to-world
  description: Find any ingress rules open to 0.0.0.0/0
  service: neutron
  resource: security_group_rule
  check:
    direction: ingress
    remote_ip_prefix: 0.0.0.0/0
  action: log
```

#### Example 5: Delete Dangerous SSH Rules (Remediation)

```yaml
- name: remediate-ssh-open-to-world
  description: Delete SSH rules open to 0.0.0.0/0
  service: neutron
  resource: security_group_rule
  check:
    direction: ingress
    protocol: tcp
    port: 22
    remote_ip_prefix: 0.0.0.0/0
  action: delete
```

**Warning:** The delete action will remove matching rules. Use with caution and always test in audit mode first.


### FloatingIp Examples

#### Example 1: Find Inactive FloatingIp Resources

```yaml
- name: find-inactive-floating_ip
  description: Find inactive floating_ip resources
  service: neutron
  resource: floating_ip
  check:
    status: inactive
  action: log
```

#### Example 2: Find Old FloatingIp Resources

```yaml
- name: find-old-floating_ip
  description: Find floating_ip resources older than 30 days
  service: neutron
  resource: floating_ip
  check:
    age_gt: 30d
  action: log
```

#### Example 3: Cleanup Unused FloatingIp Resources

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

#### Example 4: Tag Old FloatingIp Resources

```yaml
- name: tag-old-floating_ip
  description: Tag floating_ip resources older than 7 days
  service: neutron
  resource: floating_ip
  check:
    age_gt: 7d
  action: tag
  tag_name: audit-old-floating_ip
  action_tag_name: "Old FloatingIp"
```



## Complete Policy Example

Here's a complete policy file example for Neutron with security-focused rules:

```yaml
version: v1
defaults:
  workers: 50
  output: findings.jsonl
policies:
  - neutron:
    # ===========================================
    # Security Group Rule Checks (Security Focus)
    # ===========================================
    
    # Critical: SSH open to the world
    - name: critical-ssh-open-to-world
      description: Find SSH (port 22) rules open to 0.0.0.0/0
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        ethertype: IPv4
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: log
    
    # Critical: RDP open to the world
    - name: critical-rdp-open-to-world
      description: Find RDP (port 3389) rules open to 0.0.0.0/0
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        ethertype: IPv4
        protocol: tcp
        port: 3389
        remote_ip_prefix: 0.0.0.0/0
      action: log
    
    # Warning: ICMP open to the world
    - name: warning-icmp-open-to-world
      description: Find ICMP rules open to 0.0.0.0/0
      service: neutron
      resource: security_group_rule
      check:
        direction: ingress
        ethertype: IPv4
        protocol: icmp
        remote_ip_prefix: 0.0.0.0/0
      action: log
    
    # ===========================================
    # Security Group Checks (Cleanup Focus)
    # ===========================================
    
    # Find unused security groups
    - name: cleanup-unused-security-groups
      description: Find security groups not attached to any ports
      service: neutron
      resource: security_group
      check:
        unused: true
        exempt_names:
          - default
      action: log
    
    # ===========================================
    # Network Checks (Cleanup Focus)
    # ===========================================
    
    # Find unused networks
    - name: cleanup-unused-networks
      description: Find networks with no subnets
      service: neutron
      resource: network
      check:
        unused: true
        exempt_names:
          - external
          - public
      action: log
    
    # Find old networks
    - name: cleanup-old-networks
      description: Find networks older than 90 days
      service: neutron
      resource: network
      check:
        age_gt: 90d
        exempt_names:
          - external
          - public
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
- Verify resource type is supported: `{network Networks [status age_gt unused exempt_names] [log delete tag] {false false false}}`, `{security_group Security groups [status age_gt unused exempt_names] [log delete tag] {false false false}}`, `{security_group_rule Security group rules [status age_gt unused exempt_names] [log delete tag] {false false false}}`, `{floating_ip Floating IP addresses [status age_gt unused exempt_names] [log delete tag] {false false false}}`
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
