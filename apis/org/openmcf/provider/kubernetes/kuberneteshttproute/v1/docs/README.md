# KubernetesHttpRoute -- Research Documentation

This document explains why `KubernetesHttpRoute` exists, how it maps the upstream
Gateway API `HTTPRoute` into the OpenMCF component model, and the design
decisions behind its spec, validation, and IaC. It is the source-of-truth
companion to the user-facing `README.md` (getting started) and `catalog-page.md`
(catalog listing).

## Table of Contents

1. Introduction
2. The Gateway API Role-Oriented Model
3. Where HTTPRoute Sits
4. Anatomy of HTTPRouteSpec
5. The Match / Filter / Backend Model
6. Validation Rules
7. Why HTTPRoute Is a First-Class OpenMCF Component
8. Design Decisions
9. Standard vs Experimental Channel
10. Controller Landscape
11. 80/20 Scoping
12. Common Pitfalls
13. Conclusion
14. References

## Introduction

The Gateway API is the successor to the Kubernetes Ingress API: a standards-based,
role-oriented, expressive specification for north-south (and increasingly
east-west) traffic. `HTTPRoute` is its most-used resource -- the object that
actually says "requests for this host and path go to that Service." Where a
`Gateway` defines *where* traffic enters (listeners, ports, TLS), an `HTTPRoute`
defines *how* HTTP traffic is matched, transformed, and forwarded.

`KubernetesHttpRoute` brings that resource into OpenMCF as a first-class
deployment component, at 100% fidelity with the upstream standard channel, so
customers never have to fall back to a raw `KubernetesManifest` to express HTTP
routing.

## The Gateway API Role-Oriented Model

The Gateway API deliberately splits responsibilities across personas:

- **Infrastructure provider** owns the `GatewayClass` (which controller, e.g.
  Istio or Envoy Gateway, implements Gateways).
- **Cluster operator** owns `Gateway` objects (listeners, ports, TLS, addresses).
- **Application developer** owns `*Route` objects (`HTTPRoute`, `GRPCRoute`, ...)
  that attach to a Gateway and define routing for their app.

`HTTPRoute` is squarely in the application-developer lane. A route attaches to a
Gateway via `parentRefs`; the Gateway's listeners decide which route kinds and
namespaces may attach (via `allowedRoutes`). This separation lets platform teams
own the entry point while application teams own their routing -- exactly the
boundary OpenMCF's resource graph and InfraChart composition reinforce.

## Where HTTPRoute Sits

```
KubernetesGatewayApiCrds        (install CRDs + controller)
        |
KubernetesGatewayClass          (controller class)
        |
KubernetesGateway               (listeners + TLS, the entry point)
        |  parentRefs
KubernetesHttpRoute  --------->  backend Kubernetes Services
        ^
KubernetesNamespace             (namespace FK)
```

In an InfraChart, the DAG ordering is: CRDs -> GatewayClass -> Gateway -> Route.
The route's `parentRefs.name` references the Gateway by name, and its
`backendRefs.name` references Services by name. Neither is an OpenMCF foreign key
(see Design Decisions); the DAG provides ordering, and the names match upstream
exactly.

## Anatomy of HTTPRouteSpec

The OpenMCF spec flattens the upstream `HTTPRouteSpec` after the standard
Kubernetes envelope:

| OpenMCF field | Upstream | Notes |
|---------------|----------|-------|
| `target_cluster` | (n/a) | OpenMCF cluster selector. |
| `namespace` | (metadata) | FK to `KubernetesNamespace.spec.name`. |
| `parent_refs` | `CommonRouteSpec.parentRefs` | Gateways the route attaches to. |
| `hostnames` | `HTTPRouteSpec.hostnames` | Host header values (wildcard prefix allowed). |
| `rules` | `HTTPRouteSpec.rules` | Match / filter / backend rules (1-16). |

Each rule (`KubernetesHttpRouteRule`) contains:

- `name` -- optional unique rule name.
- `matches` -- up to 64 `HTTPRouteMatch` predicates (path, headers, query params,
  method). A request matches the rule if any one match is satisfied; conditions
  within a match are ANDed.
- `filters` -- up to 16 processing steps (see below).
- `backend_refs` -- up to 16 weighted backends.
- `timeouts` -- `request` and `backend_request` durations.

## The Match / Filter / Backend Model

**Matches** select traffic. A path match has a `type` (`Exact`, `PathPrefix`,
`RegularExpression`) and a `value`; header and query-param matches add
name/value predicates with `Exact` or `RegularExpression` typing; `method`
matches the HTTP verb.

