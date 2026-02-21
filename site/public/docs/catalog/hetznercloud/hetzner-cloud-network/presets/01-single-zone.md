---
title: "Single-Zone Cloud Network"
description: "This preset creates a Hetzner Cloud private network with a single cloud subnet in the eu-central zone. It is the simplest usable network configuration -- one subnet providing private IPv4..."
type: "preset"
rank: "01"
presetSlug: "01-single-zone"
componentSlug: "hetzner-cloud-network"
componentTitle: "Hetzner Cloud Network"
provider: "hetznercloud"
icon: "package"
order: 1
---

# Single-Zone Cloud Network

This preset creates a Hetzner Cloud private network with a single cloud subnet in the eu-central zone. It is the simplest usable network configuration -- one subnet providing private IPv4 connectivity between servers, load balancers, and other resources in a single location. The /16 network CIDR leaves room to add more subnets later without replacing the network.

## When to Use

- New projects that need private connectivity between servers in one Hetzner location
- Development and staging environments where simplicity matters more than geographic reach
- Single-region production deployments where all resources share one network zone

## Key Configuration Choices

- **Single /24 subnet** (`ipRange: 10.0.1.0/24`) -- provides 254 usable host addresses, sufficient for most single-zone deployments; add more subnets as the project grows
- **eu-central zone** (`networkZone: eu-central`) -- Hetzner's largest zone with the most datacenter locations (Falkenstein, Nuremberg, Helsinki); change to `us-east`, `us-west`, or `ap-southeast` to match your server locations
- **No delete protection** -- allows easy teardown during development; enable `deleteProtection: true` before going to production
- **No custom routes** -- Hetzner's default routing handles traffic within the network; add routes only if you need a VPN gateway or NAT instance

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired network name.
