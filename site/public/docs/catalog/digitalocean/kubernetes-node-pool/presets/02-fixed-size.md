---
title: "Fixed-Size System Node Pool"
description: "This preset creates a fixed-size node pool dedicated to system workloads (ingress controllers, monitoring agents, cluster add-ons). It uses a Kubernetes taint to prevent application pods from..."
type: "preset"
rank: "02"
presetSlug: "02-fixed-size"
componentSlug: "kubernetes-node-pool"
componentTitle: "Kubernetes Node Pool"
provider: "digitalocean"
icon: "package"
order: 2
---

# Fixed-Size System Node Pool

This preset creates a fixed-size node pool dedicated to system workloads (ingress controllers, monitoring agents, cluster add-ons). It uses a Kubernetes taint to prevent application pods from scheduling on these nodes, ensuring system components have guaranteed resources.

## When to Use

- Dedicated pool for cluster infrastructure (ingress-nginx, cert-manager, monitoring)
- Workload isolation where system pods must not compete with application pods
- Stable, predictable capacity requirements

## Key Configuration Choices

- **Fixed size** (`nodeCount: 2`) -- no autoscaling. System workloads have predictable resource needs.
- **Taint** (`dedicated=system:NoSchedule`) -- prevents application pods from scheduling unless they explicitly tolerate this taint.
- **System label** (`labels: {workload: system}`) -- used with `nodeAffinity` to target system deployments.
- **Smaller instances** (`size: s-2vcpu-4gb`) -- system workloads are typically lighter than application workloads.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cluster-name>` | Name of the parent DOKS cluster | `DigitalOceanKubernetesCluster` resource `metadata.name` |

## Related Presets

- **01-autoscaling-production** -- Use instead for application workloads requiring dynamic scaling
