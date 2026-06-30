# Kubernetes TCP Route

Provision a Kubernetes Gateway API `TCPRoute` -- namespaced rules that forward raw
TCP connections arriving on a Gateway listener to backend Services. A TCP route
has no matching: connections on the listener's port are forwarded to the rule's
backends. Use it to expose non-HTTP TCP services (databases, brokers, custom
protocols) through a Gateway.

> **Experimental channel.** TCPRoute is served as
> `gateway.networking.k8s.io/v1alpha2` and requires the Gateway API experimental
> CRDs (`KubernetesGatewayApiCrds` with `install_channel: experimental`).

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1alpha2` `TCPRoute` custom resource.
- One or more rules (max 16), each forwarding to one or more weighted backend
  refs.

## Prerequisites

- Gateway API experimental-channel CRDs installed (`KubernetesGatewayApiCrds`
  with `install_channel: experimental`).
- A `Gateway` to attach to via `parentRefs` (`KubernetesGateway`) with a `TCP`
  listener.
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
| `rules[].name` | string | Optional rule name. |
| `rules[].backendRefs` | list | Weighted backends to forward to. |

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

## Stack Outputs

| Output | Description |
|--------|-------------|
| `routeName` | Name of the created TCPRoute (equals metadata.name). |
| `namespace` | Namespace the TCPRoute was created in. |

## Related Components

- [Kubernetes Gateway](kubernetesgateway)
- [Kubernetes TLS Route](kubernetestlsroute)
- [Kubernetes Gateway Class](kubernetesgatewayclass)
- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Namespace](kubernetesnamespace)
