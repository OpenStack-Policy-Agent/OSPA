## Developer Guide (OSPA)

This doc is for contributors working on OSPA locally.

### Repo layout

- `cmd/agent`: CLI entrypoint
- `pkg/auth`: OpenStack auth/session (loads from `clouds.yaml`)
- `pkg/discovery`: resource discovery (currently: Nova servers)
- `pkg/engine`: rule evaluation + worker pool + remediation execution
- `pkg/policy`: YAML policy parsing/validation
- `pkg/report`: JSONL findings output + summary
- `e2e`: OpenStack end-to-end tests (`-tags=e2e`)
- `examples`: example policy files

### Prerequisites

- Go installed (matches `go.mod`)
- Access to an OpenStack cloud and a valid `clouds.yaml`, save it to export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml
- `OS_CLOUD` set to a cloud name present in `clouds.yaml`

### Running the agent locally

- **Audit (safe)**:

```bash
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/rules.yaml --out findings.jsonl
```

- **Enforce (will remediate when policy mode=enforce)**:

```bash
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/rules.yaml --out findings.jsonl --apply
```

### Policy schema (current MVP)

Top-level:
- `version`: required (e.g. `v1`)
- `defaults.workers`: optional
- `defaults.output`: optional JSONL output path

Rule (example of `compute.server`):
- `id`: required unique string
- `resource`: must be `compute.server`
- `mode`: `audit` or `enforce`
- `filters.status`: required (e.g. `SHUTOFF`)
- `conditions.updatedOlderThanDays`: required `>0`
- `remediation`: preferred remediation action (for now supportting: `delete`)

### Tests

#### Unit tests

```bash
go test ./...
```

#### OpenStack e2e tests (real cloud)

These tests **create and delete Nova servers**. Run them only in a safe test project.

Required environment:
- `OS_CLOUD`: cloud name in `clouds.yaml`
- `OSPA_E2E_IMAGE_ID`: image ID to boot a test server
- `OSPA_E2E_FLAVOR_ID`: flavor ID
- `OSPA_E2E_NETWORK_ID`: network UUID

Optional environment:
- `OSPA_E2E_ALL_TENANTS=true` (used by the smoke list test)

Run:

```bash
OS_CLOUD=mycloud \
OSPA_E2E_IMAGE_ID=... OSPA_E2E_FLAVOR_ID=... OSPA_E2E_NETWORK_ID=... \
go test -tags=e2e ./...
```

What e2e covers today:
- Auth via `clouds.yaml` and a minimal Nova list call
- **Audit path**: creates compliant + noncompliant servers; detects violations but does not delete
- **Enforce+apply path**: creates a noncompliant server and verifies remediation deletes it

### Common troubleshooting

- **403/permission errors**: ensure the `OS_CLOUD` points at a project with permission to create/stop/delete servers.
- **Quota errors**: the e2e suite creates multiple servers; ensure quota is available.
- **Slow state transitions**: tests wait up to ~10 minutes for `ACTIVE`/`SHUTOFF` and deletion.


