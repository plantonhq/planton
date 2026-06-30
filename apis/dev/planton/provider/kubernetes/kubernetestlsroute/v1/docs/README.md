# KubernetesTlsRoute -- Research Documentation

This document explains why `KubernetesTlsRoute` exists, how it maps the upstream
Gateway API `TLSRoute` into the Planton component model, and the design decisions
behind its spec, validation, and IaC. It is the source-of-truth companion to the
user-facing `README.md` (getting started) and `catalog-page.md` (catalog listing).

## Table of Contents

1. Introduction
2. The Gateway API Role-Oriented Model
3. Where TLSRoute Sits
4. Anatomy of TLSRouteSpec
5. The SNI / Backend Model
6. Validation Rules
7. Why TLSRoute Is a First-Class Planton Component
8. Design Decisions
9. Standard vs Experimental Channel (TLSRoute graduated to v1)
10. Composing in Infra Charts
11. Controller Landscape
12. Common Pitfalls
13. Conclusion
14. References

## Introduction

The Gateway API is the successor to the Kubernetes Ingress API: a standards-based,
role-oriented, expressive specification for north-south (and increasingly
east-west) traffic. `TLSRoute` routes TLS connections by their SNI hostname and
forwards them unmodified to a backend -- a layer-4 "passthrough" route. The
backend, not the Gateway, terminates TLS, so the encrypted byte stream is carried
end to end. This is the right tool when a service must hold its own certificate or
do its own mTLS, and the Gateway should route but never decrypt.

`KubernetesTlsRoute` brings that resource into Planton as a first-class deployment
component, at 100% fidelity with the upstream standard channel, so customers never
have to fall back to a raw `KubernetesManifest` to express TLS passthrough
routing. It is a sibling of `KubernetesHttpRoute` / `KubernetesGrpcRoute` and
shares their envelope, conventions, and IaC structure.

## The Gateway API Role-Oriented Model

The Gateway API deliberately splits responsibilities across personas:

- **Infrastructure provider** owns the `GatewayClass` (which controller, e.g.
  Istio or Envoy Gateway, implements Gateways).
- **Cluster operator** owns `Gateway` objects (listeners, ports, TLS, addresses).
- **Application developer** owns `*Route` objects (`HTTPRoute`, `TLSRoute`, ...)
  that attach to a Gateway and define routing for their app.

`TLSRoute` is in the application-developer lane. A route attaches to a Gateway via
`parentRefs`; the Gateway's TLS listener decides which route kinds and namespaces
may attach (via `allowedRoutes`). A TLSRoute must attach to a listener of protocol
`TLS` -- `Passthrough` mode for core support, `Terminate` mode for the extended
`TLSRouteTermination` feature.

## Where TLSRoute Sits

```
KubernetesGatewayApiCrds        (install CRDs + controller)
        |
KubernetesGatewayClass          (controller class)
        |
KubernetesGateway               (TLS listener, Passthrough mode)
        |
KubernetesTlsRoute              (this component: SNI-based passthrough)
        |
backend Services (terminate TLS themselves)
```

## Anatomy of TLSRouteSpec

The Planton spec flattens the upstream `TLSRouteSpec` after the standard
namespaced envelope (`target_cluster`, `namespace`):

- `parent_refs` -- the Gateways this route attaches to (max 32).
- `hostnames` -- the SNI hostnames that select the route (**required**, 1 to 16;
  wildcard prefix allowed; IPs not allowed).
- `rules` -- **exactly one** routing rule (`min_items: 1`, `max_items: 1`). Each
  rule has only:
  - an optional `name`;
  - `backend_refs` (1 to 16) -- weighted backends.

There are no matches and no filters: a TLS passthrough route has no layer-7
visibility to match on or transform.

## The SNI / Backend Model

- **Hostnames** are matched against the SNI attribute of the TLS ClientHello, not
  an HTTP Host header (the Gateway cannot read HTTP -- the stream is encrypted). A
  leading `*.` is a single-label wildcard suffix match. Per RFC 6066, SNI names
  may not be IP addresses.
- **Backend ref** (the shared `KubernetesGatewayApiBackendRef`):
  group/kind/name/namespace/port/weight. Because TLS routes have no per-backend
  filters, they reuse the canonical shared backend reference directly rather than
  defining a per-route backend ref (the filter-carrying HTTP/GRPC routes flatten
  their own).

## Validation Rules

Every upstream `XValidation` and kubebuilder marker is translated to
`buf.validate`:

- `hostnames`: required (`min_items: 1`), `max_items: 16`; per-item RFC 1123
  wildcard-prefix pattern; a list-level CEL rejecting IPv4 literals (the SNI no-IP
  rule from RFC 6066).
- `rules`: `min_items: 1`, `max_items: 1` (upstream caps a TLSRoute at one rule).
- `backend_refs`: `min_items: 1`, `max_items: 16`; the shared backend ref carries
  the upstream group/kind/name patterns and port/weight bounds.
- `name` (rule): SectionName pattern (lowercase RFC 1123 subdomain, 1-253).

### A note on `isIp`

Upstream expresses the SNI no-IP rule as `self.all(h, !isIP(h))`. Planton's
build-time `buf lint` CEL environment does not register protovalidate's `isIp()`
format function (only the runtime validator does), so the rule is translated as an
equivalent IPv4 dotted-quad regex. IPv6 literals contain `:` and are already
rejected by the per-item hostname pattern, so the IPv4 guard fully covers the
reachable input space. This is documented inline in `spec.proto` so a future agent
does not "restore" `isIp` and break the lint gate.

