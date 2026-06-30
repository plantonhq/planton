# KubernetesReferenceGrant -- Research Documentation

This document explains why `KubernetesReferenceGrant` exists, how it maps the
upstream Gateway API `ReferenceGrant` into the Planton component model, and the
design decisions behind its spec, validation, and IaC. It is the source-of-truth
companion to the user-facing `README.md` (getting started) and `catalog-page.md`
(catalog listing).

## Table of Contents

1. Introduction
2. The Cross-Namespace Trust Problem
3. The Role-Oriented Model
4. Anatomy of ReferenceGrantSpec
5. The from / to Model
6. Validation Rules
7. Why ReferenceGrant Is a First-Class Planton Component
8. Design Decisions
9. No Status Subresource
10. Composing in Infra Charts
11. Controller Landscape
12. Common Pitfalls
13. Conclusion
14. References

## Introduction

The Gateway API is the successor to the Kubernetes Ingress API: a standards-based,
role-oriented, expressive specification for north-south (and increasingly
east-west) traffic. `ReferenceGrant` is the API's cross-namespace trust primitive.
By default, a Gateway API object may only reference objects in its own namespace; a
reference that crosses a namespace boundary -- a Gateway's TLS `certificateRefs`
pointing at a Secret in a cert-manager namespace, or a Route's `backendRefs`
pointing at a Service in another team's namespace -- is denied unless the *target*
namespace publishes a `ReferenceGrant` that trusts the source.

`KubernetesReferenceGrant` brings that resource into Planton as a first-class
deployment component, at 100% fidelity with the upstream standard channel, so
customers never have to fall back to a raw `KubernetesManifest` to authorize
cross-namespace references. It is the seventh and final member of the Planton
Gateway API family, alongside `KubernetesGatewayClass`, `KubernetesGateway`, and
the four route kinds.

## The Cross-Namespace Trust Problem

A multi-tenant or multi-team cluster deliberately separates concerns by namespace:
ingress in `istio-ingress`, certificates in `cert-manager`, each application in its
own namespace. The Gateway API's default-deny posture on cross-namespace references
prevents a route in one namespace from silently hijacking traffic to, or exposing,
a Service or Secret it was never meant to touch. `ReferenceGrant` is the explicit,
auditable opt-in: the owner of the *referenced* namespace declares precisely which
*kinds* of resources, from precisely which namespaces, are allowed to reference
into it. Removing the grant immediately revokes the access.

## The Role-Oriented Model

The Gateway API splits responsibilities across personas. `ReferenceGrant` lives
with the owner of the resource being referenced (the cluster operator who owns the
Secret namespace, or the team that owns the backend Service namespace). It is the
counterpart to the application developer's `*Route` objects: the route *requests* a
cross-namespace reference, and the grant *authorizes* it from the other side.

## Anatomy of ReferenceGrantSpec

The Planton spec flattens the upstream `ReferenceGrantSpec` after the standard
namespaced envelope (`target_cluster`, `namespace`):

- `from` -- the trusted sources, each a `(group, kind, namespace)` tuple
  (**required**, 1 to 16).
- `to` -- the referenceable targets in this grant's namespace, each a
  `(group, kind)` pair with an optional `name` (**required**, 1 to 16).

There are no parent refs, backend refs, hostnames, matches, or filters -- a
ReferenceGrant carries no routing, only authorization. It is the only Gateway API
family member with no runtime dependency on a Gateway or route existing.

## The from / to Model

- **`from[]`** names *who* may reference: a source `kind` (e.g. `HTTPRoute`,
  `Gateway`) in a source `namespace`, under an API `group`. `from` entries have no
  `name` -- a grant trusts *all* resources of that kind in that namespace, by
  design (naming individual sources would be brittle and is not how the upstream
  models trust).
- **`to[]`** names *what* may be referenced in this grant's namespace: a target
  `kind` (e.g. `Service`, `Secret`) under an API `group`, optionally narrowed to a
  single `name`. Omitting `name` grants access to all resources of that group/kind.
- **Entries combine with OR.** Each `from` entry is an additional trusted source;
  each `to` entry is an additional referenceable target. One grant can express a
  small matrix of trust relationships.
- **The empty group `""`** infers the Kubernetes core API group -- this is how
  `Secret` and `Service` (core kinds) are expressed.

## Validation Rules

Every upstream kubebuilder marker is translated to `buf.validate`:

- `from`: `min_items: 1`, `max_items: 16`.
- `to`: `min_items: 1`, `max_items: 16`.
- `group` (from & to): `Group` pattern -- empty (core group) or an RFC 1123
  subdomain, max 253. Deliberately **not** a `required` field: upstream marks the
  key required but its `Group` type explicitly allows the empty value, so requiring
  it would wrongly reject the legitimate core-group `""`.
- `kind` (from & to): required, `Kind` pattern (1-63, `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`).
- `from.namespace`: required, `Namespace` pattern (RFC 1123 label, 1-63).
- `to.name`: optional, `ObjectName` length bound (1-253); upstream has no character
  pattern on `ObjectName`, so none is invented.

There are **no message-level CEL rules** -- the upstream `ReferenceGrant` Go type
carries no `XValidation` and no union/discriminator fields, so unlike the route
kinds, this spec needs none.

## Why ReferenceGrant Is a First-Class Planton Component

Without it, every production Gateway deployment that spans namespaces is forced
back to a raw `KubernetesManifest` to authorize its own cross-namespace references:
no proto validation, no typed SDKs, no FK wiring, no InfraChart composability, no
UI wizards. Worse, without ReferenceGrant the other six Gateway API components can
only operate within a single namespace -- a severe limitation for any realistic
ingress topology where Gateways, certs, and backends live apart. Including it
completes the family and unlocks true multi-namespace ingress.

