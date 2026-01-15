## OSPA (OpenStack Policy Agent)

OSPA is a policy-driven audit + remediation agent for OpenStack clouds.

- **Define**: write policies/rules in YAML.
- **Discover**: concurrently enumerates OpenStack resources per service/resource.
- **Audit**: evaluates each discovered resource against rules.
- **Remediate**: optionally applies actions (log/delete/tag). **Safe by default**.
- **Extend**: add new services/resources via a scaffolding CLI (with OpenStack service/resource validation).

### Architecture (high level)

- **Service plugins**: `pkg/services/services/<service>.go`
- **Discovery**: `pkg/discovery/services/<service>.go`
- **Audit**: `pkg/audit/<service>/<resource>.go`
- **Orchestrator/worker pool**: `pkg/orchestrator/`
- **Policy validation**: `pkg/policy/validation/<service>.go`
- **Scaffolding tool**: `cmd/scaffold/`
- **E2E engine**: `e2e/`

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) and [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md).
Policy packs: see [`docs/POLICY_PACKS.md`](docs/POLICY_PACKS.md).

### Supported OpenStack resources

Legend:
- **✔**: implemented end-to-end (real OpenStack API calls + audit/remediation logic)
- **◐**: placeholder (generated stubs compile, but OpenStack-specific logic is still TODO)
- **—**: not generated yet (known to the scaffold registry, but no code exists in `pkg/` yet)


| **service** | **resource** | **main** |
|:----------:|:------------:|:--------:|
| nova | instance | ◐ |
| nova | keypair | ◐ |
| nova | server | — |
| nova | flavor | — |
| nova | hypervisor | — |
| neutron | security_group | ◐ |
| neutron | security_group_rule | ◐ |
| neutron | floating_ip | ◐ |
| neutron | network | — |
| neutron | subnet | — |
| neutron | port | — |
| neutron | router | — |
| neutron | loadbalancer | — |
| neutron | pool | — |
| neutron | member | — |
| cinder | volume | ◐ |
| cinder | snapshot | ◐ |
| cinder | backup | — |
| cinder | qos | — |
| glance | image | — |
| glance | member | — |
| keystone | user | — |
| keystone | role | — |
| keystone | project | — |
| keystone | domain | — |
| keystone | group | — |
| keystone | service | — |
| heat | stack | — |
| heat | resource | — |
| heat | template | — |
| heat | snapshot | — |
| swift | container | — |
| swift | object | — |
| swift | account | — |
| trove | instance | — |
| trove | cluster | — |
| trove | backup | — |
| trove | datastore | — |
| magnum | cluster | — |
| magnum | cluster_template | — |
| magnum | bay | — |
| magnum | baymodel | — |
| barbican | secret | — |
| barbican | container | — |
| barbican | order | — |
| manila | share | — |
| manila | share_snapshot | — |
| manila | share_network | — |
| manila | share_server | — |
| ironic | node | — |
| ironic | port | — |
| ironic | driver | — |
| ironic | chassis | — |
| designate | zone | — |
| designate | recordset | — |
| designate | record | — |
| octavia | loadbalancer | — |
| octavia | listener | — |
| octavia | pool | — |
| octavia | member | — |
| octavia | healthmonitor | — |
| senlin | cluster | — |
| senlin | profile | — |
| senlin | node | — |
| senlin | policy | — |
| zaqar | queue | — |
| zaqar | message | — |
| zaqar | subscription | — |

### Usage

- **Audit-only (default, safe)**:

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml
go run ./cmd/agent --cloud mycloud --policy ./examples/policies.yaml --out findings.jsonl
```

- **Apply remediation**:

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml
go run ./cmd/agent --cloud mycloud --policy ./examples/policies.yaml --out findings.jsonl --fix
```

Notes:
- Use `--all-tenants` only with admin credentials.
- Policies may set enforce mode, but **nothing changes unless `--fix` is set**.
- Use `--allow-actions` to restrict remediation actions.

### Output formats

- **JSONL (default)**:
  - `--out findings.jsonl --out-format jsonl`
- **CSV**:
  - `--out findings.csv --out-format csv`

### Metrics

Expose Prometheus metrics:

```bash
go run ./cmd/agent --cloud mycloud --policy ./examples/policies.yaml --metrics-addr :9090
```

### Extending OSPA (scaffold)

- **List available OpenStack services/resources**:

```bash
go run ./cmd/scaffold --list
```

- **Generate a service/resource skeleton**:

```bash
go run ./cmd/scaffold --service <name> --resources <r1,r2> --type <serviceType>
```

See [`docs/scaffold-README.md`](docs/scaffold-README.md).

### Tests

- **All unit tests**:

```bash
go test ./... -count=1
```

- **pkg-only (fast CI target)**:

```bash
go test ./pkg/... -count=1
```

- **pkg coverage helper script**:

```bash
bash ./scripts/test-pkg.sh
```

- **Scaffold tool tests**:

```bash
go test ./cmd/scaffold/... -count=1
```

- **OpenStack e2e smoke tests** (requires real cloud + `OS_CLOUD`):

```bash
export OS_CLIENT_CONFIG_FILE=path/to/clouds.yaml OS_CLOUD=mycloud
go test -tags=e2e ./e2e/... -count=1
```


