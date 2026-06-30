# OciDnsZone — Design Notes

## Design Rationale

OciDnsZone provisions a single DNS zone resource. The component supports the full range of OCI DNS zone capabilities: public and private scopes, primary and secondary types, DNSSEC, and zone transfers.

### Why use metadata.name as the zone name?

DNS zone names are globally significant identifiers (domain names). Using `metadata.name` as the zone name keeps the Planton resource identity aligned with the DNS domain. There is no separate `displayName` or `zoneName` field because the domain name is the natural identifier.

### Why are zoneType and scope enums with lowercase values?

Users type these values in YAML manifests. Lowercase (`primary`, `secondary`, `global`, `private`) is more natural to write and read than uppercase API values. The IaC module maps lowercase enum values to the uppercase strings the OCI API expects (`PRIMARY`, `SECONDARY`, `GLOBAL`, `PRIVATE`).

### Why enforce constraints via CEL instead of proto-level validation?

The validation constraints for DNS zones are cross-field: PRIVATE requires viewId, SECONDARY cannot be PRIVATE, SECONDARY requires external masters. Proto-level validation can only validate individual fields. CEL expressions operate on the full message and can express these cross-field constraints naturally.

### Why share the ExternalServer message between masters and downstreams?

External masters (inbound to SECONDARY zones) and external downstreams (outbound from PRIMARY zones) have the same structure: address, port, TSIG key. Sharing the message type avoids duplication and signals that the configuration pattern is the same.

### Why output nameservers as a comma-separated string?

OCI returns nameservers as a list of objects, but downstream consumers (registrar configuration, documentation) need a simple list of hostnames. The comma-separated string is the most portable format for passing the list through stack outputs. Individual nameservers can be split on comma by the consumer.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| metadata.name as zone name | Natural DNS identity | Cannot use a different resource name |
| Lowercase enum values | Better YAML readability | IaC module maps to uppercase |
| CEL cross-field validation | Catches invalid combos at validation time | More complex proto annotations |
| Shared ExternalServer message | DRY; consistent structure | Slightly less self-documenting per-field |
| Nameservers as comma-separated string | Portable; simple output type | Consumer must split |

## Resource Graph

```
OciDnsZone
└── oci_dns_zone (always)
    ├── zone_type (PRIMARY or SECONDARY)
    ├── scope (GLOBAL or PRIVATE, if set)
    ├── view_id (if private scope)
    ├── dnssec_state (ENABLED or DISABLED, if set)
    ├── external_masters (0..N, for SECONDARY zones)
    │   └── address, port, tsig_key_id
    ├── external_downstreams (0..N, for PRIMARY zones)
    │   └── address, port, tsig_key_id
    └── outputs: zone_id, nameservers
```

## Deferred from v1

- **dnssec_config** — deeply nested computed DNSSEC key version data (KSK/ZSK details). Operational concern for key rotation, not declarative configuration.
- **zone_transfer_servers** — computed list of OCI zone transfer servers. Not needed for composability.
- **is_protected** — read-only system flag indicating zone protection status.
- **DNSSEC key lifecycle actions** (stage/promote) — operational commands for key rotation.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciDnsZone` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