## Design Decisions

- **Standard channel, `v1`.** ReferenceGrant is served in the standard channel as
  `gateway.networking.k8s.io/v1`. The typed crd2pulumi resource emits the `v1`
  apiVersion (aliased from `v1beta1`).
- **Component-local from / to types.** Unlike the routes, ReferenceGrant does not
  reuse the shared `gateway_api.proto` reference types -- its entries are trust
  assertions about kinds, structurally unrelated to `ParentReference` /
  `BackendRef`. They are modeled as component-prefixed local messages.
- **`group` / `kind` are plain `string`, `to.name` is `optional string` (DD-008,
  dont-do #2).** For `group`/`kind` the empty/zero value is either meaningful
  (`""` = core group) or simply invalid (empty kind), with no "unset vs empty"
  distinction to preserve, so a pointer would add noise. For `to.name`, absence
  (meaning "all resources") is genuinely distinct from any concrete name, so it is
  `optional`.
- **No baked-in Planton defaults on upstream fields (dont-do #4).** Upstream
  defaults are documented in comments and left to the controller.
- **Typed crd2pulumi IaC.** `gatewayv1.NewReferenceGrant`; the spec mapping (the
  two from/to lists) lives in `references.go`. No matches/filters/refs files.
- **e2e `deferred` (DD-007).** Although ReferenceGrant has no runtime dependency on
  a Gateway/route existing (it needs only the CRDs), the current E2E harness does
  not provision the `KubernetesGatewayApiCrds` prerequisite on the bare kind
  cluster, so it is deferred like the rest of the family.

## No Status Subresource

Upstream `ReferenceGrant` deliberately omits a status subresource (the API authors
found the design hard to settle and left it for the future). Planton therefore
exports only the identifying coordinates (`reference_grant_name`, `namespace`) in
`stack_outputs.proto` -- there is no controller-reconciled status to surface. This
is faithful: inventing a status would diverge from the upstream contract.

## Composing in Infra Charts

`KubernetesReferenceGrant` is designed as a LEGO block for Infra Charts. Two
mechanisms wire it into the dependency DAG (see project decision DD-009):

1. **Data dependencies use `StringValueOrRef` (`valueFrom`).** `namespace` is a
   `StringValueOrRef` with `default_kind = KubernetesNamespace`, so the grant's own
   ("to") namespace can reference a namespace output and the platform builds the DAG
   edge automatically.
2. **Topology dependencies use `metadata.relationships`.** The `from`/`to` entries
   are trust assertions about kinds, not instance pointers, so most carry no DAG
   edge at all. The single exception is `from[].namespace` (a source namespace);
   when it is Planton-managed, the author expresses that edge explicitly:

```yaml
metadata:
  name: "{{ values.env }}-allow-frontend"
  relationships:
    - kind: KubernetesNamespace          # the source ("from") namespace, if Planton-managed
      name: "{{ values.frontend_ns }}"
      type: uses
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: "{{ values.backend_ns }}"
      fieldPath: spec.name
  from:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      namespace: "{{ values.frontend_ns }}"
  to:
    - group: ""
      kind: Service
```

Critically, the grant is a **leaf** the *consumer* depends on, not the other way
round. In the full ingress stack the relationship edges live on the Gateway/Route:

```
KubernetesNamespace (backend-ns) ── valueFrom ──> KubernetesReferenceGrant
                                                     ▲
KubernetesHttpRoute (frontend-ns) ── relationships depends_on ─┘
  (its backend_refs cross into backend-ns, which the grant authorizes)
```

The cert path composes the same way: a `KubernetesReferenceGrant` in the
cert-manager namespace (`from: Gateway`, `to: Secret`) authorizes a
`KubernetesGateway` in `istio-ingress` to reference the TLS Secret produced by a
`KubernetesCertificate`. Data edges use `valueFrom`; topology edges use
`metadata.relationships`. Both are required for the platform to build the full DAG.

## Controller Landscape

ReferenceGrant is honored by every Gateway API controller that implements
cross-namespace references (Istio, Envoy Gateway, and others). A controller MUST
NOT permit a cross-namespace reference with no grant, and MUST revoke access when a
grant is removed. The proto carries no controller-specific behavior.

## Common Pitfalls

- **The grant lives in the TARGET namespace, not the source.** A ReferenceGrant
  goes in the namespace being referenced *into* (where the Secret/Service lives),
  and lists the source under `from`. Putting it in the source namespace does
  nothing.
- **`from` entries have no name -- that is intentional.** Trust is granted to all
  resources of a kind in a namespace; narrow the *target* with `to[].name` if you
  must, not the source.
- **Use `group: ""` for core kinds.** `Secret` and `Service` are in the core API
  group; express that as the empty string, not `core` or `v1`.
- **The grant alone does not order deployment.** It is a leaf; the consuming
  Gateway/Route must declare the `depends_on`/`uses` relationship so the platform
  sequences the grant first (DD-009).

## Conclusion

`KubernetesReferenceGrant` completes the Planton Gateway API ingress layer at 100%
standard-channel fidelity, unlocking multi-namespace topologies for the entire
family while faithfully preserving the upstream trust model -- including its
deliberate absence of a status subresource.

## References

- [Gateway API ReferenceGrant](https://gateway-api.sigs.k8s.io/api-types/referencegrant/)
- Upstream types: `kubernetes-sigs/gateway-api` `apis/v1/referencegrant_types.go` (v1.5.1)
- Sibling components: `KubernetesGateway`, `KubernetesHttpRoute`, `KubernetesGrpcRoute`, `KubernetesTlsRoute`, `KubernetesTcpRoute`
- Project decision: DD-009 (infra-chart composability, plain refs + relationships)
