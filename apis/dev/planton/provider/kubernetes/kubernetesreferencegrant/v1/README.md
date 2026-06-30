# Kubernetes Reference Grant

Provision a Kubernetes Gateway API `ReferenceGrant` -- a namespaced authorization
that lets resources in OTHER namespaces reference specified kinds of resources in
THIS grant's namespace. Almost every cross-namespace reference in the Gateway API
(a Gateway's TLS `certificateRefs` into a cert namespace, a Route's `backendRefs`
into a backend namespace) requires a ReferenceGrant in the referenced namespace.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `ReferenceGrant` custom resource.
- A `from` list (trusted source namespaces + kinds) and a `to` list (referenceable
  kinds, optionally narrowed to a specific name) in this grant's namespace.

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- The target namespace (`KubernetesNamespace`) -- the "to" namespace whose
  resources the grant authorizes inbound references to.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesReferenceGrant
metadata:
  name: allow-frontend-to-backend
spec:
  namespace:
    value: backend-ns
  from:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      namespace: frontend-ns
  to:
    - group: ""
      kind: Service
```

```bash
planton apply -f referencegrant.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | The "to" namespace the grant is created in. |
| `from` | list | One to 16 trusted (group, kind, namespace) sources. |
| `to` | list | One to 16 referenceable (group, kind, optional name) targets. |

### `from[]` entry

| Field | Type | Description |
|-------|------|-------------|
| `group` | string | API group of the source kind. Empty (`""`) means the core group. |
| `kind` | string | Source kind, e.g. `HTTPRoute`, `Gateway`. Required. |
| `namespace` | string | Source namespace trusted to reference in. Required. |

### `to[]` entry

| Field | Type | Description |
|-------|------|-------------|
| `group` | string | API group of the target kind. Empty (`""`) means the core group. |
| `kind` | string | Target kind, e.g. `Service`, `Secret`. Required. |
| `name` | string | Optional. When set, narrows the grant to one named resource; when omitted, covers all resources of the group/kind. |

## Examples

### Allow a Gateway to reference TLS Secrets in another namespace

```yaml
spec:
  namespace:
    value: cert-ns
  from:
    - group: gateway.networking.k8s.io
      kind: Gateway
      namespace: istio-ingress
  to:
    - group: ""
      kind: Secret
```

### Allow Routes to reference backend Services in another namespace

```yaml
spec:
  namespace:
    value: backend-ns
  from:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      namespace: frontend-ns
    - group: gateway.networking.k8s.io
      kind: GRPCRoute
      namespace: frontend-ns
  to:
    - group: ""
      kind: Service
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace`
output via `valueFrom`. The `from`/`to` entries are trust assertions about kinds,
not foreign keys; the one genuine cross-resource reference is `from[].namespace`,
which (when Planton-managed) is wired via `metadata.relationships`. The grant
itself is a low-dependency leaf -- the consuming Gateway/Route is what must order
itself after the grant. See `docs/README.md` for the full pattern (DD-009).

## Stack Outputs

| Output | Description |
|--------|-------------|
| `referenceGrantName` | Name of the created ReferenceGrant (equals metadata.name). |
| `namespace` | Namespace the ReferenceGrant was created in. |

## Related Components

- [Kubernetes Gateway](kubernetesgateway)
- [Kubernetes HTTP Route](kuberneteshttproute)
- [Kubernetes GRPC Route](kubernetesgrpcroute)
- [Kubernetes TLS Route](kubernetestlsroute)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
