---
title: "StatefulSet with Persistent Volumes"
description: "This preset deploys a 3-replica StatefulSet with persistent volume claims, a pod disruption budget, and data volume mounts. Each replica gets its own 10Gi persistent volume."
type: "preset"
rank: "02"
presetSlug: "02-with-persistent-volumes"
componentSlug: "statefulset"
componentTitle: "StatefulSet"
provider: "kubernetes"
icon: "package"
order: 2
---

# StatefulSet with Persistent Volumes

This preset deploys a 3-replica StatefulSet with persistent volume claims, a pod disruption budget, and data volume mounts. Each replica gets its own 10Gi persistent volume.

## When to Use

- Stateful applications requiring data persistence across pod restarts
- Distributed systems that need stable storage per instance (e.g., custom databases, distributed caches)
- Workloads that need high availability with pod disruption budgets

## Key Configuration Choices

- **3 replicas** with PDB (`minAvailable: 2`) -- maintains quorum during rolling updates and node drains
- **10Gi PVC per replica** -- each pod gets a dedicated persistent volume; adjust size to your data requirements
- **Volume mounted at `/data`** -- standard data directory; your application should persist data here
- **ReadWriteOnce access** -- each volume is bound to a single pod; standard for block storage

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image repository | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |
| `<your-storage-class>` | Kubernetes StorageClass name (e.g., `gp3`, `standard-rwo`, `premium-rwo`) | `kubectl get storageclass` |

## Related Presets

- **01-standard** -- StatefulSet without persistent volumes
