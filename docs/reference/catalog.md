# Resource Catalog

This page provides a comprehensive list of all OpenStack resources that OSPA can audit.

## Implementation Status Legend

| Status | Meaning |
|--------|---------|
| ✔ | **Implemented** - Full audit and remediation support |
| ◐ | **Placeholder** - Scaffolding exists, implementation in progress |
| — | **Not generated** - Known resource, no code yet |

---

## Core Services

### Neutron (Networking)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `network` | ✔ | status, age_gt, unused, exempt_names | log, delete, tag |
| `security_group` | ✔ | status, age_gt, unused, exempt_names | log, delete, tag |
| `security_group_rule` | ✔ | direction, ethertype, protocol, port, remote_ip_prefix, port_range_wide, exempt_names | log, delete |
| `floating_ip` | ◐ | status, age_gt, unused, exempt_names | log, delete, tag |
| `subnet` | ✔ | status, age_gt, unused, exempt_names | log, delete, tag |
| `port` | ◐ | status, age_gt, unused, exempt_names, no_security_group | log, delete, tag |
| `router` | ◐ | status, age_gt, unused, exempt_names | log, delete, tag |
| `loadbalancer` | — | — | — |
| `pool` | — | — | — |
| `member` | — | — | — |

### Nova (Compute)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `instance` | ◐ | status, age_gt, exempt_names | log, delete |
| `keypair` | ◐ | age_gt, unused, exempt_names | log, delete |
| `server` | — | — | — |
| `flavor` | — | — | — |
| `hypervisor` | — | — | — |

### Cinder (Block Storage)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `volume` | ◐ | status, age_gt, unused, exempt_names | log, delete, tag |
| `snapshot` | ◐ | status, age_gt, unused, exempt_names | log, delete, tag |
| `backup` | — | — | — |
| `qos` | — | — | — |

---

## Additional Services

### Glance (Image)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `image` | — | — | — |
| `member` | — | — | — |

### Keystone (Identity)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `user` | — | — | — |
| `role` | — | — | — |
| `project` | — | — | — |
| `domain` | — | — | — |
| `group` | — | — | — |
| `service` | — | — | — |

### Heat (Orchestration)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `stack` | — | — | — |
| `resource` | — | — | — |
| `template` | — | — | — |
| `snapshot` | — | — | — |

### Swift (Object Storage)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `container` | — | — | — |
| `object` | — | — | — |
| `account` | — | — | — |

### Octavia (Load Balancing)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `loadbalancer` | — | — | — |
| `listener` | — | — | — |
| `pool` | — | — | — |
| `member` | — | — | — |
| `healthmonitor` | — | — | — |

### Barbican (Key Manager)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `secret` | — | — | — |
| `container` | — | — | — |
| `order` | — | — | — |

### Manila (Shared File Systems)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `share` | — | — | — |
| `share_snapshot` | — | — | — |
| `share_network` | — | — | — |
| `share_server` | — | — | — |

### Trove (Database)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `instance` | — | — | — |
| `cluster` | — | — | — |
| `backup` | — | — | — |
| `datastore` | — | — | — |

### Magnum (Container Infrastructure)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `cluster` | — | — | — |
| `cluster_template` | — | — | — |
| `bay` | — | — | — |
| `baymodel` | — | — | — |

### Ironic (Bare Metal)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `node` | — | — | — |
| `port` | — | — | — |
| `driver` | — | — | — |
| `chassis` | — | — | — |

### Designate (DNS)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `zone` | — | — | — |
| `recordset` | — | — | — |
| `record` | — | — | — |

### Senlin (Clustering)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `cluster` | — | — | — |
| `profile` | — | — | — |
| `node` | — | — | — |
| `policy` | — | — | — |

### Zaqar (Messaging)

| Resource | Status | Checks | Actions |
|----------|--------|--------|---------|
| `queue` | — | — | — |
| `message` | — | — | — |
| `subscription` | — | — | — |

---

## Adding New Resources

To add support for a new resource:

1. **Use the scaffold tool** for quick generation:

   ```bash
   go run ./cmd/scaffold --service <service> --resources <resource1,resource2>
   ```

2. **Implement the generated stubs** - see [Adding Resources](../developer-guide/adding-resources.md)

3. **Submit a PR** - contributions welcome!

## Check Types Reference

| Check | Description | Example |
|-------|-------------|---------|
| `status` | Match resource status | `status: ERROR` |
| `age_gt` | Resource older than duration | `age_gt: 30d` |
| `unused` | Resource not in use | `unused: true` |
| `exempt_names` | Skip matching names | `exempt_names: ["system-*"]` |
| `direction` | Rule direction (security_group_rule) | `direction: ingress` |
| `ethertype` | Ethernet type (security_group_rule) | `ethertype: IPv4` |
| `protocol` | Network protocol | `protocol: tcp` |
| `port` | Port number | `port: 22` |
| `remote_ip_prefix` | CIDR range | `remote_ip_prefix: "0.0.0.0/0"` |
| `port_range_wide` | Port range >100 ports | `port_range_wide: true` |
| `unassociated` | Floating IP not attached | `unassociated: true` |
| `shared_network` | Network shared across tenants | `shared_network: true` |
| `no_security_group` | Port has no security groups | `no_security_group: true` |
| `image_name` | Deprecated image names (instance) | `image_name: ["ubuntu-14*"]` |
| `no_keypair` | No SSH keypair (instance) | `no_keypair: true` |
| `encrypted` | Volume/snapshot not encrypted | `encrypted: false` |
| `attached` | Volume not attached | `attached: false` |
| `has_backup` | Volume has no backup | `has_backup: false` |
| `visibility` | Image visibility (Glance) | `visibility: public` |
| `password_expired` | User password expired (Keystone) | `password_expired: true` |
| `mfa_enabled` | MFA not enabled (Keystone) | `mfa_enabled: false` |
| `inactive_days` | User inactive N days (Keystone) | `inactive_days: 90` |
| `has_admin_role` | User has admin role (Keystone) | `has_admin_role: true` |
| `token_provider` | Token provider type (Keystone) | `token_provider: fernet` |

## Rule Classification Fields

These fields classify rules at the policy level and are not check conditions:

| Field | Description | Example |
|-------|-------------|---------|
| `severity` | Finding severity classification | `severity: high` |
| `category` | Finding category classification | `category: security` |
| `guide_ref` | Security Guide reference | `guide_ref: "Check-Block-09"` |

## Action Types Reference

| Action | Description | Destructive |
|--------|-------------|-------------|
| `log` | Report violation only | No |
| `tag` | Add tag to resource | No |
| `delete` | Delete the resource | **Yes** |

