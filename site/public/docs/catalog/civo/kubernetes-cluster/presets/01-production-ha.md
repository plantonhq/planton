---
title: "Production HA Kubernetes Cluster"
description: "This preset creates a highly available Civo Kubernetes (K3s) cluster with 3 worker nodes, automatic patch upgrades, and VPC networking. This is the most common production configuration, providing..."
type: "preset"
rank: "01"
presetSlug: "01-production-ha"
componentSlug: "kubernetes-cluster"
componentTitle: "Kubernetes Cluster"
provider: "civo"
icon: "package"
order: 1
---

# Production HA Kubernetes Cluster

This preset creates a highly available Civo Kubernetes (K3s) cluster with 3 worker nodes, automatic patch upgrades, and VPC networking. This is the most common production configuration, providing resilience against node failures and zero-downtime Kubernetes upgrades.

## When to Use

- Production workloads requiring high availability
- Applications that need to survive node failures without downtime
- Environments where automatic Kubernetes patch upgrades reduce operational burden

## Key Configuration Choices

- **Highly available** (`highlyAvailable: true`) -- multiple control plane nodes for fault tolerance
- **3 worker nodes** (`nodeCount: 3`) -- allows pod scheduling across nodes for resilience; minimum for production HA
- **Medium node size** (`size: g4s.kube.medium`) -- balanced CPU/RAM for general workloads; scale up for compute-intensive applications
- **Auto-upgrade** (`autoUpgrade: true`) -- Civo applies Kubernetes patch updates automatically, reducing operational overhead
- **VPC networking** (`network`) -- cluster traffic stays within the private network
- **Kubernetes 1.29** (`kubernetesVersion`) -- latest stable; update to the current supported version on Civo

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | `CivoVpc` status outputs |
| `1.29.2` | Kubernetes version (check Civo's supported versions) | `civo kubernetes versions` |

## Related Presets

- **02-development** -- Use instead for non-HA, single-node dev/test clusters at lower cost
