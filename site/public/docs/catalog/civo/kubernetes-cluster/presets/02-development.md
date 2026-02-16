---
title: "Development Kubernetes Cluster"
description: "This preset creates a minimal, cost-effective single-node Kubernetes cluster for development and testing. No HA, no auto-upgrade, smallest node size. Ideal for local development, CI pipelines, and..."
type: "preset"
rank: "02"
presetSlug: "02-development"
componentSlug: "kubernetes-cluster"
componentTitle: "Kubernetes Cluster"
provider: "civo"
icon: "package"
order: 2
---

# Development Kubernetes Cluster

This preset creates a minimal, cost-effective single-node Kubernetes cluster for development and testing. No HA, no auto-upgrade, smallest node size. Ideal for local development, CI pipelines, and proof-of-concept deployments.

## When to Use

- Development and testing environments
- CI/CD pipeline test clusters (create/destroy frequently)
- Learning and experimentation with Kubernetes on Civo
- Proof-of-concept deployments before scaling to production

## Key Configuration Choices

- **No HA** (`highlyAvailable` omitted) -- single control plane node to minimize cost
- **Single worker node** (`nodeCount: 1`) -- lowest cost; sufficient for dev workloads
- **Small node size** (`size: g4s.kube.small`) -- minimal resources; upgrade to medium when needed
- **No auto-upgrade** (`autoUpgrade` omitted) -- manual control over Kubernetes version in dev environments
- **VPC networking** (`network`) -- private networking even for dev clusters

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | `CivoVpc` status outputs |
| `1.29.2` | Kubernetes version | `civo kubernetes versions` |

## Related Presets

- **01-production-ha** -- Use instead for production workloads requiring HA and automatic upgrades
