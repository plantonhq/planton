# Kubernetes gRPC Route

Provision a Kubernetes Gateway API `GRPCRoute` -- namespaced gRPC routing rules
that attach to a Gateway and forward matching requests to backend Services.
Match by hostname, gRPC service/method, or header; transform with filters; and
split traffic across weighted backends.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `GRPCRoute` custom resource.
- One or more rules, each with matches, optional filters, and backend refs.
- Optional per-rule and per-backend filters (request/response header
  modification, request mirror, or an implementation-specific extension).

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`). Its listener
  should accept HTTP/2 (gRPC requires HTTP/2; over `HTTP` this is h2c).
- The target namespace (`KubernetesNamespace`).
- The backend gRPC Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
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
openmcf apply -f grpcroute.yaml
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
| `rules[].filters` | list | Request/response header modify, request mirror, extension ref. |
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

## Composing in Infra Charts

`KubernetesGrpcRoute` is a leaf in the ingress DAG: it attaches to a `Gateway`
and forwards to backend Services. Two mechanisms wire it to its neighbors (see
DD-009 in the project knowledge base):

1. **Data dependencies use `valueFrom`.** `namespace` is a `StringValueOrRef`, so
   it can reference a `KubernetesNamespace` output and the platform builds the DAG
   edge automatically.
2. **Topology dependencies use `metadata.relationships`.** `parentRefs` and
   `backendRefs` are **plain** upstream references (arrays of multi-field objects),
   not foreign keys -- a plain name creates no automatic DAG edge. Express those
   edges explicitly so the chart deploys in the right order:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrpcRoute
metadata:
  name: "{{ values.env }}-greeter-route"
  relationships:
    - kind: KubernetesGateway
      name: "{{ values.env }}-gateway"
      type: depends_on
    - kind: KubernetesService          # if the backend is OpenMCF-managed
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
  hostnames:
    - "api.{{ values.domain }}"
  rules:
    - matches:
        - method:
            service: "{{ values.grpc_service }}"
      backendRefs:
        - name: "{{ values.service_name }}"
          port: 9000
```

Full ingress stack DAG (mixing `valueFrom` data edges and `metadata.relationships`
topology edges):

```
KubernetesCertManager -> KubernetesClusterIssuer -> KubernetesCertificate
   -> (Secret) -> KubernetesGateway -> KubernetesGrpcRoute / KubernetesHttpRoute
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created GRPCRoute (equals metadata.name). |
| `namespace` | Namespace the GRPCRoute was created in. |

## Related Components

- [Kubernetes Gateway](kubernetesgateway)
- [Kubernetes Gateway Class](kubernetesgatewayclass)
- [Kubernetes HTTP Route](kuberneteshttproute)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
