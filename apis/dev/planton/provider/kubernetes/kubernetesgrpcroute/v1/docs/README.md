# KubernetesGrpcRoute -- Research Documentation

This document explains why `KubernetesGrpcRoute` exists, how it maps the upstream
Gateway API `GRPCRoute` into the Planton component model, and the design
decisions behind its spec, validation, and IaC. It is the source-of-truth
companion to the user-facing `README.md` (getting started) and `catalog-page.md`
(catalog listing).

## Table of Contents

1. Introduction
2. The Gateway API Role-Oriented Model
3. Where GRPCRoute Sits
4. Anatomy of GRPCRouteSpec
5. The Match / Filter / Backend Model
6. Validation Rules
7. Why GRPCRoute Is a First-Class Planton Component
8. Design Decisions
9. Standard vs Experimental Channel
10. Composing in Infra Charts
11. Controller Landscape
12. Common Pitfalls
13. Conclusion
14. References

## Introduction

The Gateway API is the successor to the Kubernetes Ingress API: a standards-based,
role-oriented, expressive specification for north-south (and increasingly
east-west) traffic. `GRPCRoute` is the resource that routes gRPC requests --
matching by hostname, by gRPC service/method, or by HTTP/2 header, then
forwarding to backend Services. Where a `Gateway` defines *where* traffic enters
(listeners, ports, TLS), a `GRPCRoute` defines *how* gRPC traffic is matched,
transformed, and forwarded.

`KubernetesGrpcRoute` brings that resource into Planton as a first-class
deployment component, at 100% fidelity with the upstream standard channel, so
customers never have to fall back to a raw `KubernetesManifest` to express gRPC
routing. It is the sibling of `KubernetesHttpRoute` and shares its envelope,
conventions, and IaC structure.

## The Gateway API Role-Oriented Model

The Gateway API deliberately splits responsibilities across personas:

- **Infrastructure provider** owns the `GatewayClass` (which controller, e.g.
  Istio or Envoy Gateway, implements Gateways).
- **Cluster operator** owns `Gateway` objects (listeners, ports, TLS, addresses).
- **Application developer** owns `*Route` objects (`HTTPRoute`, `GRPCRoute`, ...)
  that attach to a Gateway and define routing for their app.

`GRPCRoute` is squarely in the application-developer lane. A route attaches to a
Gateway via `parentRefs`; the Gateway's listeners decide which route kinds and
namespaces may attach (via `allowedRoutes`). gRPC requires HTTP/2, so the parent
listener must accept HTTP/2 -- h2c over an `HTTP` listener, or HTTP/2 (via ALPN)
over an `HTTPS` listener.

## Where GRPCRoute Sits

```
KubernetesGatewayApiCrds        (install CRDs + controller)
        |
KubernetesGatewayClass          (controller class)
        |
KubernetesGateway               (listeners, ports, TLS; HTTP/2-capable)
        |
KubernetesGrpcRoute             (this component: service/method routing)
        |
backend gRPC Services
```

## Anatomy of GRPCRouteSpec

The Planton spec flattens the upstream `GRPCRouteSpec` after the standard
namespaced envelope (`target_cluster`, `namespace`):

- `parent_refs` -- the Gateways this route attaches to (max 32).
- `hostnames` -- authority values that select the route (max 16; wildcard prefix
  allowed).
- `rules` -- 1 to 16 routing rules. Each rule has:
  - `matches` (max 64) -- each a `method` matcher and/or `headers` matchers;
  - `filters` (max 16) -- header modifiers, request mirror, or extension ref;
  - `backend_refs` (max 16) -- weighted backends.

Unlike `HTTPRouteRule`, `GRPCRouteRule` has no `timeouts` field.

## The Match / Filter / Backend Model

- **Match** (`KubernetesGrpcRouteMatch`): a `method` matcher
  (`KubernetesGrpcRouteMethodMatch` with `type` = `Exact`|`RegularExpression`,
  `service`, `method`) and/or `headers` matchers. Conditions within a match are
  ANDed; multiple matches are ORed. At least one of `service`/`method` must be set
  for an `Exact` method match.
- **Filter** (`KubernetesGrpcRouteFilter`): a discriminated union with a `type`
  discriminator selecting exactly one config field. GRPC supports four variants:
  `RequestHeaderModifier`, `ResponseHeaderModifier`, `RequestMirror`,
  `ExtensionRef`. (There is no redirect, URL rewrite, or CORS filter -- those are
  HTTP-only.)
- **Backend ref** (`KubernetesGrpcRouteBackendRef`): the flattened upstream
  backend (group/kind/name/namespace/port/weight) plus optional per-backend
  filters.

## Validation Rules

Every upstream `XValidation` and kubebuilder marker is translated to
`buf.validate`:

- Method match: at least one of `service`/`method` (for `Exact`); `service` and
  `method` character regexes gated on the `Exact` type.
- Header match / header filter: HeaderName pattern; name uniqueness within a
  match and within `set`/`add`.
- Filter union: one biconditional per variant (`has(field) == (type == 'X')`).
- Rule level: `RequestHeaderModifier` / `ResponseHeaderModifier` each at most once.
- Request mirror: `percent` xor `fraction`.
- Closed enums (`type` fields) use CEL `in [...]`; open sets (hostnames, header
  names) use the upstream regex (DD-008).

