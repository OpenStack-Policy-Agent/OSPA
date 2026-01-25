# Running the Agent

This guide covers how to run OSPA in different modes and configurations.

## Environment Variables
Those variables should be available in the system in order for the OSPA command to work.

| Variable | Description |
|----------|-------------|
| `OS_CLIENT_CONFIG_FILE` | Path to clouds.yaml |
| `OS_CLOUD` | Default cloud name |

```bash
export OS_CLIENT_CONFIG_FILE=/path/to/clouds.yaml
export OS_CLOUD=mycloud # optional (could be set later)

go run ./cmd/agent --policy policy.yaml --out findings.json
```

## Basic Usage

```bash
go run ./cmd/agent \
  --cloud mycloud \
  --policy policy.yaml \
  --out findings.json
```

Or with a built binary:

```bash
./ospa --cloud mycloud --policy policy.yaml --out findings.json
```

## Modes of Operation

### Audit Mode (Default)

Safe, read-only scanning:

```bash
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy policy.yaml \
  --out findings.json
```

- Discovers resources
- Evaluates against policies
- Reports violations
- **Makes no changes**

### Remediation Mode

Applies fixes to violations:

```bash
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy policy.yaml \
  --out findings.json \
  --fix
```

!!! warning
    Remediation mode can modify or delete resources. Always test policies in audit mode first.

## CLI Reference

### Required Flags

| Flag | Description |
|------|-------------|
| `--cloud` | Cloud name from clouds.yaml (or use `$OS_CLOUD`) |
| `--policy` | Path to policy YAML file |

### Output Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--out` | stdout | Output file path |
| `--out-format` | json | Output format (json, csv) |

### Execution Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--workers` | 16 | Number of concurrent workers |
| `--all-tenants` | false | Audit all projects (admin only) |
| `--fix` | false | Enable remediation actions |
| `--allow-actions` | all | Comma-separated list of allowed actions |

### Logging Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--log-format` | text | Log format (text, json) |
| `--log-level` | info | Log level (debug, info, warn, error) |

### Metrics Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--metrics-addr` | disabled | Prometheus metrics address (e.g., `:9090`) |

## Examples

### Standard Audit

```bash
go run ./cmd/agent \
  --cloud mycloud \
  --policy ./examples/policies.yaml \
  --out findings.json
```

### All Tenants (Admin)

```bash
go run ./cmd/agent \
  --cloud admin-cloud \
  --policy policy.yaml \
  --out findings.json \
  --all-tenants
```

### With Remediation

```bash
go run ./cmd/agent \
  --cloud mycloud \
  --policy policy.yaml \
  --out findings.json \
  --fix
```

### Restricted Actions

```bash
# Only allow log and tag, block delete
go run ./cmd/agent \
  --cloud mycloud \
  --policy policy.yaml \
  --out findings.json \
  --fix \
  --allow-actions log,tag
```

### CSV Output

```bash
go run ./cmd/agent \
  --cloud mycloud \
  --policy policy.yaml \
  --out findings.csv \
  --out-format csv
```

### High Concurrency

```bash
go run ./cmd/agent \
  --cloud mycloud \
  --policy policy.yaml \
  --out findings.json \
  --workers 100
```
### Debug Logging

```bash
go run ./cmd/agent \
  --cloud mycloud \
  --policy policy.yaml \
  --out findings.json \
  --log-level debug \
  --log-format json
```

## Performance Tuning

### Workers

The `--workers` flag controls concurrency:

```bash
# Small cloud
--workers 10

# Medium cloud
--workers 50

# Large cloud with many resources
--workers 100
```

!!! tip
    Start with fewer workers and increase if audit takes too long. Too many workers may hit API rate limits.

### Memory

For large environments, ensure sufficient memory:

```bash
# Increase Go's memory limit if needed
GOMEMLIMIT=4GiB go run ./cmd/agent ...
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (even if violations found) |
| 1 | Error (policy load failed, auth failed, etc.) |

## Automation

### Cron Job

```bash
# /etc/cron.d/ospa-audit
0 2 * * * root /opt/ospa/ospa --cloud prod --policy /etc/ospa/policy.yaml --out /var/log/ospa/findings-$(date +\%Y\%m\%d).json
```

### CI/CD Pipeline

```yaml
# .github/workflows/audit.yml
name: Security Audit
on:
  schedule:
    - cron: '0 6 * * *'
jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: |
          go run ./cmd/agent \
            --cloud ${{ secrets.OS_CLOUD }} \
            --policy policy.yaml \
            --out findings.json
      - uses: actions/upload-artifact@v4
        with:
          name: audit-findings
          path: findings.json
```

### Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: ospa-audit
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: ospa
            image: ospa:latest
            args:
            - --cloud=mycloud
            - --policy=/config/policy.yaml
            - --out=/output/findings.json
            volumeMounts:
            - name: clouds
              mountPath: /etc/openstack
            - name: policy
              mountPath: /config
            - name: output
              mountPath: /output
          volumes:
          - name: clouds
            secret:
              secretName: openstack-clouds
          - name: policy
            configMap:
              name: ospa-policy
          - name: output
            persistentVolumeClaim:
              claimName: ospa-output
          restartPolicy: OnFailure
```

