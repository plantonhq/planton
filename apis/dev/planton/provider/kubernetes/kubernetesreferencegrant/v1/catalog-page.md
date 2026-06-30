# Kubernetes Reference Grant

Provision a Kubernetes Gateway API `ReferenceGrant` -- the cross-namespace trust
primitive that authorizes resources in other namespaces to reference specified
kinds of resources in this grant's namespace. Required for any Gateway API topology
that spans namespaces (Gateways referencing cert Secrets, Routes referencing
backend Services across namespace boundaries).

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `ReferenceGrant` custom resource.
- A `from` list (trusted source namespaces + kinds) and a `to` list (referenceable
  kinds, optionally a specific name) scoped to this grant's namespace.

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- The target ("to") namespace (`KubernetesNamespace`).

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

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `from[].group` / `to[].group` | string | API group; empty (`""`) means the core group. |
| `to[].name` | string | Narrows the grant to a single named target; omit to cover all of the group/kind. |

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
- [Kubernetes TCP Route](kubernetestcproute)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
