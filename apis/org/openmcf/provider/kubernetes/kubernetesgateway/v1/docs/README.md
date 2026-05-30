# KubernetesGateway -- Research Documentation

Comprehensive background on the Kubernetes Gateway API `Gateway` resource, the
upstream specification this component mirrors, and the design decisions behind
the OpenMCF `KubernetesGateway` component.

## Table of Contents

1. [Introduction](#introduction)
2. [The Gateway API Role-Oriented Model](#the-gateway-api-role-oriented-model)
3. [Where Gateway Sits](#where-gateway-sits)
4. [Anatomy of GatewaySpec](#anatomy-of-gatewayspec)
5. [Listener Distinctness and TLS Rules](#listener-distinctness-and-tls-rules)
6. [Why Gateway Is a First-Class OpenMCF Component](#why-gateway-is-a-first-class-openmcf-component)
7. [Design Decisions](#design-decisions)
8. [Controller Landscape](#controller-landscape)
9. [80/20 Scoping](#8020-scoping)
10. [Common Pitfalls](#common-pitfalls)
11. [Conclusion](#conclusion)
12. [References](#references)

## Introduction

The Kubernetes Gateway API is the successor to the Ingress API. It is a
standards-based, role-oriented, expressive specification for configuring L4/L7
traffic routing into a cluster, developed by SIG-Network and implemented by 20+
controllers (Istio, Envoy Gateway, NGINX, Cilium, Traefik, Kong, and the major
cloud load balancers). It reached GA (`v1`) in October 2023 and continues to
evolve through quarterly releases; this component tracks **v1.5.1**.

A `Gateway` is the central object: it represents an instance of traffic-handling
infrastructure (typically a load balancer or proxy deployment) and declares the
**listeners** -- the ports, protocols, and TLS settings -- on which it accepts
connections. Routes then attach to a Gateway to describe how matching traffic is
dispatched to backends.

## The Gateway API Role-Oriented Model

The Gateway API deliberately separates concerns across three personas, each
owning a distinct resource:

```
Infrastructure Provider          Cluster Operator              Application Developer
        │                              │                               │
        ▼                              ▼                               ▼
   GatewayClass  ───selects───►     Gateway      ◄───parentRef───   HTTPRoute / GRPCRoute
   (controller)                  (listeners,                        TLSRoute / TCPRoute
                                  addresses, TLS)                   (host/path/backends)
```

- **GatewayClass** (infrastructure provider): names the controller implementation.
- **Gateway** (cluster operator): owns the public entry point -- ports, protocols, TLS, and which routes/namespaces may attach.
- **Routes** (application developer): describe routing rules and attach to a Gateway via `parentRefs`.

`KubernetesGateway` is the OpenMCF representation of the middle layer. It pairs
with `KubernetesGatewayClass` (already shipped) and the route components
(`KubernetesHttpRoute`, `KubernetesGrpcRoute`, `KubernetesTlsRoute`,
`KubernetesTcpRoute`) being forged alongside it.

## Where Gateway Sits

```
KubernetesGatewayApiCrds        (install CRDs + controller)
        │
        ▼
KubernetesGatewayClass          (controller class)  ──gatewayClassName FK──┐
        │                                                                  │
KubernetesNamespace ──namespace FK──┐                                      │
                                    ▼                                      ▼
                              KubernetesGateway  (listeners, TLS, addresses)
                                    ▲          ▲
        KubernetesCertificate ──────┘          └────── HTTPRoute / GRPCRoute / TLSRoute / TCPRoute
        (TLS Secret for HTTPS                          (attach via parentRefs)
         listener certificateRefs)
```

## Anatomy of GatewaySpec

The upstream `GatewaySpec` (standard channel, v1.5.1) has the following fields,
all mirrored in `KubernetesGatewaySpec` after the OpenMCF envelope
(`target_cluster`, `namespace`):

| Upstream field | OpenMCF field | Notes |
|----------------|---------------|-------|
| `gatewayClassName` | `gateway_class_name` | Foreign key to `KubernetesGatewayClass`. |
| `listeners` | `listeners` | 1-64 listeners; the heart of the spec. |
| `addresses` | `addresses` | Up to 16 requested addresses (Extended support). |
| `infrastructure` | `infrastructure` | Labels/annotations/parametersRef for created resources. |
| `allowedListeners` | `allowed_listeners` | Which ListenerSets may attach. |
| `tls` | `tls` | Gateway-wide frontend (mutual TLS) and backend client-cert config. |
| `defaultScope` | *(excluded)* | Experimental; absent from the standard CRD. See Design Decisions. |

### Listener

A listener is the core unit: `name`, `hostname`, `port`, `protocol`, `tls`, and
`allowedRoutes`. The protocol determines which fields are meaningful:

- **HTTP** -- cleartext; no TLS, hostname-aware.
- **HTTPS** -- TLS terminated at the Gateway; requires a certificate.
- **TLS** -- TLS either terminated or passed through (for TLSRoute); requires `mode`.
- **TCP / UDP** -- connection/datagram forwarding; no hostname, no TLS.

### Per-listener TLS vs gateway-wide TLS

There are two distinct TLS concepts, and the component models both:

- **Listener `tls`** (`ListenerTLSConfig`): how an individual HTTPS/TLS listener
  terminates or passes through TLS, including the certificate Secrets it serves.
- **Gateway `tls`** (`GatewayTLSConfig`): gateway-wide **frontend** client
  certificate validation (mutual TLS for inbound connections, with optional
  per-port overrides) and **backend** client-certificate material the Gateway
  presents when connecting to upstreams.

## Listener Distinctness and TLS Rules

The upstream spec enforces several cross-cutting rules via CEL `XValidation`.
`KubernetesGateway` translates each one faithfully into `buf.validate` so the
same errors surface at author time in OpenMCF UIs and CLIs:

- Each listener `name` is unique within the Gateway.
- The combination of `port`, `protocol`, and `hostname` is unique across listeners.
- `tls` must not be set when `protocol` is HTTP, TCP, or UDP.
- An HTTPS listener may only `Terminate` (mode unset or `Terminate`).
- A TLS listener must declare its `tls.mode`.
- `hostname` must not be set for TCP/UDP listeners.
- A `Terminate` listener must provide `certificateRefs` or `options`.
- Requested IPAddress and Hostname address values are each unique.
- Per-port frontend TLS overrides target unique ports.

## Why Gateway Is a First-Class OpenMCF Component

Without this component, customers wanting Gateway API ingress are forced to use
`KubernetesManifest` (raw YAML), which sacrifices:

1. **Proto validation** -- the distinctness and TLS rules above are enforced before apply.
2. **Foreign-key wiring** -- `namespace` and `gateway_class_name` reference other OpenMCF resources, enabling InfraChart DAG ordering and Planton UI resource pickers.
3. **Typed IaC** -- both the Pulumi and Terraform modules construct the CRD from typed inputs, catching structural errors at compile/plan time.
4. **Composability** -- the Gateway's outputs (`gateway_name`, `namespace`) feed Route components that attach to it.

## Design Decisions

- **100% standard-channel fidelity.** Every field of the standard-channel
  `GatewaySpec` is represented. The only omission is the experimental
  `defaultScope` field, which is absent from the standard CRD and from the typed
  Pulumi resource OpenMCF provisions with; including it would have no
  deployable target.
- **Value fields are strings, validated per upstream kind.** Gateway API uses a
  mix of open patterns and closed enums. Open-set values (`protocol`, address
  `type`) are validated with the upstream **regex**, preserving custom
  domain-prefixed values. Closed enums (`tls.mode`, namespace `from`, frontend
  validation `mode`, selector `operator`) are validated with CEL membership
  checks. Both keep exact upstream casing, so no case-mapping is needed in the
  IaC layer and CEL rules translate verbatim. (This supersedes an earlier
  project decision that assumed proto enums; recorded in the project's design
  decisions.)
- **No baked-in OpenMCF defaults for upstream fields.** Upstream kubebuilder
  defaults (e.g. `tls.mode=Terminate`, address `type=IPAddress`, route
  `from=Same`) are controller/CRD-enforced and are documented in field comments
  rather than set as OpenMCF defaults, so controller behavior is never
  second-guessed.
- **Certificate references stay plain.** `listeners[].tls.certificate_refs`
  reuses the shared structured `SecretObjectReference` (group/kind/name/namespace)
  rather than a `StringValueOrRef` foreign key. Wrapping an array of multi-field
  references would distort the upstream structure; a typed FK to
  `KubernetesCertificate` is being evaluated as a consistent, family-wide change.
- **Typed crd2pulumi resources.** The Pulumi module uses `gatewayv1.NewGateway`,
  not an untyped `CustomResource`.

## Controller Landscape

`gatewayClassName` ultimately resolves to a controller. Common choices:

| Controller | Typical GatewayClass controllerName |
|------------|-------------------------------------|
| Istio | `istio.io/gateway-controller` |
| Envoy Gateway | `gateway.envoyproxy.io/gatewayclass-controller` |
| NGINX Gateway Fabric | `gateway.nginx.org/nginx-gateway-controller` |
| Cilium | `io.cilium/gateway-controller` |

The Gateway itself is controller-agnostic; the GatewayClass selects the
implementation, and implementation-specific tuning flows through
`infrastructure.parametersRef` or listener `tls.options`.

## 80/20 Scoping

- **In scope:** the full standard-channel `GatewaySpec` -- listeners (all five
  core protocols), per-listener TLS, allowed routes, requested addresses,
  infrastructure metadata, gateway-wide frontend/backend TLS, and allowed
  listeners.
- **Out of scope:** the experimental `defaultScope` field; the `ListenerSet`
  resource itself (the `allowed_listeners` field is retained for fidelity and
  forward compatibility); CRD `status` (reconciled asynchronously by the
  controller and observed via kubectl, not stored in stack outputs).

## Common Pitfalls

- **Setting `tls` on an HTTP/TCP/UDP listener.** TLS only applies to HTTPS/TLS
  listeners; the spec rejects it otherwise.
- **Forgetting a certificate on a Terminate listener.** A terminating listener
  must reference at least one TLS Secret (or supply implementation `options`).
- **Duplicate listeners.** Two listeners may not share the same name, nor the
  same port/protocol/hostname combination.
- **Expecting status in outputs.** Assigned addresses and listener conditions
  are controller-managed; query them with `kubectl get gateway`. Stack outputs
  expose only the stable identifiers (`gateway_name`, `namespace`,
  `gateway_class_name`).
- **Missing prerequisites.** The Gateway will not program until the Gateway API
  CRDs and a controller-backed GatewayClass are present.

## Conclusion

`KubernetesGateway` brings the central Gateway API object into OpenMCF at full
standard-channel fidelity, with validation, typed IaC, and foreign-key
composability. Together with `KubernetesGatewayClass` and the route components,
it completes a declarative, controller-agnostic ingress layer that customers can
compose in InfraCharts without dropping to raw YAML.

## References

- [Gateway API: Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/)
- [Gateway API: TLS configuration](https://gateway-api.sigs.k8s.io/guides/tls/)
- [Gateway API v1.5.1 specification](https://github.com/kubernetes-sigs/gateway-api/tree/v1.5.1)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

## Composing in Infra Charts

`KubernetesGateway` is the mid-tier hub of the ingress DAG: it depends on a
GatewayClass and (for TLS) a certificate Secret, and routes attach to it. Two
mechanisms wire it to its neighbors (see project decision DD-009):

1. **Data dependencies use `valueFrom`.** `namespace` (-> `KubernetesNamespace`)
   and `gateway_class_name`
   (-> `KubernetesGatewayClass.status.outputs.gateway_class_name`) are
   `StringValueOrRef` fields, so the platform builds those DAG edges
   automatically.
2. **Topology dependencies use `metadata.relationships`.**
   `listeners[].tls.certificate_refs` is a **plain** reference (an array of
   multi-field upstream objects), not a foreign key -- a plain Secret name creates
   no automatic DAG edge. Express the Gateway -> Certificate dependency explicitly
   so the certificate is provisioned first:

```yaml
metadata:
  name: "{{ values.env }}-gateway"
  relationships:
    - kind: KubernetesCertificate
      name: "{{ values.domain }}-cert"
      type: uses
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  gateway_class_name:
    valueFrom:
      kind: KubernetesGatewayClass
      name: "{{ values.env }}-gateway-class"
      fieldPath: status.outputs.gateway_class_name
  listeners:
    - name: https
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificate_refs:
          # literal Secret name, typically KubernetesCertificate.status.outputs.secret_name
          - name: "{{ values.domain }}-cert"
```

Full ingress stack:
`CertManager -> ClusterIssuer -> Certificate -> (Secret) -> Gateway -> HTTPRoute / GRPCRoute`.
Data edges use `valueFrom`; the plain `certificate_refs` edge uses
`metadata.relationships`.
