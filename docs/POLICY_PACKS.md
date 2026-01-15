# Policy Packs

OSPA policy packs are reusable YAML policies designed to jump-start audits.

## Available Packs

- `examples/policies/baseline-security.yaml`
  - Baseline checks for common OpenStack resources.
  - Safe by default (log-only actions).

## Running A Pack

```bash
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/policies/baseline-security.yaml --out findings.jsonl
```

## Customizing Packs

- Adjust `defaults.workers` for scale.
- Change `action` from `log` to `delete` or `tag` only with `--fix`.
- Use `--allow-actions` to restrict remediation actions.

