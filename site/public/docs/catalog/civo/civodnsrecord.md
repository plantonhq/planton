---
title: "DNS Record"
description: "DNS Record deployment documentation"
icon: "package"
order: 100
componentName: "civodnsrecord"
---

# Civo DNS Record: Technical Research and Architecture

This document provides comprehensive research on DNS record management with Civo, covering the deployment landscape, architectural decisions, and production best practices that inform the CivoDnsRecord component.

## Table of Contents

1. [Introduction](#introduction)
2. [DNS Records Fundamentals](#dns-records-fundamentals)
3. [Civo DNS Service Overview](#civo-dns-service-overview)
4. [Deployment Methods](#deployment-methods)
5. [80/20 Scoping Decision](#8020-scoping-decision)
6. [Implementation Landscape](#implementation-landscape)
7. [Production Best Practices](#production-best-practices)
8. [OpenMCF's Approach](#openmcfs-approach)
9. [Common Pitfalls](#common-pitfalls)
10. [Conclusion](#conclusion)

---

## Introduction

DNS (Domain Name System) records are the fundamental building blocks that translate human-readable domain names into machine-readable IP addresses and route traffic to the appropriate destinations. Managing DNS records effectively is critical for any production infrastructure.

Civo, as a developer-friendly cloud provider, offers DNS management services that integrate with their broader cloud platform. This component (CivoDnsRecord) enables declarative management of individual DNS records within Civo-managed zones.

### Why Separate DNS Record Management?

While the CivoDnsZone component supports embedded record definitions, there are scenarios where managing records separately provides advantages:

1. **Independent Lifecycle**: Records can be created, updated, or deleted without affecting the zone
2. **Modular Configuration**: Teams can manage their own records without zone-level access
3. **Dynamic Updates**: Records can be modified by automation pipelines
4. **Cross-Stack References**: Records can reference outputs from other deployments

---

## DNS Records Fundamentals

### Record Types Overview

| Type | Purpose | Example Use Case |
|------|---------|------------------|
| **A** | Maps hostname to IPv4 address | `www → 192.0.2.1` |
| **AAAA** | Maps hostname to IPv6 address | `www → 2001:db8::1` |
| **CNAME** | Creates alias to another hostname | `app → www.example.com` |
| **MX** | Routes email to mail servers | `@ → mail.example.com` |
| **TXT** | Stores text data (SPF, DKIM, verification) | `@ → v=spf1 include:...` |
| **SRV** | Service locator records | `_sip._tcp → sipserver.example.com` |
| **NS** | Delegates zone to nameservers | `subdomain → ns1.example.com` |

### Record Structure

Each DNS record consists of:

1. **Name**: The hostname (relative to the zone apex)
2. **Type**: The record type (A, AAAA, CNAME, etc.)
3. **Value**: The record data (IP address, hostname, text)
4. **TTL**: Time-to-live in seconds (how long to cache)
5. **Priority**: For MX/SRV records (lower = higher priority)

### Special Name Values

| Name | Meaning |
|------|---------|
| `@` | Zone apex (root domain) |
| `*` | Wildcard (matches any subdomain) |
| `www` | Specific subdomain |

---

## Civo DNS Service Overview

### Platform Context

Civo is a UK-based cloud provider known for:
- Fast Kubernetes cluster provisioning
- Developer-friendly APIs
- Competitive pricing
- Growing global presence

Their DNS service integrates with the broader Civo ecosystem, allowing domains to be managed alongside compute, storage, and Kubernetes resources.

### Civo DNS Features

1. **Zone Management**: Create and manage DNS zones for domains
2. **Record Types**: Support for A, AAAA, CNAME, MX, TXT, SRV, NS
3. **API Access**: Full API for programmatic management
4. **Terraform Provider**: Official Civo Terraform provider
5. **Pulumi Support**: Via Pulumi Civo provider

### Civo DNS API

The Civo API provides endpoints for:

```
# Zone operations
GET    /v2/dns                    # List zones
POST   /v2/dns                    # Create zone
GET    /v2/dns/{zone_id}          # Get zone
DELETE /v2/dns/{zone_id}          # Delete zone

# Record operations
GET    /v2/dns/{zone_id}/records           # List records
POST   /v2/dns/{zone_id}/records           # Create record
GET    /v2/dns/{zone_id}/records/{id}      # Get record
PUT    /v2/dns/{zone_id}/records/{id}      # Update record
DELETE /v2/dns/{zone_id}/records/{id}      # Delete record
```

---

## Deployment Methods

### 1. Manual (Civo Dashboard)

The Civo web dashboard provides a UI for DNS management:
- Navigate to DNS → Select Zone → Add Record
- Fill in name, type, value, TTL, priority
- Click Create

**Pros**: Visual, immediate feedback
**Cons**: Manual, no version control, not reproducible

### 2. Civo CLI

```bash
# List DNS zones
civo dns list

# Create a record
civo dns record create my-zone.com \
  --name www \
  --type A \
  --value 192.0.2.1 \
  --ttl 3600

# List records in a zone
civo dns record list my-zone.com
```

**Pros**: Scriptable, quick operations
**Cons**: Imperative, no state management

### 3. Terraform

```hcl
resource "civo_dns_domain_record" "www" {
  domain_id = civo_dns_domain_name.example.id
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 3600
}
```

**Pros**: Declarative, state management, drift detection
**Cons**: HCL learning curve, state file management

### 4. Pulumi

```go
record, err := civo.NewDnsRecord(ctx, "www", &civo.DnsRecordArgs{
    DomainId: zone.ID(),
    Name:     pulumi.String("www"),
    Type:     pulumi.String("A"),
    Value:    pulumi.String("192.0.2.1"),
    Ttl:      pulumi.Int(3600),
})
```

**Pros**: Real programming language, type safety
**Cons**: More complex setup

### 5. Direct API

```bash
curl -X POST "https://api.civo.com/v2/dns/{zone_id}/records" \
  -H "Authorization: Bearer $CIVO_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "www",
    "type": "A",
    "value": "192.0.2.1",
    "ttl": 3600
  }'
```

**Pros**: Direct control, no dependencies
**Cons**: No state management, error-prone

---

## 80/20 Scoping Decision

### In-Scope (Essential Features)

Based on analysis of common DNS record use cases, these features cover ~80% of production needs:

| Feature | Rationale |
|---------|-----------|
| **A records** | Most common record type for web services |
| **AAAA records** | IPv6 support becoming standard |
| **CNAME records** | Essential for aliases and CDN integration |
| **MX records** | Required for email routing |
| **TXT records** | SPF, DKIM, DMARC, domain verification |
| **SRV records** | Service discovery (VoIP, XMPP, etc.) |
| **NS records** | Zone delegation |
| **TTL control** | Cache management (60-86400 seconds) |
| **Priority** | MX/SRV routing priority |

### Out-of-Scope (Advanced Features)

These features are less commonly needed and add complexity:

| Feature | Rationale |
|---------|-----------|
| **CAA records** | Not supported by Civo API |
| **NAPTR records** | Specialized telecom use case |
| **PTR records** | Reverse DNS, typically managed by IP provider |
| **Batch operations** | Can be achieved with multiple resources |
| **DNSSEC** | Not available on Civo platform |
| **Geo-routing** | Not supported by Civo |

---

## Implementation Landscape

### Pulumi Implementation

The Pulumi module uses the official `pulumi-civo` provider:

```go
import (
    "github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
)

// Create DNS record
record, err := civo.NewDnsRecord(ctx, name, &civo.DnsRecordArgs{
    DomainId: pulumi.String(spec.ZoneId),
    Name:     pulumi.String(spec.Name),
    Type:     pulumi.String(recordType),
    Value:    pulumi.String(spec.Value),
    Ttl:      pulumi.Int(spec.Ttl),
    Priority: pulumi.Int(spec.Priority),
})
```

Key considerations:
- Record type must be converted from enum to string
- Priority is optional (only for MX/SRV)
- TTL defaults to 3600 if not specified

### Terraform Implementation

The Terraform module uses the official `civo/civo` provider:

```hcl
resource "civo_dns_domain_record" "this" {
  domain_id = var.zone_id
  name      = var.name
  type      = var.type
  value     = var.value
  ttl       = var.ttl
  priority  = var.priority
}
```

Key considerations:
- `domain_id` refers to the zone ID
- Type is a string value
- Priority is optional

---

## Production Best Practices

### TTL Strategy

| Record Purpose | Recommended TTL | Rationale |
|----------------|-----------------|-----------|
| Static web servers | 3600-86400 | Rarely changes |
| Load balancers | 300-900 | May need failover |
| Development | 60-300 | Frequent changes |
| Migration prep | 60-300 | Reduce propagation delay |
| Email (MX) | 3600-86400 | Stability critical |

### Record Naming Conventions

```
# Good naming
www           # Main website
api           # API endpoint
app           # Application
cdn           # CDN origin
mail          # Mail server
staging-app   # Environment-prefixed

# Avoid
a             # Too cryptic
server1       # Infrastructure-focused
192-0-2-1     # IP in name
```

### Email Configuration Checklist

For professional email setup:

1. **MX Records**: Point to mail servers with priority
2. **SPF Record**: Authorize sending servers
3. **DKIM Record**: Email signing verification
4. **DMARC Record**: Policy for failed authentication
5. **Autodiscover**: For email client configuration

### High Availability

For critical services:

1. Multiple A records for round-robin
2. Low TTLs during changes
3. Monitor DNS resolution
4. Geographic distribution of nameservers

---

## OpenMCF's Approach

### Component Design

CivoDnsRecord follows OpenMCF's principles:

1. **Declarative API**: YAML manifests define desired state
2. **Kubernetes Resource Model**: Standard metadata/spec/status structure
3. **Validation**: Built-in proto validation rules
4. **Dual IaC**: Both Pulumi and Terraform implementations
5. **Idempotent**: Same manifest always produces same result

### Integration Points

CivoDnsRecord integrates with:

1. **CivoDnsZone**: Reference zone_id from zone outputs
2. **CivoKubernetesCluster**: Point records to cluster endpoints
3. **CivoLoadBalancer**: DNS for load balancer IPs

### Example Pipeline

```yaml
# 1. Create DNS Zone
apiVersion: civo.openmcf.org/v1
kind: CivoDnsZone
metadata:
  name: example-zone
spec:
  domain_name: example.com

---
# 2. Create A Record referencing zone
apiVersion: civo.openmcf.org/v1
kind: CivoDnsRecord
metadata:
  name: www-record
spec:
  zone_id: "${civo-dns-zone.example-zone.status.outputs.zone_id}"
  name: "www"
  type: A
  value: "192.0.2.1"
```

---

## Common Pitfalls

### 1. CNAME at Zone Apex

**Problem**: CNAME records at the zone apex (@) break RFC standards.
**Solution**: Use A/AAAA records at apex, or use providers with CNAME flattening.

### 2. Missing MX Priority

**Problem**: MX records without priority fail validation.
**Solution**: Always specify priority (10, 20, 30 for failover ordering).

### 3. TTL Too Low

**Problem**: Very low TTLs (< 60) cause excessive DNS queries.
**Solution**: Use minimum 60 seconds, prefer 300+ for stable records.

### 4. SPF Multiple Records

**Problem**: Multiple TXT records with SPF cause failures.
**Solution**: Consolidate into single SPF record using `include:`.

### 5. Propagation Expectations

**Problem**: Expecting instant DNS updates.
**Solution**: Allow up to TTL duration for propagation, use dig to verify.

### 6. Case Sensitivity

**Problem**: DNS names are case-insensitive but configs may differ.
**Solution**: Use lowercase consistently for all record names.

---

## Conclusion

CivoDnsRecord provides a clean, declarative interface for managing DNS records in Civo's cloud platform. By focusing on the essential 80% of use cases—standard record types, TTL control, and priority settings—this component enables teams to manage DNS infrastructure with the same rigor as other cloud resources.

Key takeaways:

1. **Start Simple**: Basic A/CNAME records for web services
2. **Email Requires Planning**: MX + SPF + DKIM + DMARC
3. **TTL Matters**: Balance cache efficiency vs. change agility
4. **Validate Before Production**: Use dig to verify resolution
5. **Version Control**: Treat DNS as code alongside application configs

For implementation details, see the IaC modules in `iac/pulumi/` and `iac/tf/`.

---

## References

- [Civo DNS Documentation](https://www.civo.com/docs/dns)
- [Civo API Reference](https://www.civo.com/api/dns)
- [Civo Terraform Provider](https://registry.terraform.io/providers/civo/civo/latest/docs)
- [RFC 1035 - DNS](https://tools.ietf.org/html/rfc1035)
- [RFC 7208 - SPF](https://tools.ietf.org/html/rfc7208)