## Why TLSRoute Is a First-Class Planton Component

Without it, TLS passthrough routing forces customers back to raw
`KubernetesManifest` YAML: no proto validation, no typed SDKs, no FK wiring, no
InfraChart composability, no UI wizards. `KubernetesTlsRoute` closes that gap for
the TLS-passthrough slice of the ingress story, alongside `KubernetesHttpRoute`
and `KubernetesGrpcRoute`.

## Design Decisions

- **Standard channel, `v1`.** TLSRoute graduated to the standard channel and is
  served as `gateway.networking.k8s.io/v1` (it was experimental `v1alpha2` /
  `v1alpha3` in earlier releases). The typed crd2pulumi resource emits the `v1`
  apiVersion (aliased from the older versions).
- **Plain `parent_refs` / `backend_refs` (DD-009).** Kept as plain upstream
  references, not `StringValueOrRef`, because they are arrays of multi-field
  objects; wrapping them would break 100% fidelity and distort the typed CRD
  shape. Their DAG edges are expressed via `metadata.relationships` (see below).
- **Reuse the shared backend ref.** TLS routes have no per-backend filters, so
  they consume the canonical `KubernetesGatewayApiBackendRef` directly -- the
  first consumer of that shared type (hardened to full upstream field fidelity in
  this change).
- **DD-008 value modeling.** Open sets use the upstream regex; no proto enums and
  no baked-in Planton defaults on upstream fields.
- **Typed crd2pulumi IaC.** `gatewayv1.NewTLSRoute`; the spec mapping is split
  across `parent_refs.go` and `rules.go` (no matches/filters).

## Standard vs Experimental Channel (TLSRoute graduated to v1)

Unlike `KubernetesTcpRoute` (which is experimental-only, `v1alpha2`), TLSRoute is
served in the **standard** channel as `v1`. The standard-channel `TLSRouteSpec`
has no `useDefaultGateways` field (that experimental field is stripped from the
standard CRD), so there is nothing experimental to exclude here -- the spec is the
full standard-channel surface.

## Composing in Infra Charts

`KubernetesTlsRoute` is designed as a LEGO block for Infra Charts. Two mechanisms
wire it into the dependency DAG (see project decision DD-009):

1. **Data dependencies use `StringValueOrRef` (`valueFrom`).** `namespace` is a
   `StringValueOrRef` with `default_kind = KubernetesNamespace`, so it can
   reference a namespace output and the platform builds the DAG edge automatically.
2. **Topology dependencies use `metadata.relationships`.** `parent_refs` and
   `backend_refs` are plain references -- a plain name string creates no automatic
   DAG edge. Infra-chart authors express those edges explicitly:

```yaml
metadata:
  name: "{{ values.env }}-secure-route"
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
      sectionName: tls
  hostnames:
    - "secure.{{ values.domain }}"
  rules:
    - backendRefs:
        - name: "{{ values.service_name }}"
          port: 8443
```

The full ingress stack composes as:

```
KubernetesCertManager            (runs_on cluster)
  -> KubernetesClusterIssuer     (cert_manager_namespace: valueFrom CertManager)
    -> KubernetesCertificate     (issuerRef: valueFrom ClusterIssuer)
      -> (Kubernetes Secret)     (Certificate.status.outputs.secret_name)
        -> KubernetesGateway     (gateway_class_name + namespace: valueFrom;
                                   certificate_refs -> Certificate: relationships uses)
          -> KubernetesTlsRoute  (namespace: valueFrom;
                                  parent_refs -> Gateway: relationships depends_on;
                                  backend_refs -> Service: relationships uses)
```

(For pure TLS passthrough the backend holds its own certificate, so the
cert-manager prefix is optional; it applies when the Gateway terminates TLS in the
extended `Terminate` mode.)

Data edges use `valueFrom`; topology edges (the plain refs) use
`metadata.relationships`. Both are required for the platform to build the full DAG.

## Controller Landscape

TLSRoute is implemented by the major Gateway API controllers that support
layer-4 routing (Istio, Envoy Gateway, and others). The proto carries no
controller-specific behavior; upstream defaults are documented in comments and
left to the controller rather than baked into the spec.

## Common Pitfalls

- **Listener must be `TLS` protocol.** A TLSRoute attached to an `HTTP`/`HTTPS`
  listener will not attach. The parent listener must use protocol `TLS`
  (`Passthrough` for core support).
- **No HTTP-level matching.** TLSRoutes route only by SNI hostname -- there is no
  path, header, or method matching. Use `KubernetesHttpRoute` when the Gateway
  terminates TLS and you need layer-7 routing.
- **Exactly one rule.** Upstream caps a TLSRoute at one rule; express traffic
  splitting through multiple weighted `backendRefs` in that single rule.
- **Plain refs need relationships in charts.** Setting only `parentRefs[].name`
  does not create a DAG edge; add `metadata.relationships` (DD-009).

## Conclusion

`KubernetesTlsRoute` completes the TLS-passthrough slice of the Planton Gateway
API ingress layer at 100% standard-channel fidelity, mirroring its sibling routes
in every convention and fully accounting for infra-chart composability.

## References

- [Gateway API TLSRoute](https://gateway-api.sigs.k8s.io/api-types/tlsroute/)
- Upstream types: `kubernetes-sigs/gateway-api` `apis/v1/tlsroute_types.go` (v1.5.1)
- Sibling components: `KubernetesHttpRoute`, `KubernetesGrpcRoute`, `KubernetesTcpRoute`
- Project decision: DD-009 (infra-chart composability, plain refs + relationships)
