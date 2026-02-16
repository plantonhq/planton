---
title: "Development Kubernetes Cluster"
description: "This preset creates a minimal Scaleway Kapsule cluster with a shared (mutualized) control plane and a small 2-node default pool using DEV1-M instances. It is the fastest path to a working Kubernetes..."
type: "preset"
rank: "01"
presetSlug: "01-dev-minimal"
componentSlug: "kapsule-cluster"
componentTitle: "Kapsule Cluster"
provider: "scaleway"
icon: "package"
order: 1
---

# Development Kubernetes Cluster

This preset creates a minimal Scaleway Kapsule cluster with a shared (mutualized) control plane and a small 2-node default pool using DEV1-M instances. It is the fastest path to a working Kubernetes environment for development, testing, and learning.

## When to Use

- Development and staging environments
- Learning Kubernetes on Scaleway
- Running small workloads that do not need autoscaling or auto-upgrade

## Key Configuration Choices

- **Shared control plane** (`type: kapsule`) -- no additional cost for the API server; suitable for non-critical workloads
- **Cilium CNI** (`cni: cilium`) -- eBPF-based networking with Hubble observability; the Scaleway-recommended default
- **DEV1-M nodes** (`nodeType: DEV1-M`) -- 3 vCPU, 4 GB RAM per node; the most affordable option for running typical development workloads
- **Fixed 2-node pool** (`size: 2`) -- enough to run multiple pods with basic redundancy; no autoscaler overhead
- **No auto-upgrade** -- manual control over Kubernetes patch versions; add `autoUpgrade` for hands-off maintenance
- **Public IPs on nodes** (default) -- nodes are directly reachable; set `publicIpDisabled: true` for production security
- **Cleanup on delete** (`deleteAdditionalResources: true`) -- LBs, PVCs, and routes created by Kubernetes are cleaned up when the cluster is destroyed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network for cluster networking | Scaleway console or `ScalewayPrivateNetwork` status outputs |

## Related Presets

- **02-production-autoscaling** -- Use instead for production with autoscaling, auto-upgrade, private nodes, and autohealing
