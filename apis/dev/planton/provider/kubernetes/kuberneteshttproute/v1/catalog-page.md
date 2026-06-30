# Kubernetes HTTP Route

Provision a Kubernetes Gateway API `HTTPRoute` -- namespaced HTTP routing rules
that attach to a Gateway and forward matching requests to backend Services.
Match by hostname, path, header, query parameter, or method; transform with
filters; and split traffic across weighted backends.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `HTTPRoute` custom resource.
- One or more rules, each with matches, optional filters, and backend refs.
- Optional per-rule and per-backend filters (header modification, redirect, URL
  rewrite, request mirror, CORS, or an implementation-specific extension).

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`).
- The target namespace (`KubernetesNamespace`).
- The backend Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHttpRoute
metadata:
  name: web-route
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
  hostnames:
    - app.example.com
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: web
          port: 8080
```

```bash
planton apply -f httproute.yaml
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
| `hostnames` | list | Host header values that select this route. |
| `rules[].matches` | list | Path, header, query-param, and method matchers. |
| `rules[].filters` | list | Header modify, redirect, URL rewrite, request mirror, CORS, extension ref. |
| `rules[].backendRefs` | list | Weighted backends to forward to. |
| `rules[].timeouts` | object | Request and backend-request timeouts. |

## Examples

### Host + path routing

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
  hostnames:
    - app.example.com
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /api
      backendRefs:
        - name: api
          port: 8080
```

### Weighted canary split

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
  hostnames:
    - app.example.com
  rules:
    - backendRefs:
        - name: web-stable
          port: 8080
          weight: 90
        - name: web-canary
          port: 8080
          weight: 10
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created HTTPRoute (equals metadata.name). |
| `namespace` | Namespace the HTTPRoute was created in. |

## Related Components

- [Kubernetes Gateway](kubernetesgateway)
- [Kubernetes Gateway Class](kubernetesgatewayclass)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
