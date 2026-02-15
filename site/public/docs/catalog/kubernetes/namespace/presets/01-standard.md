---
title: "Standard Namespace"
description: "This preset creates a Kubernetes namespace with a small built-in resource profile and baseline pod security. Suitable for most development and staging workloads where basic resource guardrails and..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "namespace"
componentTitle: "Namespace"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Namespace

This preset creates a Kubernetes namespace with a small built-in resource profile and baseline pod security. Suitable for most development and staging workloads where basic resource guardrails and security policies are sufficient.

## When to Use

- Development or staging environments
- Workloads that do not need custom resource quotas
- Teams that want a quick, safe starting point with reasonable defaults

## Key Configuration Choices

- **Resource profile** (`small`) -- applies built-in resource quotas appropriate for small workloads; other options: `medium`, `large`, `xlarge`
- **Pod security** (`baseline`) -- prevents known privilege escalations while remaining compatible with most workloads; stricter than `privileged`, more permissive than `restricted`
- **No network isolation** -- no network policies applied; ingress and egress are unrestricted

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-team-name>` | Team or project that owns this namespace | Your organization's team registry |

## Related Presets

- **02-production-with-quotas** -- Custom resource quotas, network isolation, and restricted pod security
- **03-istio-enabled** -- Adds Istio service mesh sidecar injection
