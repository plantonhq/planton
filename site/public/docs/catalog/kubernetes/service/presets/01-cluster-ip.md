---
title: "ClusterIP Service"
description: "This preset creates a standard ClusterIP service that exposes a deployment internally within the cluster. The most common Kubernetes Service type for inter-service communication."
type: "preset"
rank: "01"
presetSlug: "01-cluster-ip"
componentSlug: "service"
componentTitle: "Service"
provider: "kubernetes"
icon: "package"
order: 1
---

# ClusterIP Service

This preset creates a standard ClusterIP service that exposes a deployment internally within the cluster. The most common Kubernetes Service type for inter-service communication.

## When to Use

- Internal service-to-service communication within the cluster
- Backend services that do not need direct external access
- Services fronted by an ingress controller for external traffic

## Key Configuration Choices

- **ClusterIP type** -- reachable only within the cluster; the default and most common service type
- **Port 80 -> 8080** -- clients connect to port 80, traffic is forwarded to container port 8080
- **Label selector** -- routes traffic to pods matching the `app` label

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the service | Your namespace management |
| `<your-app-label>` | Label value matching your deployment's pods | Your deployment manifest's `metadata.labels.app` |

## Related Presets

- **02-load-balancer** -- Exposes the service via a cloud load balancer with a public IP
- **03-headless** -- Headless service for StatefulSet DNS resolution
