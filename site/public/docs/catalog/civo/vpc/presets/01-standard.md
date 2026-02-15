---
title: "Standard VPC"
description: "This preset creates a standard Civo private network with an explicit /24 CIDR range in the London region. Suitable for most workloads where you need network isolation for instances, databases, and..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "civo"
icon: "package"
order: 1
---

# Standard VPC

This preset creates a standard Civo private network with an explicit /24 CIDR range in the London region. Suitable for most workloads where you need network isolation for instances, databases, and Kubernetes clusters.

## When to Use

- Any new project requiring a private network on Civo
- Isolating workloads from the default network
- Environments where instances, databases, and Kubernetes clusters need private connectivity

## Key Configuration Choices

- **Region** (`region: lon1`) -- London datacenter; change to match your preferred region (`fra1`, `nyc1`, `phx1`, `mum1`)
- **CIDR range** (`ipRangeCidr: 10.0.0.0/24`) -- explicit /24 block providing 254 usable addresses; omit to let Civo auto-allocate
- **Not default** (`isDefaultForRegion` omitted) -- does not replace the region's default network; set to `true` if you want this to be the default

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `my-vpc` | A descriptive name for your network | Your naming convention |
| `lon1` | Target Civo region | Civo dashboard or `civo region ls` |
| `10.0.0.0/24` | Private CIDR block (max /24 on Civo) | Your IP address plan |
