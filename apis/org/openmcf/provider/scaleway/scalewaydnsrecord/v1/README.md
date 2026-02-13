# Scaleway DNS Record

## Overview

The **ScalewayDnsRecord** resource kind provides a declarative interface for creating and managing individual DNS records within a Scaleway DNS zone. This is the **standalone** DNS record kind -- the DAG-friendly counterpart to ScalewayDnsZone's inline records.

Use this kind when record values come from other infrastructure resources, so that infra charts can express explicit dependency edges. Each ScalewayDnsRecord wraps a single `scaleway_domain_record` Terraform resource.

## Key Features

- **All 13 Scaleway record types** -- A, AAAA, ALIAS, CAA, CNAME, DNAME, MX, NS, PTR, SOA, SRV, TXT, TLSA
- **Cross-resource wiring** -- Both `zone_name` and `data` support `StringValueOrRef` for infra chart dependency graphs
- **Simpler than DigitalOcean** -- Scaleway embeds SRV weight/port and CAA flags/tag in the `data` field using standard DNS format, resulting in a cleaner spec
- **FQDN output** -- Scaleway computes the fully qualified domain name, eliminating manual string concatenation
- **Zone protection** -- `keep_empty_zone` prevents accidental zone deletion when destroying the last record
- **Dual IaC backend** -- Deploy using either Pulumi (Go) or Terraform with identical specifications

## Standalone vs Inline Records

| Approach | When to Use | DAG Visibility |
|---|---|---|
| **Standalone ScalewayDnsRecord** (this kind) | Records whose values come from other infrastructure resources (A records pointing to LB IPs, CNAMEs to cluster endpoints) | Each record is a separate DAG node with explicit dependency edges |
| **Inline ScalewayDnsZone records** | Static records known at zone creation time: MX, SPF, DMARC, CAA, domain verification TXT records | Records are part of the zone's lifecycle, not individually visible in the DAG |

Both approaches can be used together. A common pattern is to use inline records for email and security configuration, while using standalone records for infrastructure-dependent DNS entries.

## Scaleway Terraform Resource Mapping

| OpenMCF Kind | Terraform Resource | Relationship |
|---|---|---|
| ScalewayDnsRecord | `scaleway_domain_record` | 1:1 |

## Spec Fields

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `zone_name` | StringValueOrRef | Yes | -- | DNS zone name. Direct value or reference to ScalewayDnsZone output. |
| `name` | string | No | "" | Record name relative to zone. Empty = zone apex. Immutable. |
| `type` | RecordType | Yes | -- | DNS record type. Immutable. See record type reference below. |
| `data` | StringValueOrRef | Yes | -- | Record value. Direct value or cross-resource reference. |
| `ttl` | uint32 | No | 3600 | Time to live in seconds. |
| `priority` | uint32 | No | 0 | Priority for MX/SRV records. Lower = higher priority. |
| `keep_empty_zone` | bool | No | true | Keep zone when this is the last record destroyed. |

## Record Type Reference

Scaleway embeds type-specific fields (weight, port, flags, tag) in the `data` field using standard DNS record format. Only `priority` exists as a separate field.

| Type | Description | Data Format Example |
|---|---|---|
| **A** | IPv4 address | `"192.0.2.1"` |
| **AAAA** | IPv6 address | `"2001:db8::1"` |
| **ALIAS** | Auto-resolved alias (apex-safe) | `"www.example.com."` |
| **CAA** | Certificate Authority Authorization | `'0 issue "letsencrypt.org"'` |
| **CNAME** | Canonical name (alias) | `"target.example.com."` |
| **DNAME** | Delegation name (subtree redirect) | `"other.example.com."` |
| **MX** | Mail exchange | `"mail.example.com."` (use `priority` field) |
| **NS** | Nameserver | `"ns1.example.com."` |
| **PTR** | Pointer (reverse DNS) | `"host.example.com."` |
| **SOA** | Start of authority | SOA parameters |
| **SRV** | Service locator | `"10 5060 sipserver.example.com."` (weight port target) |
| **TXT** | Text record | `"v=spf1 include:_spf.google.com ~all"` |
| **TLSA** | TLS Association (DANE) | `"3 1 1 abcdef..."` (usage selector matching cert-data) |

## Stack Outputs

| Output | Description |
|---|---|
| `record_id` | Scaleway record ID (format: `{zone}/{id}`) |
| `fqdn` | Fully qualified domain name, computed by Scaleway |

## Dependencies

**Upstream:**
- **ScalewayDnsZone** -- `zone_name` references `status.outputs.zone_name`
- **Any resource** -- `data` can reference any resource's outputs (Load Balancer IPs, Instance public IPs, Kapsule cluster endpoints, etc.)

**Downstream:**
- Generally a leaf resource (DAG Layer 3+). The `fqdn` output is available for any downstream consumer that needs the full domain name.

## Important Constraints

### No Tags
Scaleway DNS records do not support tags in the API. Unlike most other Scaleway resources, the DNS service does not accept labels. The FQDN and `metadata.name` serve as the primary identifiers.

### Immutability
- **`name`** and **`type`** cannot be changed after creation (ForceNew). Changing them requires record recreation.
- **`data`**, **`ttl`**, and **`priority`** can be updated in-place.

### Zone Protection
- `keep_empty_zone` defaults to `true` (recommended). This prevents accidental zone deletion when destroying the last record in a zone.
- Set to `false` only when the zone itself is ephemeral and should be cleaned up automatically.

### No Advanced DNS Features
Scaleway DNS does not support: DNSSEC, traffic routing policies (geo, weighted, latency, failover), or health-check-based DNS. Dynamic record types (geo_ip, http_service, view, weighted) are supported by the Scaleway API but deferred from v1. For advanced DNS features, consider Cloudflare or AWS Route53.

## Scaleway Documentation

- [Scaleway Domains and DNS](https://www.scaleway.com/en/docs/network/domains-and-dns/)
- [Terraform: scaleway_domain_record](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/domain_record)
- [Pulumi: scaleway.domain.Record](https://www.pulumi.com/registry/packages/scaleway/api-docs/domain/record/)
