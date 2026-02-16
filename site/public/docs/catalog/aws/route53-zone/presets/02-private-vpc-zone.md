---
title: "Private VPC DNS Zone"
description: "This preset creates a private Route53 hosted zone that resolves DNS queries only within associated VPCs. Private zones enable split-horizon DNS, where internal services use private domain names..."
type: "preset"
rank: "02"
presetSlug: "02-private-vpc-zone"
componentSlug: "route53-zone"
componentTitle: "Route53 Zone"
provider: "aws"
icon: "package"
order: 2
---

# Private VPC DNS Zone

This preset creates a private Route53 hosted zone that resolves DNS queries only within associated VPCs. Private zones enable split-horizon DNS, where internal services use private domain names (e.g., `db.internal.example.com`) that are not resolvable from the public internet.

## When to Use

- Internal service discovery within a VPC (e.g., `api.internal.example.com` resolving to private IPs)
- Split-horizon DNS where internal and external clients resolve the same domain to different addresses
- Private DNS for databases, caches, and other backend services that should not have public DNS records

## Key Configuration Choices

- **Private zone** (`isPrivate: true`) -- DNS records resolve only within associated VPCs; invisible to the public internet
- **Single VPC association** -- Associates the zone with one VPC; add more entries to `vpcAssociations` for multi-VPC environments
- **DNS support required** -- The associated VPC must have `enableDnsHostnames` and `enableDnsSupport` enabled

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<vpc-id>` | VPC ID to associate with this private zone (e.g., `vpc-0123456789abcdef0`) | AWS VPC console or `AwsVpc` status outputs |
| `<aws-region>` | AWS region where the VPC is located (e.g., `us-east-1`) | Your deployment region |

## Related Presets

- **01-public-zone** -- Use instead for internet-facing domains that need globally resolvable DNS records
