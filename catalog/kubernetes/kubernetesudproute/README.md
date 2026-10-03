# Kubernetes UDP Route

Provision a Kubernetes Gateway API `UDPRoute` -- namespaced rules that forward
UDP datagrams arriving on a Gateway listener to backend Services. A UDP route
has no matching: traffic on the listener's port is forwarded to the rule's
backends. Use it to expose datagram workloads (DNS servers, syslog collectors,
game servers, and other UDP protocols) through a Gateway.

UDPRoute is a GA, standard-channel resource served as
`gateway.networking.k8s.io/v1` (Gateway API v1.6.1). The default standard-channel
install of `KubernetesGatewayApiCrds` includes it; no experimental channel is
needed.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `UDPRoute` custom resource.
- One or more rules (max 16), each forwarding to one or more weighted backend
  refs.

## Prerequisites

- Gateway API standard-channel CRDs installed (`KubernetesGatewayApiCrds`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`) with a `UDP`
  listener, backed by a controller that supports layer-4 routing.
- The target namespace (`KubernetesNamespace`).
- The backend Services the route forwards to.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesUdpRoute
metadata:
  name: dns-route
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
      sectionName: dns
  rules:
    - backendRefs:
        - name:
            value: coredns
          port: 53
```

```bash
planton apply -f udproute.yaml
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
| `parentRefs` | list | Gateways (and optional listener `sectionName`) the route attaches to. Each `name` is a reference (defaults to `KubernetesGateway`). |
| `rules[].name` | string | Optional rule name (unique within the route). |
| `rules[].backendRefs` | list | One to 16 weighted backends; each `name` is a reference (defaults to `KubernetesService`). |

## Examples

### DNS forwarding

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
      sectionName: dns
  rules:
    - backendRefs:
        - name:
            value: coredns
          port: 53
```

### Weighted backends (canary)

```yaml
spec:
  namespace:
    value: app-ns
  parentRefs:
    - name:
        value: my-gateway
      sectionName: dns
  rules:
    - name: dns-forward
      backendRefs:
        - name:
            value: dns-backend
          port: 53
          weight: 90
        - name:
            value: dns-backend-canary
          port: 53
          weight: 10
```

## Composing in Infra Charts

`KubernetesUdpRoute` is a leaf in the ingress DAG: it attaches to a `Gateway` and
forwards to backend Services. Every one of those neighbor references is a foreign
key -- `namespace`, `parentRefs[].name` (defaults to `KubernetesGateway`), and
`backendRefs[].name` (defaults to `KubernetesService`) -- so wiring them with
`valueFrom` creates real dependency edges and the platform orders the deployment
automatically. No manual relationship declarations are needed:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesUdpRoute
metadata:
  name: "{{ values.env }}-dns-route"
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
      sectionName: dns
  rules:
    - backendRefs:
        - name:
            valueFrom:
              kind: KubernetesService
              name: "{{ values.service_name }}"
              fieldPath: status.outputs.service_name
          port: 53
```

When a parent or backend is not Planton-managed (an externally created Gateway,
a custom backend kind), pass the literal name with `value:` instead.

Full ingress stack DAG:

```
KubernetesGatewayApiCrds -> KubernetesGateway -> KubernetesUdpRoute
```

## Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created UDPRoute (equals metadata.name). |
| `namespace` | Namespace the UDPRoute was created in. |

## Related Kinds

- [Kubernetes Gateway](../kubernetesgateway)
- [Kubernetes Gateway Class](../kubernetesgatewayclass)
- [Kubernetes TCP Route](../kubernetestcproute)
- [Kubernetes TLS Route](../kubernetestlsroute)
- [Kubernetes Gateway API CRDs](../kubernetesgatewayapicrds)
- [Kubernetes Namespace](../kubernetesnamespace)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
