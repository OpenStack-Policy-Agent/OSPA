# Output Formats

OSPA supports multiple output formats for audit findings. This guide covers each format and how to work with the output.

## Available Formats

| Format | Flag | Best For |
|--------|------|----------|
| JSON | `--out-format json` | Programmatic processing, logging |
| CSV | `--out-format csv` | Spreadsheets, reporting |

## JSON Format

JSON Lines (JSON) is the default format. Each line is a complete JSON object.

### Usage

```bash
go run ./cmd/agent \
  --policy policy.yaml \
  --out findings.json \
  --out-format json
```

### Structure

Each line contains a finding record:

```json
{
  "rule_id": "critical-ssh-open-to-world",
  "resource_id": "abc123-def456",
  "resource_name": "ingress/tcp:22 from 0.0.0.0/0",
  "project_id": "project-123",
  "service": "neutron",
  "resource_type": "security_group_rule",
  "status": "ACTIVE",
  "compliant": false,
  "observation": "rule matches policy criteria",
  "action": "log",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `rule_id` | string | Policy rule name |
| `resource_id` | string | OpenStack resource ID |
| `resource_name` | string | Human-readable resource name |
| `project_id` | string | OpenStack project/tenant ID |
| `service` | string | OpenStack service name |
| `resource_type` | string | Resource type |
| `status` | string | Resource status |
| `compliant` | boolean | Whether resource is compliant |
| `observation` | string | Description of finding |
| `action` | string | Configured action |
| `timestamp` | string | ISO 8601 timestamp |
| `error` | string | Error message (if any) |

### Processing with jq

Filter violations:

```bash
# Only non-compliant
cat findings.json | jq 'select(.compliant == false)'

# Group by rule
cat findings.json | jq -s 'group_by(.rule_id) | .[] | {rule: .[0].rule_id, count: length}'

# List unique resource IDs
cat findings.json | jq -r '.resource_id' | sort -u

# Extract for reporting
cat findings.json | jq -r '[.rule_id, .resource_id, .observation] | @tsv'
```

## CSV Format

CSV format is useful for spreadsheets and traditional reporting tools.

### Usage

```bash
go run ./cmd/agent \
  --policy policy.yaml \
  --out findings.csv \
  --out-format csv
```

### Structure

```csv
rule_id,resource_id,resource_name,project_id,service,resource_type,status,compliant,observation,action,timestamp
critical-ssh-open-to-world,abc123,ingress/tcp:22,project-123,neutron,security_group_rule,ACTIVE,false,rule matches,log,2024-01-15T10:30:00Z
```

### Opening in Excel

1. Open Excel
2. File > Open > Select `findings.csv`
3. Follow import wizard, select comma delimiter

## Stdout Output

If `--out` is not specified, output goes to stdout:

```bash
# Pipe to jq
go run ./cmd/agent --policy policy.yaml | jq '.resource_id'

# Redirect manually
go run ./cmd/agent --policy policy.yaml > findings.json
```

## Summary Statistics

At the end of each run, OSPA prints a summary:

```
2024/01/15 10:30:01 Audit complete: 150 scanned, 12 violations, 0 errors
```

To capture this programmatically, parse stderr:

```bash
go run ./cmd/agent --policy policy.yaml --out findings.json 2>&1 | grep "Audit complete"
```

## Multiple Output Files

To output multiple formats, run twice or post-process:

```bash
# Run once
go run ./cmd/agent --policy policy.yaml --out findings.json

# Convert to CSV
cat findings.json | jq -r '[.rule_id, .resource_id, .compliant, .observation] | @csv' > findings.csv
```

## Integration Examples

### Splunk

```bash
# Forward JSON to Splunk HEC
go run ./cmd/agent --policy policy.yaml --out - | \
  while read line; do
    curl -k https://splunk:8088/services/collector/event \
      -H "Authorization: Splunk $HEC_TOKEN" \
      -d "{\"event\": $line}"
  done
```

### Elasticsearch

```bash
# Index findings to Elasticsearch
cat findings.json | while read line; do
  curl -X POST "http://elasticsearch:9200/ospa-findings/_doc" \
    -H "Content-Type: application/json" \
    -d "$line"
done
```

### Prometheus Pushgateway

The metrics endpoint provides real-time metrics. For batch push:

```bash
cat << EOF | curl --data-binary @- http://pushgateway:9091/metrics/job/ospa
# TYPE ospa_violations gauge
ospa_violations $(jq -s 'map(select(.compliant == false)) | length' findings.json)
EOF
```

### Slack Notification

```bash
# Count violations and notify
VIOLATIONS=$(jq -s 'map(select(.compliant == false)) | length' findings.json)
if [ "$VIOLATIONS" -gt 0 ]; then
  curl -X POST https://slack.com/api/chat.postMessage \
    -H "Authorization: Bearer $SLACK_TOKEN" \
    -d "channel=security" \
    -d "text=OSPA found $VIOLATIONS policy violations"
fi
```

## Best Practices

### 1. Use JSON for Automation

```bash
# Easy to process programmatically
go run ./cmd/agent --policy policy.yaml --out findings.json
```

### 2. Use CSV for Reporting

```bash
# Easy to share with stakeholders
go run ./cmd/agent --policy policy.yaml --out report.csv --out-format csv
```

### 3. Rotate Output Files

```bash
# Include date in filename
go run ./cmd/agent \
  --policy policy.yaml \
  --out "findings-$(date +%Y%m%d-%H%M%S).json"
```

### 4. Compress Large Files

```bash
# Compress after run
gzip findings.json
```

### 5. Validate Output

```bash
# Check JSON is valid
cat findings.json | jq empty && echo "Valid JSON"
```

