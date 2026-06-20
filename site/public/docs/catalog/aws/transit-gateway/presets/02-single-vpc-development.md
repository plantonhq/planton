---
title: "Single VPC Development"
description: "Minimal Transit Gateway with a single VPC attachment for development or testing. This is the simplest possible TGW setup, useful for validating connectivity patterns before scaling to production."
type: "preset"
rank: "02"
presetSlug: "02-single-vpc-development"
componentSlug: "transit-gateway"
componentTitle: "Transit Gateway"
provider: "aws"
icon: "package"
order: 2
---

# Single VPC Development

Minimal Transit Gateway with a single VPC attachment for development or testing. This is the simplest possible TGW setup, useful for validating connectivity patterns before scaling to production.

## When to Use

- Development or testing environments where you need a TGW to prototype
- Staging for future multi-VPC architecture
- Single-AZ cost-optimized setups

## Key Configuration Choices

- **Single VPC attachment** -- minimum viable setup
- **Single subnet** -- cost-optimized (no multi-AZ overhead)
- **Default routing enabled** -- ready for additional VPCs when needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<dev-vpc-id>` | Development VPC ID | AwsVpc status.outputs.vpc_id |
| `<dev-private-subnet>` | Dev VPC private subnet | AwsSubnet status.outputs.subnet_id |

## Related Presets

- **01-multi-vpc-hub** -- production multi-VPC pattern
