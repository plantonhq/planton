---
title: "Internal Service Discovery"
description: "This preset creates a private DNS zone for internal service discovery within a single VPC. Services register A records pointing to their private IP addresses, enabling hostname-based discovery..."
type: "preset"
rank: "01"
presetSlug: "01-internal-service-discovery"
componentSlug: "private-zone"
componentTitle: "Private Zone"
provider: "alicloud"
icon: "package"
order: 1
---

# Internal Service Discovery

This preset creates a private DNS zone for internal service discovery within a single VPC. Services register A records pointing to their private IP addresses, enabling hostname-based discovery without relying on IP addresses.

## When to Use

- Microservices that need to discover each other by hostname within a VPC
- Internal APIs, databases, and caches that should be reachable by friendly names
- Development and staging environments where external DNS is unnecessary
- Replacing hardcoded IP addresses with stable hostnames

## Key Configuration Choices

- **Single VPC** -- the zone is attached to one VPC. Add more `vpcAttachments` entries to share across VPCs.
- **A records** -- straightforward hostname-to-IP mapping. Use CNAME records for aliases.
- **Default TTL (60s)** -- suitable for services whose IPs may change during deployments. Increase TTL for stable endpoints.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region |
| `<zone-name>` | Private zone name (e.g., `svc.internal`, `services.corp`) | Your naming convention |
| `<vpc-id>` | VPC ID to attach the zone to | VPC console or AliCloudVpc output |
| `<service-name>` | Service hostname prefix (e.g., `api`, `cache`, `db`) | Your service inventory |
| `<private-ip>` | Private IP address of the service | ECS/container private IP |

## Post-Deployment Steps

1. Deploy the manifest to create the private zone and VPC attachment
2. Resources in the VPC can immediately resolve `<service-name>.<zone-name>` to `<private-ip>`
3. Add more records by updating `spec.records` and redeploying

## Related Presets

- **02-multi-vpc-database-zone** -- use when the zone needs to be shared across multiple VPCs or regions
