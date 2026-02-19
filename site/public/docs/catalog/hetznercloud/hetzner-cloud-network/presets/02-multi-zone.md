---
title: "Multi-Zone Cloud Network"
description: "This preset creates a Hetzner Cloud private network spanning two network zones, enabling private IPv4 connectivity between resources in geographically separate locations. Servers in eu-central can..."
type: "preset"
rank: "02"
presetSlug: "02-multi-zone"
componentSlug: "hetzner-cloud-network"
componentTitle: "Hetzner Cloud Network"
provider: "hetznercloud"
icon: "package"
order: 2
---

# Multi-Zone Cloud Network

This preset creates a Hetzner Cloud private network spanning two network zones, enabling private IPv4 connectivity between resources in geographically separate locations. Servers in eu-central can communicate with servers in us-east over the private network without traversing the public internet. Delete protection is enabled because multi-zone networks represent a deliberate production investment that should not be accidentally removed.

## When to Use

- Production services deployed across multiple Hetzner regions for geographic redundancy or latency optimization
- Applications with backend services in one region and edge/API servers in another that need private communication
- Any deployment where a single network zone is a single point of failure

## Key Configuration Choices

- **Two zones** (`eu-central` + `us-east`) -- provides cross-Atlantic private connectivity; substitute zones to match your server locations (`us-west`, `ap-southeast` are also available)
- **Non-overlapping /24 subnets** (`10.0.1.0/24`, `10.0.2.0/24`) -- each zone gets its own subnet with 254 usable addresses; the /16 network CIDR leaves room for additional subnets in either zone
- **Delete protection enabled** (`deleteProtection: true`) -- prevents accidental deletion of a production network; must be explicitly disabled before the network can be removed
- **No custom routes** -- Hetzner automatically routes traffic between subnets in the same network, even across zones

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired network name. Adjust `networkZone` values and subnet CIDRs to match your target regions.

## Related Presets

- **01-single-zone** -- simpler single-zone variant for projects that don't need cross-region connectivity
