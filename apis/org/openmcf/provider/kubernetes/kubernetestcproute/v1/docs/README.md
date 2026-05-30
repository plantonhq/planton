# KubernetesTcpRoute -- Research Documentation

This document explains why `KubernetesTcpRoute` exists, how it maps the upstream
Gateway API `TCPRoute` into the OpenMCF component model, and the design decisions
behind its spec, validation, and IaC. It is the source-of-truth companion to the
user-facing `README.md` (getting started) and `catalog-page.md` (catalog listing).

## Table of Contents

1. Introduction
2. The Experimental Channel (read this first)
3. The Gateway API Role-Oriented Model
4. Where TCPRoute Sits
5. Anatomy of TCPRouteSpec
6. The Backend Model
7. Validation Rules
8. Why TCPRoute Is a First-Class OpenMCF Component
9. Design Decisions
10. Composing in Infra Charts
11. Controller Landscape
12. Common Pitfalls
13. Conclusion
14. References

## Introduction

The Gateway API is the successor to the Kubernetes Ingress API: a standards-based,
role-oriented, expressive specification for north-south (and increasingly
east-west) traffic. `TCPRoute` is the simplest route kind: it forwards raw TCP
connections arriving on a Gateway listener to a set of backend Services. There is
no matching of any kind -- the listener's port selects the traffic, and the route
forwards it (splitting by weight across backends). Use it to put non-HTTP TCP
services (Postgres, Redis, Kafka, a custom protocol) behind a Gateway.

`KubernetesTcpRoute` brings that resource into OpenMCF as a first-class deployment
component at 100% upstream fidelity, so customers never have to fall back to a raw
`KubernetesManifest` to express TCP routing.

## The Experimental Channel (read this first)

Unlike `KubernetesHttpRoute`, `KubernetesGrpcRoute`, and `KubernetesTlsRoute`
(all standard channel), **TCPRoute exists only in the Gateway API experimental
channel** and is served as `gateway.networking.k8s.io/v1alpha2`. Consequences:

- The prerequisite CRDs must be installed from the experimental channel:
  `KubernetesGatewayApiCrds` with `install_channel: experimental`. The standard
  channel has no TCPRoute CRD, so a standard-channel install makes this component
  undeployable.
- The Gateway controller must support TCPRoute (not all do).
- The typed crd2pulumi resource is `gatewayv1alpha2.NewTCPRoute`, and the
  Terraform `kubernetes_manifest` uses apiVersion `gateway.networking.k8s.io/v1alpha2`.

This is the family's first wholly-experimental resource; the version and channel
are dictated entirely by upstream and the regenerated crd2pulumi types.

## The Gateway API Role-Oriented Model

The Gateway API splits responsibilities across personas:

- **Infrastructure provider** owns the `GatewayClass` (which controller implements
  Gateways).
- **Cluster operator** owns `Gateway` objects (listeners, ports, TLS).
- **Application developer** owns `*Route` objects that attach to a Gateway.

`TCPRoute` is in the application-developer lane. A route attaches to a Gateway via
`parentRefs`; the Gateway's TCP listener decides which route kinds and namespaces
may attach (via `allowedRoutes`). A TCPRoute must attach to a listener of protocol
`TCP`.

## Where TCPRoute Sits

```
KubernetesGatewayApiCrds        (install EXPERIMENTAL CRDs + controller)
        |
KubernetesGatewayClass          (controller class)
        |
KubernetesGateway               (TCP listener)
        |
KubernetesTcpRoute              (this component: raw TCP forwarding)
        |
backend Services
```

## Anatomy of TCPRouteSpec

The OpenMCF spec flattens the upstream `TCPRouteSpec` after the standard
namespaced envelope (`target_cluster`, `namespace`):

- `parent_refs` -- the Gateways this route attaches to (max 32).
- `use_default_gateways` -- experimental default-Gateway attachment (`All` /
  `None`); see Design Decisions.
- `rules` -- 1 to 16 routing rules. Each rule has only:
  - an optional `name`;
  - `backend_refs` (1 to 16) -- weighted backends.

There are no hostnames, matches, or filters: a TCP route has no application-layer
visibility.

## The Backend Model

A TCP route forwards to the shared `KubernetesGatewayApiBackendRef`
(group/kind/name/namespace/port/weight). Because TCP routes have no per-backend
filters, they reuse the canonical shared backend reference directly rather than
defining a per-route backend ref (the filter-carrying HTTP/GRPC routes flatten
their own). If a backend is invalid, the implementation rejects connection
attempts in proportion to the backend's weight.

## Validation Rules

Every upstream `XValidation` and kubebuilder marker is translated to
`buf.validate`:

- `parent_refs`: `max_items: 32`.
- `use_default_gateways`: closed enum CEL `in ['All', 'None']`.
- `rules`: `min_items: 1`, `max_items: 16`.
- `backend_refs`: `min_items: 1`, `max_items: 16`; the shared backend ref carries
  the upstream group/kind/name patterns and port/weight bounds.
