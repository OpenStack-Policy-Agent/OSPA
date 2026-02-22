# OpenStack Security Guide Compliance Matrix

This document maps OSPA's coverage to the [OpenStack Security Guide](https://docs.openstack.org/security-guide/) per-service checklists. It provides a compliance matrix showing which checklist items OSPA can help verify and how.

## Summary

| Service | Total Items | Config-level (manual) | API-driven | Documented |
|---------|-------------|----------------------|------------|------------|
| Identity (Keystone) | 7 | 7 | 0 | 7 |
| Networking (Neutron) | 5 | 5 | 0 | 5 |
| Compute (Nova) | 5 | 5 | 0 | 5 |
| Block Storage (Cinder) | 9 | 8 | 1 | 9 |
| Image (Glance) | 5 | 5 | 0 | 5 |
| Key Manager (Barbican) | 4 | 4 | 0 | 4 |
| Shared File Systems (Manila) | 8 | 8 | 0 | 8 |
| **Total** | **43** | **42** | **1** | **43** |

## Coverage Types

### Config-level (manual)

These checklist items require inspection of configuration files on the host system (ownership, permissions, TLS settings, auth configuration, etc.). OSPA does not have direct access to configuration files. These items are **documented** in the policy guide for manual verification by operators during deployment audits or compliance reviews.

### API-driven

These checklist items can be verified automatically by OSPA via the OpenStack API. OSPA can enforce these checks in policy rules and report violations (or remediate them where actions are defined).

---

## Identity (Keystone) — 7 items

| ID | Description | Coverage Type | OSPA Status |
|----|-------------|---------------|-------------|
| Check-Identity-01 | User/group ownership of config files set to keystone | Config-level (manual) | Documented |
| Check-Identity-02 | Strict permissions (640) on configuration files | Config-level (manual) | Documented |
| Check-Identity-03 | TLS enabled for Identity | Config-level (manual) | Documented |
| Check-Identity-05 | max_request_body_size set to default (114688) | Config-level (manual) | Documented |
| Check-Identity-06 | Admin token disabled | Config-level (manual) | Documented |
| Check-Identity-07 | insecure_debug set to false | Config-level (manual) | Documented |
| Check-Identity-08 | Fernet token provider used | Config-level (manual) | Documented |

---

## Networking (Neutron) — 5 items

| ID | Description | Coverage Type | OSPA Status |
|----|-------------|---------------|-------------|
| Check-Neutron-01 | User/group ownership of config files set to root/neutron | Config-level (manual) | Documented |
| Check-Neutron-02 | Strict permissions (640) on configuration files | Config-level (manual) | Documented |
| Check-Neutron-03 | Keystone used for authentication | Config-level (manual) | Documented |
| Check-Neutron-04 | Secure protocol (TLS) for authentication | Config-level (manual) | Documented |
| Check-Neutron-05 | TLS enabled on Neutron API server | Config-level (manual) | Documented |

---

## Compute (Nova) — 5 items

| ID | Description | Coverage Type | OSPA Status |
|----|-------------|---------------|-------------|
| Check-Compute-01 | User/group ownership of config files set to root/nova | Config-level (manual) | Documented |
| Check-Compute-02 | Strict permissions (640) on configuration files | Config-level (manual) | Documented |
| Check-Compute-03 | Keystone used for authentication | Config-level (manual) | Documented |
| Check-Compute-04 | Secure protocol (TLS) for authentication | Config-level (manual) | Documented |
| Check-Compute-05 | Nova communicates with Glance over TLS | Config-level (manual) | Documented |

---

## Block Storage (Cinder) — 9 items

| ID | Description | Coverage Type | OSPA Status |
|----|-------------|---------------|-------------|
| Check-Block-01 | User/group ownership of config files set to root/cinder | Config-level (manual) | Documented |
| Check-Block-02 | Strict permissions (640) on configuration files | Config-level (manual) | Documented |
| Check-Block-03 | Keystone used for authentication | Config-level (manual) | Documented |
| Check-Block-04 | TLS enabled for authentication | Config-level (manual) | Documented |
| Check-Block-05 | Cinder communicates with Nova over TLS | Config-level (manual) | Documented |
| Check-Block-06 | Cinder communicates with Glance over TLS | Config-level (manual) | Documented |
| Check-Block-07 | NAS operating in a secure environment | Config-level (manual) | Documented |
| Check-Block-08 | Max request body size set to default (114688) | Config-level (manual) | Documented |
| Check-Block-09 | Volume encryption feature enabled | API-driven | Enforced (encrypted check) |

