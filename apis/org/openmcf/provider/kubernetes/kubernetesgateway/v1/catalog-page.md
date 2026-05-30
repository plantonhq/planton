# Kubernetes Gateway

Provision a Kubernetes Gateway API `Gateway` -- a namespaced ingress entry point
that binds listeners (port, protocol, and optional TLS) to network addresses and
is programmed by the controller behind its GatewayClass. Routes attach to it to
define host, path, and backend routing.

## What Gets Created

- A namespaced `gateway.networking.k8s.io/v1` `Gateway` custom resource.
- One or more listeners (HTTP, HTTPS, TLS, TCP, or UDP), with optional
  per-listener TLS termination/passthrough and route-attachment policy.
- Optional requested addresses, infrastructure labels/annotations, and
  gateway-wide frontend/backend TLS configuration.

## Prerequisites

- Gateway API CRDs installed on the cluster (`KubernetesGatewayApiCrds`).
- A controller-backed `GatewayClass` (`KubernetesGatewayClass`), e.g. Istio or Envoy Gateway.
- The target namespace (`KubernetesNamespace`).
- For HTTPS listeners, a TLS Secret (e.g. from `KubernetesCertificate`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGateway
metadata:
  name: web-gateway
spec:
  namespace:
    value: istio-ingress
  gatewayClassName:
    value: istio
  listeners:
    - name: https
      hostname: app.example.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name: app-tls
```

```bash
openmcf apply -f gateway.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace to create the Gateway in. |
| `gatewayClassName` | reference | GatewayClass that selects the controller. |
| `listeners` | list | At least one listener (name, port, protocol). |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `addresses` | list | Requested IP/hostname addresses. |
| `infrastructure` | object | Labels, annotations, and a per-Gateway parametersRef. |
| `allowedListeners` | object | Which ListenerSets may attach. |
| `tls` | object | Gateway-wide frontend (mutual TLS) and backend client-cert config. |

## Examples

### HTTPS with TLS termination

```yaml
spec:
  namespace:
    value: istio-ingress
  gatewayClassName:
    value: istio
  listeners:
    - name: https
      hostname: app.example.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name: app-tls
```

### Multi-protocol (HTTP + HTTPS + TCP)

```yaml
spec:
  namespace:
    value: istio-ingress
  gatewayClassName:
    value: istio
  listeners:
    - name: http
      port: 80
      protocol: HTTP
    - name: https
      hostname: app.example.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name: app-tls
    - name: postgres
      port: 5432
      protocol: TCP
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `gatewayName` | Name of the created Gateway (target of Route parentRefs). |
| `namespace` | Namespace the Gateway was created in. |
| `gatewayClassName` | Name of the GatewayClass this Gateway belongs to. |

## Related Components

- [Kubernetes Gateway API CRDs](kubernetesgatewayapicrds)
- [Kubernetes Gateway Class](kubernetesgatewayclass)
- [Kubernetes Certificate](kubernetescertificate)
- [Kubernetes Namespace](kubernetesnamespace)
