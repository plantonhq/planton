# KubernetesGateway

> A Kubernetes Gateway API `Gateway`: a namespaced ingress entry point that binds Listeners (port + protocol + optional TLS) to addresses, programmed by the controller behind its GatewayClass.

## Overview

`KubernetesGateway` is a first-class Planton component that provisions an
upstream Gateway API `Gateway` resource at 100% fidelity with the standard
channel of Gateway API v1.5.1. It models listeners, per-listener TLS
termination/passthrough, requested addresses, infrastructure labels/annotations,
gateway-wide frontend (mutual TLS) and backend client-certificate
configuration, and route/listener-set attachment policy.

Unlike a raw `KubernetesManifest`, this component gives you proto validation,
foreign-key wiring (to `KubernetesNamespace` and `KubernetesGatewayClass`),
typed Pulumi and Terraform modules, and InfraChart composability.

## Prerequisites

- The Gateway API CRDs are installed on the target cluster
  (`KubernetesGatewayApiCrds`).
- A controller-backed `GatewayClass` exists (`KubernetesGatewayClass`) -- for
  example Istio or Envoy Gateway.
- The target namespace exists (`KubernetesNamespace`).
- For HTTPS listeners, a `kubernetes.io/tls` Secret with the certificate/key
  (commonly produced by a cert-manager `KubernetesCertificate`).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
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
planton pulumi up --manifest gateway.yaml
```

## How It Works

1. The Gateway is created in `spec.namespace` and references a `GatewayClass`
   via `spec.gatewayClassName`; the named controller programs it.
2. Each listener binds a port and protocol; HTTPS/TLS listeners carry TLS
   configuration. Listeners must be distinct by name and by
   port/protocol/hostname.
3. Routes (`HTTPRoute`, `GRPCRoute`, `TLSRoute`, `TCPRoute`) attach to the
   Gateway by referencing its name in their `parentRefs`, subject to each
   listener's `allowedRoutes` policy.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `target_cluster` | `KubernetesClusterSelector` | no | Target cluster selector. |
| `namespace` | `StringValueOrRef` -> `KubernetesNamespace` | yes | Namespace to create the Gateway in. |
| `gateway_class_name` | `StringValueOrRef` -> `KubernetesGatewayClass` | yes | GatewayClass that selects the controller. |
| `listeners` | `[]KubernetesGatewayListener` | yes (1-64) | Logical endpoints (name, port, protocol, tls, allowed_routes). |
| `addresses` | `[]KubernetesGatewayAddress` | no (max 16) | Requested addresses (IPAddress/Hostname). |
| `infrastructure` | `KubernetesGatewayInfrastructure` | no | Labels/annotations/parametersRef for created resources. |
| `allowed_listeners` | `KubernetesGatewayAllowedListeners` | no | Which ListenerSets may attach. |
| `tls` | `KubernetesGatewayTlsConfig` | no | Gateway-wide frontend mTLS and backend client-cert config. |

### Listener fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Unique listener name (Route `parentRef.sectionName` target). |
| `hostname` | string | no | Virtual host to match (HTTP/HTTPS/TLS); wildcard prefix allowed. |
| `port` | int32 (1-65535) | yes | Network port. |
| `protocol` | string | yes | `HTTP`, `HTTPS`, `TLS`, `TCP`, `UDP`, or a domain-prefixed custom protocol. |
| `tls.mode` | string | no | `Terminate` (default) or `Passthrough`. |
| `tls.certificate_refs` | `[]SecretObjectReference` | conditionally | TLS Secrets to terminate with (required for Terminate). |
| `allowed_routes` | `KubernetesGatewayAllowedRoutes` | no | Which Route kinds/namespaces may attach. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `gateway_name` | Name of the created Gateway (equals `metadata.name`); the target of Route `parentRefs`. |
| `namespace` | Namespace the Gateway was created in. |
| `gateway_class_name` | Name of the GatewayClass this Gateway belongs to. |

## Related Components

- [`KubernetesGatewayApiCrds`](../kubernetesgatewayapicrds/) -- installs the Gateway API CRDs (prerequisite).
- [`KubernetesGatewayClass`](../kubernetesgatewayclass/) -- defines the controller class this Gateway references.
- [`KubernetesCertificate`](../kubernetescertificate/) -- provisions the TLS Secret HTTPS listeners reference.
- [`KubernetesNamespace`](../kubernetesnamespace/) -- the namespace the Gateway is created in.