Two upstream rules are intentionally **not** translated, matching the
`KubernetesHttpRoute` sibling: the aggregate "total matches across all rules
< 128" XValidation (a 16-way unrolled CEL sum -- a controller-enforced aggregate,
unmaintainable as proto CEL, and already bounded by `rules<=16`/`matches<=64`),
and the experimental rule-name-uniqueness rule (`<gateway:experimental>`).

## Why GRPCRoute Is a First-Class Planton Component

Without it, gRPC routing forces customers back to raw `KubernetesManifest` YAML:
no proto validation, no typed SDKs, no FK wiring, no InfraChart composability, no
UI wizards. `KubernetesGrpcRoute` closes that gap for the gRPC half of the
ingress story, alongside `KubernetesHttpRoute`.

## Design Decisions

- **Standard channel only.** Experimental `sessionPersistence` (on the rule) and
  `useDefaultGateways` (on CommonRouteSpec) are excluded -- they are absent from
  the standard-channel CRD and the typed crd2pulumi resource, so they have no
  deployable target.
- **Plain `parent_refs` / `backend_refs` (DD-009).** Kept as plain upstream
  references, not `StringValueOrRef`, because they are arrays of multi-field
  objects; wrapping them would break 100% fidelity and distort the typed CRD
  shape. Their DAG edges are expressed via `metadata.relationships` (see below).
- **Flat backend ref.** The six backend fields plus filters are flattened to
  match the upstream inline `GRPCBackendRef` and the flat crd2pulumi shape.
- **DD-008 value modeling.** Closed enums use CEL `in`; open sets use the upstream
  regex; no proto enums and no baked-in Planton defaults on upstream fields.
- **Typed crd2pulumi IaC.** `gatewayv1.NewGRPCRoute`, with rule-level and
  backend-level filter builders kept separate because crd2pulumi emits distinct
  Go types for the same JSON shape.

## Standard vs Experimental Channel

Planton provisions the standard-channel GRPCRoute. Excluded experimental fields:

| Field | Reason |
|-------|--------|
| `GRPCRouteRule.sessionPersistence` | Experimental; absent from the standard CRD and typed resource. |
| `CommonRouteSpec.useDefaultGateways` | Experimental; routes always flatten `parent_refs` directly. |

## Composing in Infra Charts

`KubernetesGrpcRoute` is designed as a LEGO block for Infra Charts. Two
mechanisms wire it into the dependency DAG (see project decision DD-009):

1. **Data dependencies use `StringValueOrRef` (`valueFrom`).** `namespace` is a
   `StringValueOrRef` with `default_kind = KubernetesNamespace`, so it can
   reference a namespace output and the platform builds the DAG edge
   automatically.
2. **Topology dependencies use `metadata.relationships`.** `parent_refs` and
   `backend_refs` are plain references -- a plain name string creates no automatic
   DAG edge. Infra-chart authors express those edges explicitly:

```yaml
metadata:
  name: "{{ values.env }}-greeter-route"
  relationships:
    - kind: KubernetesGateway
      name: "{{ values.env }}-gateway"
      type: depends_on
    - kind: KubernetesService          # when the backend is Planton-managed
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
  rules:
    - matches:
        - method:
            service: "{{ values.grpc_service }}"
      backendRefs:
        - name: "{{ values.service_name }}"
          port: 9000
```

The full ingress stack composes as:

```
KubernetesCertManager            (runs_on cluster)
  -> KubernetesClusterIssuer     (cert_manager_namespace: valueFrom CertManager)
    -> KubernetesCertificate     (issuerRef: valueFrom ClusterIssuer)
      -> (Kubernetes Secret)     (Certificate.status.outputs.secret_name)
        -> KubernetesGateway     (gateway_class_name + namespace: valueFrom;
                                   certificate_refs -> Certificate: relationships uses)
          -> KubernetesGrpcRoute (namespace: valueFrom;
                                  parent_refs -> Gateway: relationships depends_on;
                                  backend_refs -> Service: relationships uses)
```

Data edges use `valueFrom`; topology edges (the plain refs) use
`metadata.relationships`. Both are required for the platform to build the full DAG.

## Controller Landscape

GRPCRoute is implemented by the major Gateway API controllers (Istio, Envoy
Gateway, Contour, and others). The proto carries no controller-specific behavior;
upstream defaults (for example method match `type = Exact`) are documented in
comments and left to the controller rather than baked into the spec.

## Common Pitfalls

- **Listener must accept HTTP/2.** A GRPCRoute attached to a plain HTTP/1-only
  listener will not serve gRPC. Use h2c (`HTTP`) or HTTP/2 over `HTTPS`.
- **Plain refs need relationships in charts.** Setting only `parentRefs[].name`
  does not create a DAG edge; add `metadata.relationships` (DD-009).
- **`Exact` method match needs a value.** An `Exact` method matcher with neither
  `service` nor `method` is rejected.

## Conclusion

`KubernetesGrpcRoute` completes the gRPC half of the Planton Gateway API ingress
layer at 100% standard-channel fidelity, mirroring `KubernetesHttpRoute` in every
convention and fully accounting for infra-chart composability.

## References

- [Gateway API GRPCRoute](https://gateway-api.sigs.k8s.io/api-types/grpcroute/)
- Upstream types: `kubernetes-sigs/gateway-api` `apis/v1/grpcroute_types.go` (v1.5.1)
- Sibling component: `KubernetesHttpRoute`
- Project decision: DD-009 (infra-chart composability, plain refs + relationships)
