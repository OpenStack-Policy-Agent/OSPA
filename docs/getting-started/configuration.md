# Configuration

This guide covers how to configure OSPA to connect to your OpenStack cloud.

## OpenStack Credentials

OSPA uses the standard OpenStack client configuration via `clouds.yaml`.

### Environment Variables

Set these environment variables before running OSPA:

| Variable | Required | Description |
|----------|----------|-------------|
| `OS_CLIENT_CONFIG_FILE` | Yes | Path to your `clouds.yaml` file |
| `OS_CLOUD` | Yes | Name of the cloud configuration to use |

```bash
export OS_CLIENT_CONFIG_FILE=/path/to/clouds.yaml
export OS_CLOUD=mycloud
```

### clouds.yaml Format

Create a `clouds.yaml` file with your OpenStack credentials:

```yaml
clouds:
  mycloud:
    auth:
      auth_url: https://openstack.example.com:5000/v3
      username: myuser
      password: mypassword
      project_name: myproject
      user_domain_name: Default
      project_domain_name: Default
    region_name: RegionOne
    interface: public
```

#### Using Application Credentials

For better security, use application credentials:

```yaml
clouds:
  mycloud:
    auth:
      auth_url: https://openstack.example.com:5000/v3
      application_credential_id: <app-cred-id>
      application_credential_secret: <app-cred-secret>
    region_name: RegionOne
    auth_type: v3applicationcredential
```

### Default Locations

OSPA searches for `clouds.yaml` in these locations (in order):

1. Path specified by `OS_CLIENT_CONFIG_FILE`
2. `./clouds.yaml` (current directory)
3. `~/.config/openstack/clouds.yaml`
4. `/etc/openstack/clouds.yaml`

## Policy Configuration

### Policy File

Policies are YAML files that define audit rules. See [Writing Policies](../user-guide/policies.md) for details.

```yaml
version: v1
defaults:
  workers: 50
  output: findings.json
policies:
  - neutron:
    - name: my-rule
      # ... rule definition
```

### Defaults Section

Configure default behavior for all rules:

| Option | Default | Description |
|--------|---------|-------------|
| `workers` | 16 | Number of concurrent workers |
| `output` | stdout | Default output file path |

## CLI Configuration

All configuration can also be set via CLI flags:

```bash
go run ./cmd/agent \
  --cloud mycloud \
  --policy ./policy.yaml \
  --out findings.json \
  --workers 20 \
  --all-tenants \
  --fix
```

### CLI Flags

| Flag | Description |
|------|-------------|
| `--cloud` | Cloud name from clouds.yaml |
| `--policy` | Path to policy YAML file |
| `--out` | Output file path |
| `--out-format` | Output format (json, csv) |
| `--workers` | Number of workers |
| `--all-tenants` | Audit all tenants (admin only) |
| `--fix` | Enable remediation actions |
| `--metrics-addr` | Prometheus metrics address |
| `--log-format` | Log format (text, json) |
| `--log-level` | Log level (debug, info, warn, error) |

## Multi-Cloud Configuration

You can configure multiple clouds in `clouds.yaml`:

```yaml
clouds:
  production:
    auth:
      auth_url: https://prod.example.com:5000/v3
      # ... credentials
  staging:
    auth:
      auth_url: https://staging.example.com:5000/v3
      # ... credentials
```

Then select which cloud to audit:

```bash
# Audit production
OS_CLOUD=production go run ./cmd/agent --policy policy.yaml

# Audit staging
OS_CLOUD=staging go run ./cmd/agent --policy policy.yaml
```

## Verification

Verify your configuration is working:

```bash
# Test authentication
openstack token issue

# Test with OSPA (dry run)
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy examples/policies.yaml \
  --out /dev/null
```

## Troubleshooting

### Authentication Errors

```
Error: failed to authenticate: 401 Unauthorized
```

- Verify credentials in `clouds.yaml`
- Check `auth_url` is correct and accessible
- Ensure project/domain names are correct

### Cloud Not Found

```
Error: cloud "mycloud" not found
```

- Check `OS_CLOUD` matches a cloud name in `clouds.yaml`
- Verify `OS_CLIENT_CONFIG_FILE` points to correct file
- Check YAML syntax in `clouds.yaml`

### Permission Denied

```
Error: 403 Forbidden
```

- User may lack permissions for the requested resources
- `--all-tenants` requires admin credentials
- Check project roles and assignments

