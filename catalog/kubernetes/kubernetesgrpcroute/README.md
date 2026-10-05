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
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrpcRoute
metadata:
  name: greeter-route
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
  hostnames:
    - api.example.com
  rules:
    - matches:
        - method:
            service: helloworld.Greeter
      backendRefs:
        - name:
            value: greeter
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
| `parentRefs` | list | Gateways (and optional listener `sectionName`) the route attaches to. Each `name` is a reference (defaults to `KubernetesGateway`). |
| `hostnames` | list | Authority (Host) values that select this route. |
| `rules[].matches` | list | Method (service/method) and header matchers. |
| `rules[].filters` | list | Request/response header modify, request mirror, extension ref. |
| `rules[].backendRefs` | list | Weighted backends; each `name` is a reference (defaults to `KubernetesService`). |

## Examples

### Service/method routing

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
  hostnames:
    - api.example.com
  rules:
    - matches:
        - method:
            type: Exact
            service: helloworld.Greeter
            method: SayHello
      backendRefs:
        - name:
            value: greeter
          port: 9000
```

### Weighted canary split

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
  hostnames:
    - api.example.com
  rules:
    - backendRefs:
        - name:
            value: greeter-stable
          port: 9000
          weight: 90
        - name:
            value: greeter-canary
          port: 9000
          weight: 10
```

## Composing in Infra Charts

`KubernetesGrpcRoute` is a leaf in the ingress DAG: it attaches to a `Gateway`
and forwards to backend Services. Every neighbor reference is a foreign key --
`namespace`, `parentRefs[].name` (defaults to `KubernetesGateway`), and
`backendRefs[].name` (defaults to `KubernetesService`) -- so wiring them with
`valueFrom` creates real dependency edges and the platform orders the deployment
automatically:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrpcRoute
metadata:
  name: "{{ values.env }}-greeter-route"
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  parentRefs:
    - name:
        valueFrom:
          kind: KubernetesGateway
          name: "{{ values.env }}-gateway"
          fieldPath: status.outputs.gateway_name
  hostnames:
    - "api.{{ values.domain }}"
  rules:
    - matches:
        - method:
            service: "{{ values.grpc_service }}"
      backendRefs:
        - name:
            valueFrom:
              kind: KubernetesService
              name: "{{ values.service_name }}"
              fieldPath: status.outputs.service_name
          port: 9000
```

When a parent or backend is not Planton-managed, pass the literal name with
`value:` instead.

Full ingress stack DAG:

```
KubernetesCertManager -> KubernetesClusterIssuer -> KubernetesCertificate
   -> (Secret) -> KubernetesGateway -> KubernetesGrpcRoute / KubernetesHttpRoute
```

## Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created GRPCRoute (equals metadata.name). |
| `namespace` | Namespace the GRPCRoute was created in. |

## Related Kinds

- [Kubernetes Gateway](../kubernetesgateway)
- [Kubernetes Gateway Class](../kubernetesgatewayclass)
- [Kubernetes HTTP Route](../kuberneteshttproute)
- [Kubernetes Gateway API CRDs](../kubernetesgatewayapicrds)
- [Kubernetes Namespace](../kubernetesnamespace)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
