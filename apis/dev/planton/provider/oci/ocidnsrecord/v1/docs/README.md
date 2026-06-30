# OciDnsRecord — Design Notes

## Design Rationale

OciDnsRecord manages a single DNS record set (RRSet) — all records sharing the same (domain, rtype) tuple within a zone. The component deviates from the standard OCI component pattern in several ways dictated by the underlying API.

### Why no compartmentId?

The `oci_dns_rrset` resource infers its compartment from the target zone. The `compartment_id` attribute is deprecated in both the Terraform and Pulumi providers. Including it would create a misleading field that the provider ignores.

### Why no stack outputs?

DNS record sets do not have their own OCID. They are identified by the (zone, domain, rtype) tuple — all of which are inputs to this component. There is no composable identifier that downstream components would reference, so the stack outputs proto is intentionally empty.

### Why no freeform tags?

OCI DNS record sets do not support tagging at the record level. Tags are only supported on DNS zones. This is a platform constraint, not a design choice.

### Why atomic set replacement?

The OCI DNS API manages record sets atomically — you cannot add or remove individual records within a set. Each update replaces the entire set for the (domain, rtype) tuple. This component reflects that API behavior directly. The benefit is consistency: there is no intermediate state where some records exist and others do not.

### Why is rtype a string instead of an enum?

DNS record types are IETF-standardized, numerous (A, AAAA, CNAME, MX, TXT, SRV, CAA, NS, PTR, SOA, NAPTR, TLSA, ...), and extensible as new RFCs are published. Modeling them as a proto enum would require constant updates. A plain string matches what the OCI API accepts and gives operators full flexibility.

### Why remove redundant domain/rtype from items?

The OCI provider requires each record item to carry its own `domain` and `rtype` fields, which must match the top-level values. This is an API design artifact — the values are redundant. This spec removes that redundancy; the IaC modules inject `domain` and `rtype` from the top-level fields into each item.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| No compartmentId | Avoids deprecated/ignored field | Breaks the universal OCI component pattern |
| No stack outputs | Honest API; no synthetic outputs | Cannot reference records via valueFrom |
| No freeform tags | Matches OCI API reality | No label propagation to records |
| Atomic set replacement | Consistent with OCI API; no partial states | Must redeclare all records on every update |
| rtype as string | Future-proof; matches IETF standards | No compile-time validation of record types |
| Remove item-level domain/rtype | Cleaner YAML; no redundancy | IaC module must inject values |

## Resource Graph

```
OciDnsRecord
└── oci_dns_rrset (always)
    ├── zone_name_or_id (from spec.zoneNameOrId)
    ├── domain (from spec.domain)
    ├── rtype (from spec.rtype)
    ├── view_id (optional, for private zones)
    └── items (1..N, each with rdata + ttl)
```

## Deferred from v1

- **Steering policies** — OCI DNS traffic management (weighted, geolocation, failover) is a separate service.
- **DNSSEC signing** — managed at the zone level via OciDnsZone, not at the record level.
- **Record-level tagging** — not supported by the OCI API.

## Notable API Behaviors

- **RDATA normalization** — OCI normalizes certain rdata values: IPv6 addresses are compressed, hostnames get trailing dots appended, TXT record quotes may be adjusted. The stored value may differ from the input.
- **SOA records** — managed automatically by OCI. Do not create OciDnsRecord manifests for SOA records.
- **NS records at zone apex** — managed by OCI for the zone's assigned nameservers. Custom NS records for delegation subzones are supported.
