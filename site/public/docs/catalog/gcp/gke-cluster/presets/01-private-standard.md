---
title: "Private GKE Cluster -- Standard"
description: "This preset creates a private GKE cluster with no public node IPs, the REGULAR release channel, Workload Identity enabled, and network policy enforcement. Private clusters are the GCP-recommended..."
type: "preset"
rank: "01"
presetSlug: "01-private-standard"
componentSlug: "gke-cluster"
componentTitle: "GKE Cluster"
provider: "gcp"
icon: "package"
order: 1
---

# Private GKE Cluster -- Standard

This preset creates a private GKE cluster with no public node IPs, the REGULAR release channel, Workload Identity enabled, and network policy enforcement. Private clusters are the GCP-recommended configuration for production workloads, with internet egress routed through Cloud NAT.

## When to Use

- Production GKE deployments following GCP security best practices
- Clusters where nodes should not have public IP addresses
- Standard release cadence with balanced stability and feature availability

## Key Configuration Choices

- **Private nodes** (`enablePublicNodes: false`) -- nodes have no external IPs; outbound traffic goes through Cloud NAT
- **REGULAR release channel** (`releaseChannel: REGULAR`) -- automatic upgrades on a predictable cadence (2-3 months behind RAPID)
- **Workload Identity enabled** (default) -- `disableWorkloadIdentity` is not set, so pods can securely access GCP APIs via KSA-to-GSA bindings
- **Network policy enabled** (default) -- `disableNetworkPolicy` is not set, so Calico-based network policies are enforced
- **`/28` master CIDR** (`172.16.0.0/28`) -- private control plane endpoint; must not overlap with VPC or subnet ranges
- **Secondary ranges** -- references `pods` and `services` range names from the GKE-ready subnet preset

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

- **02-private-rapid** -- Use for dev/staging clusters that need the latest Kubernetes features
