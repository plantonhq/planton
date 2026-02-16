---
title: "Internal-Only Router"
description: "This preset creates a router with no external gateway. It provides Layer 3 routing between connected subnets within the tenant but has no path to external networks. Use this for isolated environments..."
type: "preset"
rank: "02"
presetSlug: "02-internal-only"
componentSlug: "router"
componentTitle: "Router"
provider: "openstack"
icon: "package"
order: 2
---

# Internal-Only Router

This preset creates a router with no external gateway. It provides Layer 3 routing between connected subnets within the tenant but has no path to external networks. Use this for isolated environments or when external access is handled by a separate dedicated router.

## When to Use

- Air-gapped or isolated environments that must not reach the internet
- Internal routing between application tiers (e.g., web subnet to database subnet)
- Environments where a shared external router already exists and this router only handles east-west traffic

## Key Configuration Choices

- **No external gateway** -- no `externalNetworkId` configured, so no outbound internet access
- **No SNAT** -- not applicable without an external gateway
- **Admin state up** -- default (true), router forwards inter-subnet traffic immediately

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name`.

## Related Presets

- **01-edge-with-snat** -- Use instead when instances need outbound internet access
