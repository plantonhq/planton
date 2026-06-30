# Kubernetes TCP Route

Provision a Kubernetes Gateway API `TCPRoute` -- namespaced rules that forward raw
TCP connections arriving on a Gateway listener to backend Services. A TCP route
has no matching: connections on the listener's port are forwarded to the rule's
backends. Use it to expose non-HTTP TCP services (databases, brokers, custom
protocols) through a Gateway.

> **Experimental channel.** TCPRoute is part of the Gateway API **experimental**
> channel (served as `gateway.networking.k8s.io/v1alpha2`), not the standard
> channel. Install the prerequisite CRDs with
> `KubernetesGatewayApiCrds` `install_channel: experimental`, and ensure your
> Gateway controller supports TCPRoute. Its sibling `KubernetesTlsRoute`, by
> contrast, is standard-channel `v1`.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1alpha2` `TCPRoute` custom resource.
- One or more rules (max 16), each forwarding to one or more weighted backend
  refs.

## Prerequisites

- Gateway API **experimental-channel** CRDs installed on the cluster
  (`KubernetesGatewayApiCrds` with `install_channel: experimental`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`) with a listener
  of protocol `TCP`.
- The target namespace (`KubernetesNamespace`).
- The backend Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTcpRoute
metadata:
  name: postgres-route
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
      sectionName: tcp
  rules:
    - backendRefs:
        - name: postgres
          port: 5432
```

```bash
planton apply -f tcproute.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace to create the route in. |
| `rules` | list | One to 16 routing rules. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `parentRefs` | list | Gateways (and optional listener `sectionName`) the route attaches to. |
| `useDefaultGateways` | string | `All` or `None` -- experimental default-Gateway attachment. |
| `rules[].name` | string | Optional rule name (unique within the route). |
| `rules[].backendRefs` | list | One to 16 weighted backends to forward to. |

## Examples

### Port forwarding

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
      sectionName: tcp
  rules:
    - backendRefs:
        - name: postgres
          port: 5432
```

### Weighted backends (canary)

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name: my-gateway
      sectionName: tcp
  rules:
    - backendRefs:
        - name: broker-stable
          port: 9092
          weight: 90
        - name: broker-canary
          port: 9092
          weight: 10
```

## Composing in Infra Charts

`KubernetesTcpRoute` is a leaf in the ingress DAG: it attaches to a `Gateway` and
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
kind: KubernetesTcpRoute
metadata:
  name: "{{ values.env }}-postgres-route"
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
      sectionName: tcp
  rules:
    - backendRefs:
        - name: "{{ values.service_name }}"
          port: 5432
```

Full ingress stack DAG (mixing `valueFrom` data edges and `metadata.relationships`
topology edges):

```
KubernetesGatewayApiCrds (experimental) -> KubernetesGateway -> KubernetesTcpRoute
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created TCPRoute (equals metadata.name). |
| `namespace` | Namespace the TCPRoute was created in. |

## Related Components

- [Kubernetes Gateway](kubernetesgateway)
- [Kubernetes Gateway Class](kubernetesgatewayclass)
- [Kubernetes TLS Route](kubernetestlsroute)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
