---
title: "Private GKE Cluster -- Rapid Channel"
description: "This preset creates a private GKE cluster on the RAPID release channel for development or staging environments that want early access to the latest Kubernetes versions and GKE features. Otherwise..."
type: "preset"
rank: "02"
presetSlug: "02-private-rapid"
componentSlug: "gke-cluster"
componentTitle: "GKE Cluster"
provider: "gcp"
icon: "package"
order: 2
---

# Private GKE Cluster -- Rapid Channel

This preset creates a private GKE cluster on the RAPID release channel for development or staging environments that want early access to the latest Kubernetes versions and GKE features. Otherwise identical to the standard production cluster.

## When to Use

- Dev or staging clusters where you want to test upcoming Kubernetes features before production
- Teams that need the latest GKE features (e.g., new node auto-provisioning, gateway API support)
- Pre-production validation of Kubernetes version upgrades

## Key Configuration Choices

- **RAPID release channel** (`releaseChannel: RAPID`) -- receives new Kubernetes versions weeks after they are available, before REGULAR
- **Private nodes** (`enablePublicNodes: false`) -- same security posture as production
- **Different master CIDR** (`172.16.0.16/28`) -- avoids overlap if sharing a VPC with a production cluster using `172.16.0.0/28`
- **Workload Identity and network policy** -- enabled by default (same as production)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<vpc-network-self-link>` | Self-link of the VPC network | `GcpVpc` status outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<subnet-self-link>` | Self-link of the GKE-ready subnet | `GcpSubnetwork` status outputs |
| `<your-cluster-name>` | GKE cluster name (1-40 chars, lowercase) | Choose a descriptive name |
| `<router-nat-name>` | Name of the Cloud NAT configuration | `GcpRouterNat` metadata name |

## Related Presets

- **01-private-standard** -- Use for production clusters with the more conservative REGULAR release channel
