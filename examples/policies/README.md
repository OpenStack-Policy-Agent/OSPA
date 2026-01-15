# Policy Packs

This directory contains reusable policy packs and service-specific guides.

## Packs

- `baseline-security.yaml` - Baseline checks for common OpenStack resources.

## Service Guides

- `nova-policy-guide.md`
- `neutron-policy-guide.md`
- `cinder-policy-guide.md`

## Usage

```bash
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/policies/baseline-security.yaml --out findings.jsonl
```

