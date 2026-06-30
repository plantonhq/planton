# GcpDnsRecord Pulumi Module

This Pulumi module creates and manages DNS records in Google Cloud DNS Managed Zones.

## Usage

This module is typically invoked by the Planton CLI, but can also be used directly.

### With Planton CLI

```bash
planton pulumi up --manifest dns-record.yaml
```

### Standalone Usage

1. Set the stack input as an environment variable:

```bash
export PLANTON_CLOUD_RESOURCE_MANIFEST=$(cat <<EOF
apiVersion: gcp.planton.dev/v1
kind: GcpDnsRecord
metadata:
  name: www-example
spec:
  projectId: my-gcp-project
  managedZone: example-zone
  recordType: A
  name: www.example.com.
  values:
    - 192.0.2.1
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

The module reads its configuration from the `GcpDnsRecordStackInput` proto message:

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

The GCP service account needs the following roles:
- `roles/dns.admin` - Full DNS management permissions

Or more restrictive:
- `roles/dns.recordset.editor` - For record set operations only

## Troubleshooting

### Common Issues

1. **Record already exists**: Cloud DNS doesn't allow duplicate record sets with the same name and type. Delete the existing record first.

2. **Invalid DNS name**: Ensure the name ends with a trailing dot (e.g., `www.example.com.`).

3. **Zone not found**: Verify the managed zone exists in the specified project.
