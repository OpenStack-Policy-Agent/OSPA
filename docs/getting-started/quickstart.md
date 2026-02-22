# Quick Start

Get OSPA running in under 5 minutes! This tutorial walks you through your first audit.

## Step 1: Set Up Credentials

Make sure your OpenStack credentials are configured:

```bash
# Set the path to your clouds.yaml
export OS_CLIENT_CONFIG_FILE=/path/to/clouds.yaml

# Set the cloud name to use
export OS_CLOUD=mycloud
```

??? tip "Sample clouds.yaml"

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
    ```

## Step 2: Create a Simple Policy

Create a file called `my-policy.yaml`:

```yaml
version: v1
defaults:
  workers: 10
policies:
  - neutron:
    - name: find-ssh-open-to-world
      description: Find SSH rules open to 0.0.0.0/0
      resource: security_group_rule
      check:
        direction: ingress
        protocol: tcp
        port: 22
        remote_ip_prefix: 0.0.0.0/0
      action: log
```

This policy will find any security group rules that allow SSH access from anywhere.

## Step 3: Run the Audit

Run OSPA in audit mode (safe, no changes):

```bash
go run ./cmd/agent \
  --cloud "$OS_CLOUD" \
  --policy my-policy.yaml \
  --out findings.json
```

You'll see output like:

```
2024/01/15 10:30:00 Loading policy from my-policy.yaml
2024/01/15 10:30:00 Starting audit with 10 workers
2024/01/15 10:30:01 Discovered 5 security_group_rule resources
2024/01/15 10:30:01 Audit complete: 5 scanned, 2 violations, 0 errors
```

## Step 4: Review Findings

Check the output file:

```bash
cat findings.json | jq .
```

Example output:

```json
{
  "rule_id": "find-ssh-open-to-world",
  "resource_id": "abc123",
  "resource_name": "ingress/tcp:22 from 0.0.0.0/0",
  "compliant": false,
  "observation": "rule matches policy criteria: [direction=ingress protocol=tcp port=22 remote_ip_prefix=0.0.0.0/0]"
}
```

## What's Next?

Now that you've run your first audit:

<div class="grid cards" markdown>

-  **Write More Policies**

    ---

    Learn the full policy syntax and available checks.

    [→ Writing Policies](../user-guide/policies.md)


</div>

## Common Audit Scenarios

### Find Unused Resources

```yaml
- neutron:
  - name: unused-security-groups
    description: Find security groups not attached to ports
    resource: security_group
    check:
      unused: true
      exempt_names:
        - default
    action: log
```

### Find Old Resources

```yaml
- cinder:
  - name: old-snapshots
    description: Find snapshots older than 90 days
    resource: snapshot
    check:
      age_gt: 90d
    action: log
```

### Find Resources by Status

```yaml
- nova:
  - name: shutoff-instances
    description: Find instances that are shut off
    resource: instance
    check:
      status: SHUTOFF
    action: log
```