**Filters** transform traffic. `HTTPRoute` uses a discriminated union: a `type`
discriminator selects exactly one configuration field. The standard-channel
filter types are:

- `RequestHeaderModifier` / `ResponseHeaderModifier` -- set/add/remove headers.
- `RequestRedirect` -- respond with an HTTP redirect (cannot be combined with
  backends or with `URLRewrite`).
- `URLRewrite` -- rewrite host and/or path while forwarding.
- `RequestMirror` -- copy a percentage/fraction of traffic to a second backend.
- `CORS` -- emit CORS response headers.
- `ExtensionRef` -- reference an implementation-specific filter resource.

Filters may appear at the rule level (apply to all backends) or on an individual
backend ref (apply only when forwarding to that backend).

**Backends** receive traffic, split by `weight` (`weight / sum(weights)`). This
is the foundation of canary and blue/green delivery.

## Validation Rules

Every upstream `XValidation` and kubebuilder marker is translated into
`buf.validate`, so the same configuration errors are caught at apply time in
OpenMCF that the CRD would reject:

- **Path values** (Exact/PathPrefix): must be absolute, normalized (no `//`,
  `/./`, `/../`, trailing `/.`/`/..`), and free of `#`, `%2f`, `%2F`; restricted
  to valid path characters.
- **Filter union consistency**: the populated configuration field must match the
  `type` discriminator (expressed as a biconditional per variant -- equivalent to
  the upstream pair of "must be nil if type != X" / "must be set if type == X").
- **Rule-level filter rules**: `RequestRedirect` excludes backends; `RequestRedirect`
  and `URLRewrite` are mutually exclusive; non-additive filters
  (`RequestHeaderModifier`, `ResponseHeaderModifier`, `RequestRedirect`,
  `URLRewrite`, `CORS`) may appear at most once; a `replacePrefixMatch` rewrite
  requires exactly one `PathPrefix` match.
- **Request mirror**: `percent` and `fraction` are mutually exclusive.
- **CORS**: `*` cannot appear alongside other origins / methods / headers.
- **Timeouts**: `backend_request` cannot exceed `request` (skipped when `request`
  is `0s`, which disables the timeout).
- **Uniqueness**: header-match and query-param-match names are unique within a
  match; header `set`/`add` names are unique within a header filter.
- **Closed enums** use CEL `in [...]` (method, match types, path-modifier type,
  redirect scheme/status); **open sets** use the upstream regex (hostnames,
  header names, path values) -- see Design Decisions.

Every CEL message is written to be read by a human configuring the route, not a
compiler: it states the constraint and (for conditional rules) the condition in
plain English.

## Why HTTPRoute Is a First-Class OpenMCF Component

Without a first-class component, HTTP routing on the Gateway API requires raw
`KubernetesManifest` YAML, which forfeits the entire OpenMCF value proposition:
no proto validation, no typed SDKs, no Planton UI wizards, no InfraChart
composition. `HTTPRoute` is the single most common Gateway API resource, so this
gap is the most painful one to leave open. Modeling it as a typed component gives:

- compile-time-validated Pulumi and Terraform modules,
- protovalidate enforcement of every upstream rule before anything reaches the
  cluster,
- DAG-ordered composition with the Gateway, namespace, and backends.

## Design Decisions

1. **100% standard-channel fidelity.** Every standard-channel field of the
   upstream `HTTPRouteSpec` is modeled. No subsetting, because the Gateway API is
   itself "the what" -- any combination of routing features is legitimate.
2. **String value-fields, not proto enums (DD-008).** Closed enums validate with
   CEL `in [...]`; open sets (hostnames, header names, paths) validate with the
   upstream regex. This preserves custom values where upstream allows them, keeps
   CEL expressions readable, and requires zero case-mapping in IaC.
3. **Discriminated unions via discriminator + fields, never `oneof`.** Filters
   and path modifiers keep the upstream JSON structure; CEL enforces that the
   populated field matches the `type`.
4. **Plain `parent_refs` and `backend_refs` (not foreign keys).** They are arrays
   of multi-field upstream objects; wrapping each in `StringValueOrRef` would
   distort the structure. InfraChart DAG ordering sequences the Gateway and
   backends; the `name` fields match upstream exactly.
5. **No baked-in OpenMCF defaults on upstream fields.** Upstream kubebuilder
   defaults (path `type=PathPrefix`, match `type=Exact`, redirect
   `status_code=302`, CORS `max_age=5`) are documented in comments only, so the
   controller/CRD stays authoritative.
6. **Flat `HTTPBackendRef`.** Upstream embeds `BackendRef` inline and the typed
   CRD shape is flat, so the backend-ref fields are flattened directly (plus a
   `filters` list) rather than nesting the shared `BackendRef` message.
