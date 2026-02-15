---
title: "All-Subnets Auto-Allocated NAT"
description: "This preset creates a Cloud Router with a NAT gateway that covers all subnets in the region using automatically allocated external IPs. This is the simplest and most common Cloud NAT configuration,..."
type: "preset"
rank: "01"
presetSlug: "01-all-subnets-auto"
componentSlug: "router-nat"
componentTitle: "Router NAT"
provider: "gcp"
icon: "package"
order: 1
---

# All-Subnets Auto-Allocated NAT

This preset creates a Cloud Router with a NAT gateway that covers all subnets in the region using automatically allocated external IPs. This is the simplest and most common Cloud NAT configuration, ideal for giving private GKE nodes or Compute Engine VMs outbound internet access.

## When to Use

- Private GKE clusters that need internet egress for container image pulls
- Compute Engine VMs without external IPs that need outbound connectivity
- Any VPC where all subnets in a region should share NAT egress

## Key Configuration Choices

- **All subnets covered** -- `subnetworkSelfLinks` is empty, so NAT applies to every subnet in the region
- **Auto-allocated IPs** -- `natIpNames` is empty, so GCP automatically provisions and manages external IPs
- **Error-only logging** (`logFilter: ERRORS_ONLY`) -- logs port exhaustion and connection failures without the volume of full translation logging

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | GCP Console or `GcpProject` outputs |
| `<vpc-network-self-link>` | Self-link of the VPC network | `GcpVpc` status outputs |
| `<gcp-region>` | GCP region matching your subnets (e.g., `us-central1`) | Your deployment region |
| `<your-router-name>` | Name for the Cloud Router (1-63 chars, lowercase) | Choose a descriptive name (e.g., `prod-router`) |
| `<your-nat-name>` | Name for the NAT configuration (1-63 chars, lowercase) | Choose a descriptive name (e.g., `prod-nat`) |

## Related Presets

- **02-static-ip-specific-subnets** -- Use when you need stable egress IPs for partner allowlisting or compliance
