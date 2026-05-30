# KubernetesHttpRoute

> A Kubernetes Gateway API HTTPRoute: namespaced HTTP routing rules that attach
> to a Gateway and forward matching requests to backend Services.

## Overview

`KubernetesHttpRoute` models the upstream Gateway API `HTTPRoute` at 100%
standard-channel fidelity, so any HTTP routing pattern the Gateway API supports
is expressible without dropping down to a raw `KubernetesManifest`. You get proto
validation, typed Pulumi and Terraform modules, and InfraChart composition.

A route matches requests by hostname, path, header, query parameter, or method;
optionally transforms them with filters (header modification, redirect, URL
rewrite, request mirror, CORS, or an implementation-specific extension); and
forwards them to one or more weighted backends.

## Prerequisites

- The Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`).
- The target namespace (`KubernetesNamespace`).
- The backend Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
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
openmcf pulumi up --manifest httproute.yaml
```

## How It Works

1. The route attaches to the Gateway(s) named in `parentRefs` (optionally to a
   specific listener via `sectionName`).
2. Incoming requests are matched against each rule's `matches` (a request
   matches a rule if any one match is satisfied; conditions within a match are
   ANDed).
3. Matching requests are optionally transformed by `filters`, then forwarded to
   `backendRefs`, split by `weight` when multiple backends are given.

## Configuration Reference

| Field | Required | Description |
|-------|----------|-------------|
| `namespace` | yes | Namespace the route is created in (FK to `KubernetesNamespace`). |
| `parent_refs` | no | Gateways (and optional listener `section_name`) the route attaches to. |
| `hostnames` | no | Host header values that select this route (wildcard prefix allowed). |
| `rules` | yes (≥1) | Match / filter / backend rules. |

### Rule fields

| Field | Description |
|-------|-------------|
| `name` | Optional unique rule name. |
| `matches` | Path / header / query-param / method matchers (max 64). |
| `filters` | Header modify, redirect, URL rewrite, request mirror, CORS, or extension ref (max 16). |
| `backend_refs` | Weighted backends to forward to (max 16). |
| `timeouts` | `request` and `backend_request` durations. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `route_name` | Name of the created HTTPRoute (equals `metadata.name`). |
| `namespace` | Namespace the HTTPRoute was created in. |

## Related Components

- [`KubernetesGateway`](../../kubernetesgateway/v1/README.md) -- the Gateway routes attach to.
- [`KubernetesGatewayClass`](../../kubernetesgatewayclass/v1/README.md) -- the controller class.
- [`KubernetesGatewayApiCrds`](../../kubernetesgatewayapicrds/v1/README.md) -- installs the CRDs.
- [`KubernetesNamespace`](../../kubernetesnamespace/v1/README.md) -- the target namespace.
