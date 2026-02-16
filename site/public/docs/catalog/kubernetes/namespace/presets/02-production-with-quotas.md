---
title: "Production Namespace with Custom Quotas"
description: "This preset creates a hardened production namespace with custom resource quotas, default container limits, network isolation, and restricted pod security. Designed for production workloads where..."
type: "preset"
rank: "02"
presetSlug: "02-production-with-quotas"
componentSlug: "namespace"
componentTitle: "Namespace"
provider: "kubernetes"
icon: "package"
order: 2
---

# Production Namespace with Custom Quotas

This preset creates a hardened production namespace with custom resource quotas, default container limits, network isolation, and restricted pod security. Designed for production workloads where resource governance and security are mandatory.

## When to Use

- Production environments requiring resource governance
- Multi-tenant clusters where namespace-level isolation is needed
- Workloads that must comply with restricted pod security standards

## Key Configuration Choices

- **Custom resource quotas** -- 4 CPU / 8Gi memory requests, 8 CPU / 16Gi memory limits; 50 pods max; adjust to your workload profile
- **Default container limits** -- every container without explicit limits gets 100m-500m CPU, 128Mi-512Mi memory; prevents runaway resource consumption
- **Network isolation** -- ingress restricted to `kube-system` and `istio-system` namespaces; egress restricted to `10.0.0.0/8` (cluster-internal)
- **Restricted pod security** -- the strictest built-in policy; requires non-root, read-only root filesystem, drops all capabilities

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-team-name>` | Team or project that owns this namespace | Your organization's team registry |

## Related Presets

- **01-standard** -- Minimal namespace with built-in resource profile for dev/staging
- **03-istio-enabled** -- Adds Istio service mesh sidecar injection
