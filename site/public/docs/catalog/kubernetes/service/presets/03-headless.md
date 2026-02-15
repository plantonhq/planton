---
title: "Headless Service"
description: "This preset creates a headless service (ClusterIP: None) for direct pod-to-pod DNS resolution. Essential for StatefulSets and any application that needs to discover individual pod IPs rather than a..."
type: "preset"
rank: "03"
presetSlug: "03-headless"
componentSlug: "service"
componentTitle: "Service"
provider: "kubernetes"
icon: "package"
order: 3
---

# Headless Service

This preset creates a headless service (ClusterIP: None) for direct pod-to-pod DNS resolution. Essential for StatefulSets and any application that needs to discover individual pod IPs rather than a virtual IP.

## When to Use

- StatefulSets that need per-pod DNS names (e.g., `my-app-0.my-headless-service.namespace.svc.cluster.local`)
- Client-side load balancing where the application discovers and connects to individual pods
- Peer-to-peer protocols (e.g., database replication, gossip protocols)

## Key Configuration Choices

- **Headless** (`headless: true`) -- sets ClusterIP to None; DNS returns individual pod IPs instead of a virtual IP
- **ClusterIP type** -- headless services are a special case of ClusterIP with no virtual IP assigned
- **Label selector** -- routes DNS queries to pods matching the `app` label

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the service | Your namespace management |
| `<your-app-label>` | Label value matching your StatefulSet or deployment pods | Your workload manifest's `metadata.labels.app` |

## Related Presets

- **01-cluster-ip** -- Standard ClusterIP service with a virtual IP
- **02-load-balancer** -- External LoadBalancer service