7. **Typed crd2pulumi IaC.** Both IaC modules use the generated typed resource
   (`NewHTTPRoute`), catching field and structure errors at compile time.

## Standard vs Experimental Channel

The component models the **standard channel** only, matching the CRD and typed
Pulumi resource OpenMCF provisions with. The following upstream fields are
`<gateway:experimental>` and intentionally excluded, because they are absent from
the standard CRD and have no deployable target:

- `CommonRouteSpec.useDefaultGateways`
- `HTTPRouteRule.retry` and `HTTPRouteRule.sessionPersistence`
- the `ExternalAuth` filter variant (and its sub-configuration)

`CORS` and `Timeouts` *are* in the standard channel and are fully modeled. When a
future Gateway API release promotes the experimental fields, they can be added in
a subsequent revision.

## Controller Landscape

`HTTPRoute` is implemented by every conformant Gateway API controller, including
Istio, Envoy Gateway, NGINX Gateway Fabric, Cilium, Kong, Traefik, and the cloud
load-balancer controllers (GKE, AWS). Conformance levels in the spec
("Core" / "Extended" / "Implementation-specific") indicate how broadly a feature
is supported; the field comments in `spec.proto` carry the upstream support level
for each field so users can reason about portability.

## 80/20 Scoping

For HTTP routing, "the what" is the Gateway API itself -- subsetting would force
users back to raw YAML for any pattern outside the chosen subset. So scoping here
is about **channel** (standard, not experimental) and **resource set** (this
component is one of seven in the Gateway API family), not about which routing
features to expose. Within the standard channel, every field is modeled.

## Common Pitfalls

- **Redirect + backends.** A `RequestRedirect` filter terminates the request;
  putting `backend_refs` on the same rule is rejected. Use a dedicated
  redirect-only rule.
- **Redirect vs URL rewrite.** Both rewrite the request line and cannot be
  combined on one rule.
- **`replacePrefixMatch` requires a single PathPrefix match.** A prefix rewrite
  needs exactly one match whose path type is `PathPrefix` to define the prefix.
- **Cross-namespace refs need a ReferenceGrant.** A backend or parent in another
  namespace requires a `ReferenceGrant` in the target namespace.
- **CORS `*` exclusivity.** `*` cannot be listed alongside concrete origins,
  methods, or headers.
- **Timeout format.** Timeouts are GEP-2257 duration strings (e.g. `10s`, `1m`,
  `500ms`), and `backend_request` must not exceed `request`.

## Conclusion

`KubernetesHttpRoute` closes the most important gap in OpenMCF's networking
layer: declarative, validated, composable HTTP routing on the Gateway API. It
mirrors the upstream standard channel exactly, enforces every upstream rule with
human-friendly messages, and ships typed Pulumi and Terraform modules -- so HTTP
routing is a first-class OpenMCF experience rather than raw YAML.

## References

- [Gateway API HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/)
- [Gateway API specification](https://gateway-api.sigs.k8s.io/)
- [GEP-2257: Gateway API Duration format](https://gateway-api.sigs.k8s.io/geps/gep-2257/)
- [Gateway API conformance](https://gateway-api.sigs.k8s.io/concepts/conformance/)
- Upstream source: `kubernetes-sigs/gateway-api` `apis/v1/httproute_types.go` (v1.5.1)

## Composing in Infra Charts

`KubernetesHttpRoute` is a leaf in the ingress DAG: it attaches to a `Gateway`
and forwards to backend Services. Two mechanisms wire it to its neighbors (see
project decision DD-009):

1. **Data dependencies use `valueFrom`.** `namespace` is a `StringValueOrRef`, so
   it can reference a `KubernetesNamespace` output and the platform builds that DAG
   edge automatically.
2. **Topology dependencies use `metadata.relationships`.** `parentRefs` and
   `backendRefs` are **plain** upstream references (arrays of multi-field objects),
   not foreign keys -- a plain name creates no automatic DAG edge. Express those
   edges explicitly:

```yaml
metadata:
  name: "{{ values.env }}-web-route"
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
    - name: "{{ values.env }}-gateway"   # literal Gateway name
  hostnames:
    - "app.{{ values.domain }}"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: "{{ values.service_name }}"
          port: 8080
```

Full ingress stack:
`CertManager -> ClusterIssuer -> Certificate -> (Secret) -> Gateway -> HTTPRoute / GRPCRoute`.
Data edges use `valueFrom`; the plain `parentRefs` / `backendRefs` edges use
`metadata.relationships`.
