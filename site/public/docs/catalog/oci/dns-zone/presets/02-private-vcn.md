---
title: "Private VCN DNS Zone"
description: "This preset creates a private DNS zone resolvable only within VCNs attached to the specified DNS view. Private zones enable internal service discovery without exposing hostnames to the public..."
type: "preset"
rank: "02"
presetSlug: "02-private-vcn"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "oci"
icon: "package"
order: 2
---

# Private VCN DNS Zone

This preset creates a private DNS zone resolvable only within VCNs attached to the specified DNS view. Private zones enable internal service discovery without exposing hostnames to the public internet -- microservices, databases, and internal tools can be addressed by friendly names (e.g., `db.internal.example.com`) that resolve only from within the VCN.

## When to Use

- Internal service discovery for microservices within a VCN (e.g., `api.internal.example.com`)
- Providing friendly DNS names for private databases, caches, and backend services
- Split-horizon DNS where internal clients should resolve a domain differently from external clients
- Any scenario where DNS names must not be publicly resolvable

## Key Configuration Choices

- **PRIMARY zone type** (`zoneType: primary`) -- OCI is the authoritative source. SECONDARY zones cannot be private (OCI limitation enforced by CEL validation in the spec).
- **PRIVATE scope** (`scope: private`) -- the zone is only resolvable from VCNs whose DNS resolver is associated with the specified view. Queries from the public internet receive NXDOMAIN.
- **DNS view** (`viewId`) -- the view controls which VCNs can resolve records in this zone. Each VCN's DNS resolver references a view; all private zones attached to that view become resolvable from the VCN. Multiple VCNs can share a view for cross-VCN resolution.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the zone will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<dns-view-ocid>` | OCID of the private DNS view this zone belongs to | OCI Console > DNS Management > Views, or VCN resolver configuration |

## Related Presets

- **01-public-primary** -- use instead for internet-facing domains that must be publicly resolvable
