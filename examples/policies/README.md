# Policy Packs

This directory contains reusable policy packs and service-specific guides.

## Packs

- `baseline-security.yaml` - Baseline security checks including:
  - **SSH open to world** - Detects SSH (port 22) rules with `0.0.0.0/0`
  - **RDP open to world** - Detects RDP (port 3389) rules with `0.0.0.0/0`
  - **Unused security groups** - Finds security groups not attached to ports
  - **Old snapshots** - Finds Cinder snapshots older than 30 days

## Service Guides

- `nova-policy-guide.md` - Nova (Compute) resource policies
- `neutron-policy-guide.md` - Neutron (Network) resource policies including:
  - Security group rule checks (direction, protocol, port, remote_ip_prefix)
  - Security group unused detection
  - Network and floating IP auditing
- `cinder-policy-guide.md` - Cinder (Block Storage) resource policies

## Usage

### Audit Mode (Safe - No Changes)

```bash
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/policies/baseline-security.yaml --out findings.jsonl
```

### Remediation Mode (Makes Changes)

```bash
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/policies/baseline-security.yaml --out findings.jsonl --fix
```

## Security Group Rule Checks

The baseline security pack includes checks for dangerous security group rules:

| Check | Description | Severity |
|-------|-------------|----------|
| SSH open to world | Port 22 from 0.0.0.0/0 | Critical |
| RDP open to world | Port 3389 from 0.0.0.0/0 | Critical |
| All TCP open to world | Any TCP from 0.0.0.0/0 | Warning |

These checks match rules based on:
- `direction` - ingress or egress
- `ethertype` - IPv4 or IPv6
- `protocol` - tcp, udp, icmp, etc.
- `port` - Specific port number (matches port ranges)
- `remote_ip_prefix` - CIDR like 0.0.0.0/0

