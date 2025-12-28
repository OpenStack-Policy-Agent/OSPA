## OSPA (OpenStack Policy Agent) — MVP

OSPA is a lightweight OpenStack policy/compliance agent:

- **Define**: rules are written in simple YAML.
- **Detect**: scans OpenStack resources in parallel (worker pool).
- **Correct**: optionally remediates violations (**dry-run by default**).

### What the MVP does today

- Scans **Nova servers** (`compute.server`)
- Detects servers in a given status (e.g. `SHUTOFF`) whose `Updated` timestamp is older than a threshold
- Emits **JSONL findings**
- Optionally remediates by **deleting** those stale servers (only when explicitly enabled)

### Usage

- **Audit-only (default, safe)**:

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml go run ./cmd/agent --cloud mycloud --policy ./examples/rules.yaml --out findings.jsonl
```

- **Enforce (actually apply remediations)**:

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml go run ./cmd/agent --cloud mycloud --policy ./examples/rules.yaml --out findings.jsonl --apply
```

### Notes

- Use `--all-tenants` only if you have admin credentials and want to scan the whole cloud.
- `mode: enforce` in the policy enables remediation logic, but **nothing is changed unless `--apply` is set**.

### Tests

- Developer docs: see `DEVELOPMENT.md`.

- **Unit tests**:

```bash
go test ./...
```

- **OpenStack e2e smoke tests** (connect via `clouds.yaml` using `OS_CLOUD`):

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml OS_CLOUD=mycloud go test -tags=e2e ./e2e/...
```


