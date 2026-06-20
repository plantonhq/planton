---
title: "gRPC Service Routing"
description: "The most common GRPCRoute: match a public hostname and a gRPC service (and optionally a method), then forward to a backend gRPC Service. This is the standard pattern for exposing a gRPC API behind a..."
type: "preset"
rank: "01"
presetSlug: "01-grpc-service-routing"
componentSlug: "grpc-route"
componentTitle: "gRPC Route"
provider: "kubernetes"
icon: "package"
order: 1
---

# gRPC Service Routing

The most common GRPCRoute: match a public hostname and a gRPC service (and
optionally a method), then forward to a backend gRPC Service. This is the
standard pattern for exposing a gRPC API behind a Gateway.

## When to Use

- You expose a gRPC service at a hostname.
- You want all calls for a service (or a specific service/method) routed to one
  backend.
- You are wiring a gRPC app behind an existing Gateway (Istio, Envoy Gateway, ...).

## Key Configuration Choices

- **`parentRefs`** -- attaches the route to the Gateway by name; add `sectionName` to target a specific listener.
- **`hostnames`** -- the authority (Host) values that select this route; a leading `*.` is a suffix match.
- **`method.service`** -- the fully-qualified gRPC service (for example `helloworld.Greeter`); add `method` to match a single RPC, or omit `method` entirely to match all services.
- **`backendRefs[].port`** -- required when the backend is a core Service.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- The `Gateway` referenced in `parentRefs` exists (`KubernetesGateway`) and its
  listener accepts HTTP/2 (h2c over `HTTP`, or HTTP/2 over `HTTPS`).
- The target namespace exists (`KubernetesNamespace`).
- The backend gRPC Service exists in the route's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<gateway-name>` | Name of the `KubernetesGateway` this route attaches to. |
| `<api-hostname>` | Public hostname this route serves, e.g. `api.example.com`. |
| `<grpc-service>` | Fully-qualified gRPC service, e.g. `helloworld.Greeter`. |
| `<service-name>` | Name of the backend Kubernetes Service. |

Set `spec.namespace.value` to your namespace, or replace it with a `valueFrom`
reference to a `KubernetesNamespace`.
