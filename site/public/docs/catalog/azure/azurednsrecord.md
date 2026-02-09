---
title: "DNS Record"
description: "DNS Record deployment documentation"
icon: "package"
order: 100
componentName: "azurednsrecord"
---

# Azure DNS Record - Technical Research Documentation

## Introduction

Azure DNS is Microsoft's authoritative DNS hosting service, providing name resolution using Microsoft Azure infrastructure. Azure DNS Records are individual entries within a DNS zone that map domain names to various resources like IP addresses, mail servers, or other domains.

This document provides comprehensive research on deploying and managing Azure DNS Records, comparing different approaches and explaining the design decisions behind OpenMCF's implementation.

## DNS Record Fundamentals

### What is a DNS Record?

A DNS record is a database entry that provides information about a domain, including its associated IP addresses, mail servers, and other configuration data. When users query a domain name, DNS resolvers use these records to route traffic appropriately.

### Common Record Types

| Type | Purpose | Example Use Case |
|------|---------|------------------|
| **A** | Maps hostname to IPv4 address | `www.example.com → 192.0.2.1` |
| **AAAA** | Maps hostname to IPv6 address | `www.example.com → 2001:db8::1` |
| **CNAME** | Creates alias to another domain | `blog.example.com → example.ghost.io` |
| **MX** | Specifies mail servers | Email routing for domain |
| **TXT** | Stores text data | SPF, DKIM, domain verification |
| **NS** | Delegates subdomain | Subdomain to different nameservers |
| **SRV** | Service location | SIP, LDAP service discovery |
| **CAA** | Certificate authority authorization | SSL/TLS issuance control |
| **PTR** | Reverse DNS lookup | IP to hostname mapping |

## Deployment Methods Landscape

### 1. Azure Portal (Manual)

**Pros:**
- Visual interface
- No coding required
- Good for learning

**Cons:**
- Not repeatable
- No version control
- Error-prone for bulk operations
- No audit trail

### 2. Azure CLI

```bash
az network dns record-set a add-record \
  --resource-group MyResourceGroup \
  --zone-name example.com \
  --record-set-name www \
  --ipv4-address 192.0.2.1
```

**Pros:**
- Scriptable
- Can be version controlled

**Cons:**
- Imperative (describes actions, not state)
- Difficult to manage dependencies
- No drift detection

### 3. ARM Templates

```json
{
  "type": "Microsoft.Network/dnsZones/A",
  "apiVersion": "2018-05-01",
  "name": "example.com/www",
  "properties": {
    "TTL": 300,
    "ARecords": [
      { "ipv4Address": "192.0.2.1" }
    ]
  }
}
```

**Pros:**
- Declarative
- Native Azure support
- Version controllable

**Cons:**
- Verbose JSON syntax
- Limited reusability
- Azure-specific

### 4. Terraform

```hcl
resource "azurerm_dns_a_record" "www" {
  name                = "www"
  zone_name           = "example.com"
  resource_group_name = "my-rg"
  ttl                 = 300
  records             = ["192.0.2.1"]
}
```

**Pros:**
- Declarative
- State management
- Drift detection
- Multi-cloud support

**Cons:**
- HCL learning curve
- State file management
- Manual wiring between resources

### 5. Pulumi

```go
record, err := dns.NewARecord(ctx, "www", &dns.ARecordArgs{
    Name:              pulumi.String("www"),
    ZoneName:          pulumi.String("example.com"),
    ResourceGroupName: pulumi.String("my-rg"),
    Ttl:               pulumi.Int(300),
    Records:           pulumi.StringArray{pulumi.String("192.0.2.1")},
})
```

**Pros:**
- Real programming languages
- Type safety
- IDE support
- Reusable components

**Cons:**
- Requires programming knowledge
- More complex setup

## OpenMCF's Approach

### Design Philosophy

OpenMCF's AzureDnsRecord component follows the 80/20 principle:
- **80%** of DNS record use cases are covered
- **20%** of configuration options exposed

### Key Design Decisions

#### 1. Zone Reference via `value_from`

The `zone_name` field supports both literal values and references to `AzureDnsZone` resources:

```yaml
# Literal value
zone_name:
  value: example.com

# Reference to zone resource
zone_name:
  value_from:
    name: my-azure-zone
```

This enables:
- **Loose coupling**: Records can be deployed independently
- **Dependency resolution**: CLI resolves zone name at deploy time
- **Flexibility**: Mix literal and referenced values

#### 2. Single Record Per Resource

Unlike the `AzureDnsZone` component which can create multiple records, `AzureDnsRecord` creates exactly one record. This provides:
- **Atomic operations**: Each record has its own lifecycle
- **Fine-grained control**: Update individual records without affecting others
- **Clear ownership**: One manifest = one DNS record

#### 3. Unified Record Type Handling

All record types use the same `values` field, with type-specific interpretation:
- A/AAAA: IP addresses
- CNAME: Target hostname (first value only)
- MX: Mail server hostnames (priority via `mx_priority`)
- TXT: Text values
- etc.

### What's NOT Included (By Design)

1. **Traffic Manager integration**: Complex routing scenarios have dedicated components
2. **Private DNS zones**: Separate component for VNet-scoped DNS
3. **Alias records**: Azure-specific alias records not yet supported
4. **Record set grouping**: Each record type at a name gets its own resource

## Azure DNS Record Specifics

### Record Naming

- `@` - Zone apex (root domain)
- `*` - Wildcard
- Subdomain names without trailing dot

### TTL Guidelines

| Scenario | Recommended TTL |
|----------|-----------------|
| Frequently changing | 60-300 seconds |
| Standard records | 300-3600 seconds |
| Stable records | 3600-86400 seconds |
| NS delegations | 86400 seconds |

### Limits and Quotas

- Max 10,000 record sets per zone
- Max 20 records per record set
- Max 500 DNS zones per subscription (soft limit)

## Production Best Practices

### 1. Use Appropriate TTLs

Lower TTLs = faster propagation but more DNS queries (cost)
Higher TTLs = better caching but slower changes

### 2. CAA Records

Always configure CAA records to control certificate issuance:

```yaml
record_type: CAA
name: "@"
values:
  - "letsencrypt.org"
```

### 3. SPF/DKIM/DMARC

For email deliverability, always configure:
- SPF (TXT record)
- DKIM (TXT record)  
- DMARC (TXT record at `_dmarc`)

### 4. Monitoring

- Enable Azure DNS Analytics
- Monitor query volume
- Alert on failed queries

## Implementation Details

### Pulumi Module

The Pulumi module:
1. Creates Azure provider with credentials from `ProviderConfig`
2. Switches on `record_type` to call appropriate DNS record constructor
3. Exports `record_id` and `fqdn` as stack outputs

### Terraform Module

The Terraform module:
1. Uses conditional resources (`count`) based on record type
2. Creates exactly one record resource per deployment
3. Outputs record ID and FQDN using `coalesce` for type-agnostic output

## References

- [Azure DNS Documentation](https://docs.microsoft.com/en-us/azure/dns/)
- [Azure DNS REST API](https://docs.microsoft.com/en-us/rest/api/dns/)
- [Terraform azurerm_dns_* resources](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- [Pulumi Azure DNS resources](https://www.pulumi.com/registry/packages/azure/api-docs/dns/)

## Conclusion

The AzureDnsRecord component provides a streamlined, declarative interface for managing DNS records in Azure. By supporting zone references through `value_from`, it integrates seamlessly with the broader OpenMCF ecosystem while maintaining the flexibility needed for diverse DNS configurations.
