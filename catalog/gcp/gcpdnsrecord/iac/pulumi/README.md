# GcpDnsRecord Pulumi Module

This Pulumi module creates one DNS record set in a Google Cloud DNS managed
zone — static values (round-robin) or exactly one routing policy (weighted
round robin, geolocation, or primary/backup failover). It also enables the
Cloud DNS API on the target project.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest dns-record.yaml
```

### Standalone Usage

1. Set the IaC input as an environment variable. The module reads a `GcpDnsRecordIacInput` from `IAC_INPUT_YAML` (YAML content), `IAC_INPUT_YAML_FILE` (a path to that YAML) or the Pulumi config key `planton:iac-input`; the manifest goes under `target`:

```bash
export IAC_INPUT_YAML=$(cat <<EOF
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpDnsRecord
  metadata:
    name: www-example
  spec:
    projectId:
      value: my-gcp-project
    managedZone:
      value: example-zone
    type: A
    name: www.example.com.
    values:
      - value: 192.0.2.1
    ttlSeconds: 300
EOF
)
```

2. Configure GCP credentials:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

3. Run Pulumi:

```bash
pulumi up
```

## Inputs

The module reads its configuration from the `GcpDnsRecordIacInput` proto message:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target | GcpDnsRecord | Yes | The GcpDnsRecord resource manifest |
| provider_config | GcpProviderConfig | Yes | GCP provider configuration |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| fqdn | string | The fully qualified domain name of the record |
| record_type | string | The DNS record type (A, AAAA, CNAME, etc.) |
| managed_zone | string | The managed zone containing the record |
| project_id | string | The GCP project ID |
| ttl_seconds | int | The TTL in seconds |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs.

## Troubleshooting

### Common Issues

1. **Record already exists**: Cloud DNS doesn't allow duplicate record sets with the same name and type. Delete the existing record first.

2. **Invalid DNS name**: Ensure the name ends with a trailing dot (e.g., `www.example.com.`).

3. **Zone not found**: Verify the managed zone exists in the specified project.
