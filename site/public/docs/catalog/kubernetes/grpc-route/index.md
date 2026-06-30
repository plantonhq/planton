---
title: "gRPC Route"
description: "gRPC Route deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesgrpcroute"
---

# Kubernetes gRPC Route

Provision a Kubernetes Gateway API `GRPCRoute` -- namespaced gRPC routing rules
that attach to a Gateway and forward matching requests to backend Services.
Match by hostname, gRPC service/method, or header; transform with filters; and
split traffic across weighted backends.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `GRPCRoute` custom resource.
- One or more rules, each with matches, optional filters, and backend refs.
- Optional per-rule and per-backend filters (header modification, request
  mirror, or an implementation-specific extension).

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`) whose listener
  accepts HTTP/2 (gRPC requires HTTP/2; over `HTTP` this is h2c).
- The target namespace (`KubernetesNamespace`).
- The backend gRPC Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGrpcRoute
metadata:
  name: greeter-route
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
  hostnames:
    - api.example.com
  rules:
    - matches:
        - method:
            service: helloworld.Greeter
      backendRefs:
        - name: greeter
          port: 9000
```

```bash
planton apply -f grpcroute.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace to create the route in. |
| `rules` | list | At least one routing rule. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `parentRefs` | list | Gateways (and optional listener `sectionName`) the route attaches to. |
| `hostnames` | list | Authority (Host) values that select this route. |
| `rules[].matches` | list | Method (service/method) and header matchers. |
| `rules[].filters` | list | Header modify, request mirror, extension ref. |
| `rules[].backendRefs` | list | Weighted backends to forward to. |

## Examples

### Service/method routing

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
  hostnames:
    - api.example.com
  rules:
    - matches:
        - method:
            type: Exact
            service: helloworld.Greeter
            method: SayHello
      backendRefs:
        - name: greeter
          port: 9000
```

### Weighted canary split

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
  hostnames:
    - api.example.com
  rules:
    - backendRefs:
        - name: greeter-stable
          port: 9000
          weight: 90
        - name: greeter-canary
          port: 9000
          weight: 10
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created GRPCRoute (equals metadata.name). |
| `namespace` | Namespace the GRPCRoute was created in. |

## Related Components

- [Kubernetes Gateway](kubernetesgateway)
- [Kubernetes HTTP Route](kuberneteshttproute)
- [Kubernetes Gateway Class](kubernetesgatewayclass)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
