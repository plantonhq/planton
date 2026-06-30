# Kubernetes TLS Route

Provision a Kubernetes Gateway API `TLSRoute` -- namespaced TLS passthrough rules
that attach to a Gateway and forward connections, by SNI hostname, to backend
Services. The backend (not the Gateway) terminates TLS, so the encrypted stream
is forwarded end to end.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `TLSRoute` custom resource.
- Exactly one rule (the upstream maximum for a TLSRoute) that forwards to one or
  more weighted backend refs.

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`) with a listener
  of protocol `TLS` (typically `tls.mode: Passthrough`).
- The target namespace (`KubernetesNamespace`).
- The backend Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTlsRoute
metadata:
  name: secure-route
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
      sectionName: tls
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name: secure-backend
          port: 8443
```

```bash
planton apply -f tlsroute.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace to create the route in. |
| `hostnames` | list | One to 16 SNI hostnames that select this route (no IPs). |
| `rules` | list | Exactly one routing rule. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `parentRefs` | list | Gateways (and optional listener `sectionName`) the route attaches to. |
| `rules[].name` | string | Optional rule name (unique within the route). |
| `rules[].backendRefs` | list | One to 16 weighted backends to forward to. |

## Examples

### TLS passthrough by SNI

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
      sectionName: tls
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name: secure-backend
          port: 8443
```

### Weighted backends (canary)

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
      sectionName: tls
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name: secure-stable
          port: 8443
          weight: 90
        - name: secure-canary
          port: 8443
          weight: 10
```

## Composing in Infra Charts

`KubernetesTlsRoute` is a leaf in the ingress DAG: it attaches to a `Gateway` and
forwards to backend Services. Two mechanisms wire it to its neighbors (see DD-009
in the project knowledge base):

1. **Data dependencies use `valueFrom`.** `namespace` is a `StringValueOrRef`, so
   it can reference a `KubernetesNamespace` output and the platform builds the DAG
   edge automatically.
2. **Topology dependencies use `metadata.relationships`.** `parentRefs` and
   `backendRefs` are **plain** upstream references (arrays of multi-field objects),
   not foreign keys -- a plain name creates no automatic DAG edge. Express those
   edges explicitly so the chart deploys in the right order:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTlsRoute
metadata:
  name: "{{ values.env }}-secure-route"
  relationships:
    - kind: KubernetesGateway
      name: "{{ values.env }}-gateway"
      type: depends_on
    - kind: KubernetesService          # if the backend is Planton-managed
      name: "{{ values.service_name }}"
      type: uses
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  parentRefs:
    - name: "{{ values.env }}-gateway"   # literal Gateway name
      sectionName: tls
  hostnames:
    - "secure.{{ values.domain }}"
  rules:
    - backendRefs:
        - name: "{{ values.service_name }}"
          port: 8443
```

Full ingress stack DAG (mixing `valueFrom` data edges and `metadata.relationships`
topology edges):

```
KubernetesCertManager -> KubernetesClusterIssuer -> KubernetesCertificate
   -> (Secret) -> KubernetesGateway -> KubernetesTlsRoute / KubernetesHttpRoute
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created TLSRoute (equals metadata.name). |
| `namespace` | Namespace the TLSRoute was created in. |

## Related Components

- [Kubernetes Gateway](kubernetesgateway)
- [Kubernetes Gateway Class](kubernetesgatewayclass)
- [Kubernetes HTTP Route](kuberneteshttproute)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
