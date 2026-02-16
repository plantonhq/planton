---
title: "Custom Mode VPC with Regional Routing"
description: "This preset creates a VPC in custom subnet mode with regional routing and Private Services Access enabled. Custom mode gives full control over subnet CIDR ranges and regions. Private Services Access..."
type: "preset"
rank: "01"
presetSlug: "01-custom-mode-regional"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "gcp"
icon: "package"
order: 1
---

# Custom Mode VPC with Regional Routing

This preset creates a VPC in custom subnet mode with regional routing and Private Services Access enabled. Custom mode gives full control over subnet CIDR ranges and regions. Private Services Access enables private IP connectivity to Google managed services like Cloud SQL and Memorystore.

## When to Use

- Production workloads within a single GCP region
- Environments that need private connectivity to managed databases (Cloud SQL, Memorystore)
- Any project following the GCP best practice of custom-mode VPCs over auto-mode

## Key Configuration Choices

- **Custom mode** (`autoCreateSubnetworks: false`) -- no auto-created subnets; you define subnets explicitly with `GcpSubnetwork`
- **Regional routing** (`routingMode: REGIONAL`) -- Cloud Router advertises routes within one region only (simpler, lower cost)
- **Private Services Access enabled** -- creates VPC peering with Google's service network for private IPs
- **`/16` allocation for managed services** (`ipRangePrefixLength: 16`) -- 65,536 IPs reserved for Cloud SQL, Memorystore, etc.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the VPC will be created | GCP Console or `GcpProject` outputs |
| `<your-vpc-name>` | Name for this VPC network (1-63 chars, lowercase) | Choose a descriptive name (e.g., `prod-vpc`) |

## Related Presets

- **02-custom-mode-global** -- Use when workloads span multiple regions or require hybrid connectivity
