# OciDnsZone

## Overview

OciDnsZone is an OpenMCF component that deploys an OCI DNS zone. It provides a single declarative manifest to create a managed authoritative DNS zone with support for public and private scopes, primary and secondary types, DNSSEC signing, and zone transfers.

## Purpose

OCI DNS provides managed authoritative DNS with full support for public and private zones. Public (GLOBAL) zones serve DNS queries from the internet, while private zones are resolvable only within VCNs via DNS views. Secondary zones replicate from external masters for hybrid cloud DNS architectures. This component provisions the zone; individual records are managed via the OciDnsRecord component.

## Key Features

- **Primary and secondary zone types** — PRIMARY zones are the authoritative source of truth; SECONDARY zones replicate from external masters.
- **Public and private scopes** — GLOBAL zones are publicly resolvable; PRIVATE zones resolve only within VCNs.
- **DNSSEC signing** — optional DNSSEC support with OCI-managed KSK and ZSK key pairs.
- **Zone transfers** — external masters for inbound replication (SECONDARY) and external downstreams for outbound replication (PRIMARY).
- **TSIG authentication** — optional TSIG key support for authenticated zone transfers.
- **Foreign key references** — `compartmentId` and `viewId` support `valueFrom` for infra-chart composability.

## Constraints

- `zoneType`, `scope`, `viewId`, and zone name (from `metadata.name`) are ForceNew — changing them destroys and recreates the zone.
- `compartmentId` is updatable (supports compartment moves).
- `externalMasters` and `externalDownstreams` are updatable.
- `isDnssecEnabled` is updatable (can toggle after creation).
- SECONDARY zones cannot have PRIVATE scope (OCI limitation).
- SECONDARY zones require at least one external master.
- PRIVATE zones require `viewId`.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Public domain hosting | Primary zone with GLOBAL scope |
| DNSSEC-signed domain | Primary zone with `isDnssecEnabled: true` |
| Internal service discovery | Private zone with VCN DNS view |
| Hybrid cloud DNS | Secondary zone replicating from on-premises masters |
| Multi-site DNS distribution | Primary zone with external downstreams |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **Nameserver output** — OCI-assigned authoritative nameservers exported for registrar configuration.
- **DNSSEC** — OCI-managed key generation and rotation for signed zones.
