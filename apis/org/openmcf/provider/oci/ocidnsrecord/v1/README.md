# OciDnsRecord

## Overview

OciDnsRecord is an OpenMCF component that deploys an OCI DNS Record Set (RRSet). It provides a single declarative manifest to create and manage a set of DNS resource records sharing the same domain and record type within an OCI DNS zone.

## Purpose

DNS record sets are the fundamental building blocks of DNS configuration. Each record set contains one or more records of the same type for a given domain name. This component manages the full record set atomically — updates replace all records for the (domain, rtype) tuple in a single operation, preventing partial updates that could cause DNS inconsistencies.

## Key Features

- **Atomic record set management** — all records for a (domain, rtype) tuple are managed as one unit; updates replace the entire set.
- **All standard record types** — supports A, AAAA, CNAME, MX, TXT, SRV, CAA, NS, PTR, and any other DNS record type.
- **Per-record TTL** — each record item can have its own TTL value.
- **Zone reference composability** — `zoneNameOrId` supports `valueFrom` to reference an OciDnsZone resource.
- **Private zone support** — optional `viewId` enables record management in private DNS zones.

## Constraints

- `zoneNameOrId`, `domain`, `rtype`, and `viewId` are ForceNew — changing them destroys and recreates the record set.
- `items` are updatable — the full set is replaced on each update.
- RDATA may be normalized by the OCI service (e.g., IPv6 compression, trailing-dot appended to hostnames, TXT quote handling) — returned values may differ from input.
- No `compartmentId` field — the `oci_dns_rrset` resource infers the compartment from the target zone.
- No freeform tags — DNS record sets do not support OCI tagging.
- No stack outputs — record sets are identified by their (zone, domain, rtype) input tuple.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Single A record | One item with IP address and TTL |
| Round-robin load balancing | Multiple A record items pointing to different IPs |
| Email routing | MX records with priority values in rdata |
| Domain aliases | CNAME record pointing to another hostname |
| Email authentication | TXT records for SPF, DKIM, DMARC |
| Service discovery | SRV records for protocol-level service location |

## Production Features

- **Atomic updates** — prevents partial DNS states that could cause routing issues during updates.
- **Zone reference via valueFrom** — enables composability with OciDnsZone in infra charts.
