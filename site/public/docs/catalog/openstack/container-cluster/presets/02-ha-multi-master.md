---
title: "HA Cluster (Multi-Master)"
description: "This preset creates a production Magnum container cluster with 3 master nodes and 3 worker nodes. The 3-master configuration provides etcd quorum and control plane high availability. Flavors are..."
type: "preset"
rank: "02"
presetSlug: "02-ha-multi-master"
componentSlug: "container-cluster"
componentTitle: "Container Cluster"
provider: "openstack"
icon: "package"
order: 2
---

# HA Cluster (Multi-Master)

This preset creates a production Magnum container cluster with 3 master nodes and 3 worker nodes. The 3-master configuration provides etcd quorum and control plane high availability. Flavors are explicitly set for masters and workers to allow independent sizing.

## When to Use

- Production Kubernetes clusters that need control plane high availability
- Workloads where master node failure should not disrupt the cluster
- Environments with multiple worker nodes for application capacity

## Key Configuration Choices

- **3 masters** (`masterCount: 3`) -- provides etcd quorum and survives single master failure
- **3 workers** (`nodeCount: 3`) -- baseline capacity; scale up by changing `nodeCount` (the only updatable field besides `clusterTemplate`)
- **Explicit flavors** -- masters and workers can use different instance sizes (e.g., lighter masters, heavier workers)
- **Template-driven** -- networking, COE, and image settings come from the referenced template

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<cluster-template-id>` | ID of the Magnum cluster template | OpenStack console or `OpenStackContainerClusterTemplate` status outputs |
| `<master-flavor-name>` | Flavor for master nodes (e.g., `m1.large`) | `openstack flavor list` |
| `<worker-flavor-name>` | Flavor for worker nodes (e.g., `m1.xlarge`) | `openstack flavor list` |

## Related Presets

- **01-dev-single-master** -- Use instead for development with minimal resource usage
