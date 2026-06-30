# Scaleway DNS Zone

## Overview

The **ScalewayDnsZone** resource kind provides a declarative interface for creating and managing DNS zones on Scaleway, with optional inline DNS records. A DNS zone represents a delegated portion of the DNS namespace for a domain you own, managed through Scaleway's Domains and DNS service.

This is a **composite resource** wrapping:
- `scaleway_domain_zone` (1x) -- the zone itself
- `scaleway_domain_record` (0..Nx) -- one per inline record entry

Scaleway DNS zones are created by specifying a parent domain and an optional subdomain prefix. The zone name is computed as `{subdomain}.{domain}` (or just `{domain}` for root zones).

## Key Features

- **Root and subdomain zones** -- Create zones for top-level domains (`example.com`) or delegated subdomains (`staging.example.com`, `internal.example.com`).
- **Inline DNS records** -- Bundle static records (MX, SPF/TXT, CAA, NS) directly in the zone spec for convenience.
- **Cross-resource references** -- Record data supports `StringValueOrRef`, allowing inline records to reference other resources' outputs (e.g., a Load Balancer's IP address).
- **Dual IaC backend** -- Deploy using either Pulumi (Go) or Terraform with identical specifications.
- **Dynamic nameservers** -- Nameservers are assigned by Scaleway at creation time and exported as outputs for registrar configuration.

## Two Ways to Manage Records

ScalewayDnsZone supports inline records for convenience, while the standalone **ScalewayDnsRecord** (R16) kind provides DAG-friendly record management for infra chart composition.

| Approach | When to Use | DAG Visibility |
|---|---|---|
| **Inline records** (this kind's `records` field) | Static records known at zone creation time: MX, SPF, DMARC, CAA, domain verification TXT records | Records are part of the zone's lifecycle, not individually visible in the infra chart DAG |
| **Standalone ScalewayDnsRecord** | Records whose values come from other infrastructure resources (A records pointing to LB IPs, CNAMEs to cluster endpoints) | Each record is a separate node in the DAG with explicit dependency edges |

Both approaches can be used together. A common pattern is to use inline records for email and security configuration, while using standalone records for infrastructure-dependent DNS entries.

## Scaleway Terraform Resource Mapping

| Planton Kind | Terraform Resource | Relationship |
|---|---|---|
| ScalewayDnsZone | `scaleway_domain_zone` | 1:1 (the zone) |
| ScalewayDnsZone | `scaleway_domain_record` | 1:N (inline records) |

## Spec Fields

### Zone Fields

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `domain` | string | Yes | -- | Registered parent domain (e.g., "example.com"). Immutable after creation. |
| `subdomain` | string | No | "" | Subdomain prefix. Empty = root zone. Updatable after creation. |
| `records` | repeated ScalewayDnsZoneRecord | No | [] | Inline DNS records. Empty = zone only. |

### Record Fields

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | No | "" | Record name relative to zone. Empty/"@" = zone apex. |
| `type` | RecordType | Yes | -- | DNS record type (A, AAAA, ALIAS, CAA, CNAME, DNAME, MX, NS, PTR, SOA, SRV, TXT, TLSA). |
| `data` | StringValueOrRef | Yes | -- | Record value. Literal or cross-resource reference. |
| `ttl` | uint32 | No | 3600 | Time to live in seconds. |
| `priority` | uint32 | No | 0 | Priority for MX/SRV records. Lower = higher priority. |

## Stack Outputs

| Output | Description |
|---|---|
| `zone_name` | Computed zone name (`{subdomain}.{domain}` or `{domain}`). Primary cross-resource reference for ScalewayDnsRecord. |
| `name_servers` | Nameservers assigned by Scaleway. Configure at domain registrar for delegation. |
| `name_servers_default` | Scaleway's default nameservers. |
| `name_servers_master` | Master nameservers. |
| `status` | Zone status (e.g., "active"). |

## Dependencies

**Upstream:** None. ScalewayDnsZone is a foundation resource (DAG Layer 0) with no `StringValueOrRef` inputs.

**Downstream:**
- **ScalewayDnsRecord** -- References `status.outputs.zone_name` to identify which zone records belong to.
- **Infra chart templates** -- Use `zone_name` output in `valueFrom` references for record creation.
- **Domain registrar** -- Uses `name_servers` output for NS delegation configuration.

## DNS Delegation Guide

After creating a zone, DNS queries will only resolve through Scaleway after you delegate the domain at your registrar:

1. **Create the zone** using this resource (apply the manifest)
2. **Retrieve nameservers** from `status.outputs.name_servers`
3. **Update nameservers at your registrar** (Namecheap, Google Domains, Cloudflare, etc.)
4. **Wait for propagation** -- DNS delegation typically takes 24-48 hours for full global propagation
5. **Verify** -- Use `dig @<scaleway-ns> example.com` to test resolution

For subdomain zones, configure NS records at the parent zone level instead of changing the registrar's nameservers.

## Important Constraints

### No Tags
Scaleway DNS zones and records do not support tags in the API. Unlike most other Scaleway resources, the DNS service does not accept labels. The zone name and `metadata.name` serve as the primary identifiers.

### Immutability
- **`domain`** cannot be changed after creation (ForceNew). Changing requires zone recreation.
- **`subdomain`** can be updated without recreation.
- **Record `name` and `type`** are immutable in Scaleway's API. Changing them forces record recreation.

### Record Type Coverage
The local `RecordType` enum covers all 13 Scaleway-supported record types: A, AAAA, ALIAS, CAA, CNAME, DNAME, MX, NS, PTR, SOA, SRV, TXT, TLSA.

### No Advanced DNS Features
Scaleway DNS does not support: DNSSEC, traffic routing policies (geo, weighted, latency, failover), or health-check-based DNS. For these features, consider Cloudflare or AWS Route53.

## What's Not Included (Deferred)

- **DNSSEC** -- Not available in Scaleway DNS.
- **Dynamic record types** -- Geo-IP, HTTP service checks, view-based routing, weighted routing are supported by Scaleway but deferred from v1.

## Scaleway Documentation

- [Scaleway Domains and DNS](https://www.scaleway.com/en/docs/network/domains-and-dns/)
- [Terraform: scaleway_domain_zone](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/domain_zone)
- [Terraform: scaleway_domain_record](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/domain_record)