- `name` (rule): SectionName pattern (lowercase RFC 1123 subdomain, 1-253).

The experimental rule-name-uniqueness `XValidation` is not translated (consistent
with the rest of the route family -- it is `<gateway:experimental>` even within
the experimental CRD and is controller-enforced).

## Why TCPRoute Is a First-Class OpenMCF Component

Without it, TCP routing forces customers back to raw `KubernetesManifest` YAML: no
proto validation, no typed SDKs, no FK wiring, no InfraChart composability, no UI
wizards. `KubernetesTcpRoute` closes that gap for the raw-TCP slice of the ingress
story.

## Design Decisions

- **Experimental channel, `v1alpha2`.** Dictated by upstream; the typed
  crd2pulumi resource and Terraform manifest both use the `v1alpha2` apiVersion.
- **`use_default_gateways` is included.** This experimental `CommonRouteSpec`
  field is reachable on the experimental TCPRoute deployment target (it is absent
  from the standard-channel routes' targets, which is the only reason HttpRoute /
  GrpcRoute excluded it). Including it is 100% fidelity (DD-001), not divergence
  from the siblings; the difference is explained inline in `spec.proto`.
- **Plain `parent_refs` / `backend_refs` (DD-009).** Kept as plain upstream
  references, not `StringValueOrRef`; their DAG edges are expressed via
  `metadata.relationships`.
- **Reuse the shared backend ref.** TCP routes have no per-backend filters, so
  they consume the canonical `KubernetesGatewayApiBackendRef` directly.
- **DD-008 value modeling.** Closed enum (`use_default_gateways`) uses CEL `in`;
  no proto enums and no baked-in OpenMCF defaults on upstream fields.
- **Typed crd2pulumi IaC.** `gatewayv1alpha2.NewTCPRoute`; the spec mapping is
  split across `parent_refs.go` and `rules.go` (no matches/filters).

## Composing in Infra Charts

`KubernetesTcpRoute` is designed as a LEGO block for Infra Charts. Two mechanisms
wire it into the dependency DAG (see project decision DD-009):

1. **Data dependencies use `StringValueOrRef` (`valueFrom`).** `namespace` is a
   `StringValueOrRef` with `default_kind = KubernetesNamespace`, so it can
   reference a namespace output and the platform builds the DAG edge automatically.
2. **Topology dependencies use `metadata.relationships`.** `parent_refs` and
   `backend_refs` are plain references -- a plain name string creates no automatic
   DAG edge. Infra-chart authors express those edges explicitly:

```yaml
metadata:
  name: "{{ values.env }}-postgres-route"
  relationships:
    - kind: KubernetesGateway
      name: "{{ values.env }}-gateway"
      type: depends_on
    - kind: KubernetesService          # when the backend is OpenMCF-managed
      name: "{{ values.service_name }}"
      type: uses
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.env }}-ns"
      fieldPath: spec.name
  parentRefs:
    - name: "{{ values.env }}-gateway"
      sectionName: tcp
  rules:
    - backendRefs:
        - name: "{{ values.service_name }}"
          port: 5432
```

Data edges use `valueFrom`; topology edges (the plain refs) use
`metadata.relationships`. Both are required for the platform to build the full DAG.

## Controller Landscape

TCPRoute is implemented by Gateway API controllers that support layer-4 routing
(Istio, Envoy Gateway, and others) -- and only when the experimental channel is
installed. The proto carries no controller-specific behavior; upstream defaults
are documented in comments and left to the controller rather than baked into the
spec.

## Common Pitfalls

- **Experimental CRDs required.** Installing `KubernetesGatewayApiCrds` in the
  default standard channel leaves no TCPRoute CRD on the cluster, and the apply
  fails. Use `install_channel: experimental`.
- **Listener must be `TCP` protocol.** A TCPRoute attached to an HTTP/HTTPS/TLS
  listener will not attach.
- **No matching.** TCP routes cannot match on hostnames, paths, headers, or
  methods -- they forward by listener port. Use TLSRoute (SNI) or HTTPRoute
  (layer 7) when you need matching.
- **Plain refs need relationships in charts.** Setting only `parentRefs[].name`
  does not create a DAG edge; add `metadata.relationships` (DD-009).

## Conclusion

`KubernetesTcpRoute` completes the raw-TCP slice of the OpenMCF Gateway API
ingress layer at 100% fidelity with the upstream experimental channel, mirroring
its sibling routes in every convention and fully accounting for infra-chart
composability.

## References

- [Gateway API TCPRoute](https://gateway-api.sigs.k8s.io/api-types/tcproute/)
- Upstream types: `kubernetes-sigs/gateway-api` `apis/v1alpha2/tcproute_types.go` (v1.5.1)
- Sibling components: `KubernetesTlsRoute`, `KubernetesHttpRoute`, `KubernetesGrpcRoute`
- Project decision: DD-009 (infra-chart composability, plain refs + relationships)
