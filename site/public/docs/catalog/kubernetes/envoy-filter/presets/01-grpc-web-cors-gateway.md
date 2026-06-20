---
title: "Add a CORS HTTP filter to a gateway (gRPC-Web)"
description: "The canonical \"escape hatch\" use: insert Envoy's native CORS HTTP filter into a gateway's HTTP connection manager so a browser can call a gRPC-Web (or any cross-origin) backend through the gateway...."
type: "preset"
rank: "01"
presetSlug: "01-grpc-web-cors-gateway"
componentSlug: "envoy-filter"
componentTitle: "Envoy Filter"
provider: "kubernetes"
icon: "package"
order: 1
---

# Add a CORS HTTP filter to a gateway (gRPC-Web)

The canonical "escape hatch" use: insert Envoy's native CORS HTTP filter into a gateway's HTTP
connection manager so a browser can call a gRPC-Web (or any cross-origin) backend through the
gateway. Native `HTTPRoute` CORS only lands in Istio 1.26+, so on many meshes this EnvoyFilter
is still how CORS is delivered at the edge.

## When to Use

- You expose a gRPC-Web or browser-facing API through an Istio gateway and the browser's
  preflight (`OPTIONS`) requests need CORS handling that no first-class API yet provides on your
  Istio version.
- You need to insert a stock Envoy HTTP filter (CORS, ext_authz, rate limiting) at the gateway
  before a typed Istio API exists for it.

## Key Configuration Choices

- **`target_refs` -> the Gateway** -- attaches the patch precisely to one gateway's proxies
  (preferred over a broad `workload_selector` for gateways; waypoints require target_refs).
- **`apply_to: HTTP_FILTER` + `context: GATEWAY`** -- patches the HTTP filter chain of the
  gateway's listeners.
- **`sub_filter.name: envoy.filters.http.router`** -- anchors the insertion at the terminal
  router filter so `INSERT_BEFORE` places CORS just ahead of it (CORS must run before routing).
- **`patch.value`** -- the free-form Envoy filter config. Fill in the CORS policy your app needs
  (allowed origins/methods/headers) under `typed_config`; the value shown registers the filter.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the gateway exists (`KubernetesIstio`, plus the target Gateway).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace the gateway and this EnvoyFilter live in (e.g. `edge`). |
| `<gateway-name>` | Name of the Gateway to attach the CORS filter to. |

This is an expert-only escape hatch. Prefer a first-class typed API (native `HTTPRoute` CORS on
Istio 1.26+) where your version supports it, and graduate off this EnvoyFilter when you can.
