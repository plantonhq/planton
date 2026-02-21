---
title: "Public Primary DNS Zone"
description: "This preset creates a publicly resolvable, authoritative DNS zone hosted on OCI's managed DNS service. The zone is configured as PRIMARY (OCI is the source of truth for all records) with GLOBAL scope..."
type: "preset"
rank: "01"
presetSlug: "01-public-primary"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "oci"
icon: "package"
order: 1
---

# Public Primary DNS Zone

This preset creates a publicly resolvable, authoritative DNS zone hosted on OCI's managed DNS service. The zone is configured as PRIMARY (OCI is the source of truth for all records) with GLOBAL scope (resolvable from anywhere on the internet). This is the standard starting point for hosting DNS for any internet-facing domain -- web applications, APIs, email, and supporting services.

## When to Use

- Hosting DNS for a public domain (e.g., `example.com`) on OCI's managed nameservers
- Setting up DNS for internet-facing applications, APIs, or websites deployed on OCI
- Migrating an existing domain's DNS hosting to OCI from another provider
- Any scenario requiring authoritative public DNS resolution

## Key Configuration Choices

- **PRIMARY zone type** (`zoneType: primary`) -- OCI is the authoritative source for all records. Records are managed directly through OCI DNS (API, Console, or IaC). Use SECONDARY instead only when OCI should replicate from an external primary server.
- **GLOBAL scope** (`scope: global`) -- the zone is publicly resolvable from any DNS resolver on the internet. This is the correct choice for all internet-facing domains.
- **DNSSEC not enabled** -- DNSSEC is omitted by default. Enable it by adding `isDnssecEnabled: true` after the zone is created and nameserver delegation is confirmed. Enabling DNSSEC before delegation is complete can cause resolution failures.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the zone will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |

## Related Presets

- **02-private-vcn** -- use instead for internal DNS zones that should only be resolvable within a VCN (e.g., `internal.example.com` for microservice discovery)