---

## Image (Glance) — 5 items

| ID | Description | Coverage Type | OSPA Status |
|----|-------------|---------------|-------------|
| Check-Image-01 | User/group ownership of config files set to root/glance | Config-level (manual) | Documented |
| Check-Image-02 | Strict permissions (640) on configuration files | Config-level (manual) | Documented |
| Check-Image-03 | Keystone used for authentication | Config-level (manual) | Documented |
| Check-Image-04 | TLS enabled for authentication | Config-level (manual) | Documented |
| Check-Image-05 | Masked port scans prevented (copy_from restricted) | Config-level (manual) | Documented |

---

## Key Manager (Barbican) — 4 items

| ID | Description | Coverage Type | OSPA Status |
|----|-------------|---------------|-------------|
| Check-Key-Manager-01 | Ownership of config files set to root/barbican | Config-level (manual) | Documented |
| Check-Key-Manager-02 | Strict permissions (640) on configuration files | Config-level (manual) | Documented |
| Check-Key-Manager-03 | OpenStack Identity used for authentication | Config-level (manual) | Documented |
| Check-Key-Manager-04 | TLS enabled for authentication | Config-level (manual) | Documented |

---

## Shared File Systems (Manila) — 8 items

| ID | Description | Coverage Type | OSPA Status |
|----|-------------|---------------|-------------|
| Check-Shared-01 | User/group ownership of config files set to root/manila | Config-level (manual) | Documented |
| Check-Shared-02 | Strict permissions (640) on configuration files | Config-level (manual) | Documented |
| Check-Shared-03 | OpenStack Identity used for authentication | Config-level (manual) | Documented |
| Check-Shared-04 | TLS enabled for authentication | Config-level (manual) | Documented |
| Check-Shared-05 | Manila communicates with Compute over TLS | Config-level (manual) | Documented |
| Check-Shared-06 | Manila communicates with Networking over TLS | Config-level (manual) | Documented |
| Check-Shared-07 | Manila communicates with Block Storage over TLS | Config-level (manual) | Documented |
| Check-Shared-08 | Max request body size set to default (114688) | Config-level (manual) | Documented |

---

## Additional API-driven Security Checks

Beyond the 43 official OpenStack Security Guide checklist items, OSPA provides additional API-driven security checks that complement the guide. These help operators identify common misconfigurations and risks discoverable via the OpenStack API:

| Area | Check | Description |
|------|-------|-------------|
| **Networking** | Open security group rules | Rules allowing ingress from 0.0.0.0/0 (all IPs) or overly wide port ranges |
| **Networking** | Unassociated floating IPs | Floating IPs not attached to any port—potential waste and attack surface |
| **Networking** | Shared networks | Networks shared across tenants—multi-tenancy concern |
| **Networking** | Ports without security groups | Ports with no security group assignments |
| **Compute** | Instances without keypairs | Instances lacking SSH keypair—impacts access control |
| **Compute** | Deprecated image names | Instances using deprecated or EOL image families |
| **Image** | Public images | Images with public visibility—data exposure risk |
| **Block Storage** | Unencrypted volumes/snapshots | Volumes or snapshots without encryption enabled |
| **Block Storage** | Unattached volumes | Volumes not attached to any instance |
| **Block Storage** | Volumes without backups | Volumes lacking backup protection |
| **Identity** | Password expired | Users with expired passwords |
| **Identity** | MFA not enabled | Users without multi-factor authentication |
| **Identity** | Inactive users | Users inactive for extended periods |
| **Identity** | Admin role assignment | Users with admin role—privilege audit |

These additional checks can be configured via OSPA policy rules and linked to Security Guide items using the `guide_ref` field where applicable.

---

## Reference

- [OpenStack Security Guide](https://docs.openstack.org/security-guide/)
