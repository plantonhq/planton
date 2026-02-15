---
title: "General-Purpose Subnet"
description: "This preset creates a simple subnet for Compute Engine VMs, Cloud Run with VPC access, or other non-GKE workloads. No secondary IP ranges are defined since alias IPs are not needed outside of GKE."
type: "preset"
rank: "02"
presetSlug: "02-general-purpose"
componentSlug: "subnetwork"
componentTitle: "Subnetwork"
provider: "gcp"
icon: "package"
order: 2
---

# General-Purpose Subnet

This preset creates a simple subnet for Compute Engine VMs, Cloud Run with VPC access, or other non-GKE workloads. No secondary IP ranges are defined since alias IPs are not needed outside of GKE.

## When to Use

- Subnets for Compute Engine instances or instance groups
- Cloud Run services that need VPC connectivity (Serverless VPC Access)
- Any workload that does not require GKE secondary IP ranges

## Key Configuration Choices

- **`/24` CIDR range** (`10.1.0.0/24`) -- 256 IPs, suitable for small to medium compute deployments
- **No secondary ranges** -- not needed for non-GKE workloads
- **Private Google Access** (`privateIpGoogleAccess: true`) -- VMs without external IPs can access Google APIs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | GCP Console or `GcpProject` outputs |
| `<vpc-network-self-link>` | Self-link of the parent VPC network | `GcpVpc` status outputs |
| `<your-subnet-name>` | Name for this subnet (1-63 chars, lowercase) | Choose a descriptive name (e.g., `compute-us-central1`) |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |

## Related Presets

- **01-gke-ready** -- Use when the subnet will host a GKE cluster (requires secondary IP ranges)
