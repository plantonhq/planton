---
title: "Standard VPC"
description: "This preset creates a DigitalOcean VPC with an explicit /16 CIDR block, providing a private isolated network for Droplets, Kubernetes clusters, databases, and load balancers within a single region...."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "digitalocean"
icon: "package"
order: 1
---

# Standard VPC

This preset creates a DigitalOcean VPC with an explicit /16 CIDR block, providing a private isolated network for Droplets, Kubernetes clusters, databases, and load balancers within a single region. This is the most common production VPC configuration.

## When to Use

- Any new DigitalOcean environment that needs private networking
- Production workloads requiring predictable, non-overlapping IP ranges
- Environments where multiple resources (Droplets, DOKS, databases) must communicate privately

## Key Configuration Choices

- **Explicit /16 CIDR** (`ipRangeCidr: 10.10.0.0/16`) -- provides 65,536 IPs, sufficient for most production environments. Adjust the second octet (`10.10`, `10.20`, etc.) to avoid overlap when running multiple VPCs.
- **Region** (`region: nyc1`) -- placeholder; change to your target region. DigitalOcean VPCs are regional and cannot span regions.
- **Not default for region** -- `isDefaultForRegion` omitted (defaults to `false`). Only set to `true` if this should be the auto-assigned VPC for new resources in the region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc1` | Target DigitalOcean region slug | [DigitalOcean Regions API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Regions) |
| `10.10.0.0/16` | VPC CIDR block (must be /16, /20, or /24) | Your IP address management plan |
