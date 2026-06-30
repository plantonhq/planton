# Azure DNS Record

## Overview

The **AzureDnsRecord** deployment component creates individual DNS records within an existing Azure DNS Zone. This component provides a declarative way to manage DNS records across different record types (A, AAAA, CNAME, MX, TXT, NS, SRV, CAA, PTR) without the complexity of direct Azure Resource Manager API interactions.

## Purpose

This component simplifies DNS record management by:

- **Declarative Configuration**: Define DNS records as YAML manifests with full validation
- **Reference Resolution**: Wire records to existing DNS zones using `value_from` references
- **Type Safety**: Enforce correct record configurations through protobuf validation
- **Multi-Record Support**: Create any supported DNS record type with consistent patterns

## Key Features

- **All Standard Record Types**: A, AAAA, CNAME, MX, TXT, NS, SRV, CAA, PTR
- **Zone Reference**: Link to `AzureDnsZone` resources using `value_from` for zone_name
- **Configurable TTL**: Set custom TTL values per record (default: 300 seconds)
- **MX Priority**: Dedicated field for mail exchange record priorities
- **Dual IaC Support**: Both Pulumi and Terraform implementations

## Example Usage

### Basic A Record

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: www-record
spec:
  resource_group: my-dns-rg
  zone_name:
    value: example.com
  record_type: A
  name: www
  values:
    - "192.0.2.1"
  ttl_seconds: 300
```

### A Record with Zone Reference

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: api-record
spec:
  resource_group: my-dns-rg
  zone_name:
    value_from:
      name: my-azure-zone
  record_type: A
  name: api
  values:
    - "192.0.2.10"
```

### MX Record

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: mail-record
spec:
  resource_group: my-dns-rg
  zone_name:
    value: example.com
  record_type: MX
  name: "@"
  values:
    - "mail1.example.com"
    - "mail2.example.com"
  mx_priority: 10
```

## Deployment

Deploy using the Planton CLI:

```bash
# Using Pulumi
planton pulumi up --manifest dns-record.yaml

# Using Terraform/OpenTofu
planton tofu apply --manifest dns-record.yaml
```

## Best Practices

1. **Use Zone References**: When deploying multiple records to the same zone, use `value_from` to reference the zone resource
2. **Appropriate TTLs**: Use shorter TTLs (60-300s) for records you may need to change quickly
3. **Zone Apex**: Use `@` for the record name when targeting the zone apex (root domain)
4. **CNAME Restrictions**: Remember that CNAME records cannot coexist with other record types at the same name
