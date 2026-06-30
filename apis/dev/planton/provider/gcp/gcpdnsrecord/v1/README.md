# GcpDnsRecord

## Overview

`GcpDnsRecord` is a Planton deployment component for managing individual DNS records within a Google Cloud DNS Managed Zone. It provides a declarative way to create, update, and manage DNS records across your GCP infrastructure.

## Purpose

This component simplifies DNS record management by:

- **Declarative Configuration**: Define DNS records as code using YAML manifests
- **Type Safety**: Strong validation ensures record configurations are correct before deployment
- **Multi-Record Support**: Create A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, and SOA records
- **Round-Robin**: Support multiple values for load distribution
- **Wildcard Records**: Create wildcard DNS entries for flexible subdomain routing

## Key Features

- ✅ **All Common Record Types**: Support for A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA
- ✅ **TTL Configuration**: Customizable time-to-live settings (1-86400 seconds)
- ✅ **Multiple Values**: Round-robin DNS with multiple record values
- ✅ **Wildcard Support**: Create `*.example.com.` records
- ✅ **Validation**: Built-in validation for DNS name formats and record configurations
- ✅ **Integration**: References existing GcpDnsZone resources or external zones

## Benefits

### Compared to Manual DNS Management
- **Version Control**: Track DNS changes in git
- **Consistency**: Apply the same DNS patterns across environments
- **Automation**: Integrate DNS provisioning into CI/CD pipelines

### Compared to Inline Zone Records
- **Modularity**: Manage records independently from zones
- **Flexibility**: Different teams can manage different records
- **Granularity**: Fine-grained access control per record

## Example Usage

### Basic A Record

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsRecord
metadata:
  name: www-example-com
spec:
  projectId: my-gcp-project
  managedZone: example-zone
  recordType: A
  name: www.example.com.
  values:
    - 192.0.2.1
```

### Deploy with CLI

```bash
planton pulumi up --manifest dns-record.yaml
```

Or with Terraform:

```bash
planton terraform apply --manifest dns-record.yaml
```

## Best Practices

1. **Always use FQDN**: DNS names must end with a trailing dot (e.g., `www.example.com.`)
2. **Set appropriate TTL**: Use lower TTL (60-300s) for records that may change frequently
3. **Use descriptive names**: Name your resources clearly (e.g., `api-production-a-record`)
4. **Validate before deploy**: Run `planton validate` to check your manifests

## Related Components

- **GcpDnsZone**: Parent resource for managing Cloud DNS zones
- **GcpProject**: GCP project where the DNS zone resides
