---
title: "Standard Gateway API CRDs"
description: "This preset installs the Kubernetes Gateway API CRDs in the standard channel. The Gateway API is the next-generation Kubernetes API for managing ingress and service mesh traffic, replacing the legacy..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "gateway-api-crds"
componentTitle: "Gateway API CRDs"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Gateway API CRDs

This preset installs the Kubernetes Gateway API CRDs in the standard channel. The Gateway API is the next-generation Kubernetes API for managing ingress and service mesh traffic, replacing the legacy Ingress resource.

## When to Use

- You want to use Gateway, GatewayClass, HTTPRoute, and ReferenceGrant resources
- You need the stable/standard set of Gateway API resources
- Your service mesh or ingress controller supports the Gateway API (e.g., Istio, Envoy Gateway, NGINX Gateway Fabric)

## Key Configuration Choices

- **Standard channel** -- includes stable resources: Gateway, GatewayClass, HTTPRoute, ReferenceGrant
- **Version** (`v1.2.1`) -- pinned for reproducibility; check [Gateway API releases](https://github.com/kubernetes-sigs/gateway-api/releases) for updates
- **No namespace** -- CRDs are cluster-scoped; no namespace is needed

## Placeholders to Replace

No placeholders -- this preset is directly deployable.
