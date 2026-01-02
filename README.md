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
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml go run ./cmd/agent --cloud mycloud --policy ./examples/policies.yaml --out findings.jsonl
```

- **Enforce (actually apply remediations)**:

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml go run ./cmd/agent --cloud mycloud --policy ./examples/policies.yaml --out findings.jsonl --apply
```

### Notes

- Use `--all-tenants` only if you have admin credentials and want to scan the whole cloud.
- `mode: enforce` in the policy enables remediation logic, but **nothing is changed unless `--apply` is set**.

### Extending OSPA

To add support for new OpenStack services or resource types:

- **List Available Services**: Use `go run ./cmd/scaffold --list` to see all supported OpenStack services and resources
- **Scaffold Tool**: Use `go run ./cmd/scaffold --service <name> --resources <list>` to generate boilerplate code (validates against known OpenStack services)
- **Developer Guide**: See [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) for complete development workflow
- **Architecture**: See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for detailed architecture explanation
- **Scaffold Docs**: See [`docs/scaffold-README.md`](docs/scaffold-README.md) for scaffold tool usage

### Tests

- Developer docs: see [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md).

- **Unit tests**:

```bash
go test ./...
```

- **OpenStack e2e smoke tests** (connect via `clouds.yaml` using `OS_CLOUD`):

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml OS_CLOUD=mycloud go test -tags=e2e ./e2e/...
```


