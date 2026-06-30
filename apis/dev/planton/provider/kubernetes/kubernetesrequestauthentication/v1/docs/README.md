# KubernetesRequestAuthentication -- Design & Research

This document captures the deployment landscape, the modeling rationale, and the
fidelity decisions behind the `KubernetesRequestAuthentication` Planton component
(kind 864). It models repeated nested rule lists, JWT semantics, and the
`selector` / `target_refs` attachment fork.

## 1. What RequestAuthentication is

`RequestAuthentication` is an Istio security CRD (`security.istio.io`, served at both
`v1` and `v1beta1` -- identical schema) that defines **which JSON Web Tokens are
accepted** at the workloads it selects. It answers one question: *is this request's
token valid, and what identity does it carry?*

It is distinct from `PeerAuthentication` (transport-level service-to-service mTLS)
and `AuthorizationPolicy` (allow/deny decisions). Crucially, RequestAuthentication
**does not require** a token: a request with no token passes it untouched. To force
authentication, pair it with an AuthorizationPolicy that requires `requestPrincipals`.

## 2. Source of truth and version

Translated proto-to-proto from the upstream `istio.io/api` clone pinned to tag
**1.26.8** (`security/v1beta1/request_authentication.proto`). The local clone is
authoritative; no specs are pulled from the internet. The crd2pulumi typed
SDK and the CRDs installed by `KubernetesIstioBaseCrds` are likewise generated from
Istio `release-1.26`, so the proto, the typed Pulumi resource, and the cluster CRD
all agree on the schema.

This component uses a repeated `jwt_rules` list with nested repeated `from_headers`
/ `output_claim_to_headers`, the `selector` / `target_refs` attachment fork, and a
`google.protobuf.Duration` field (`timeout`).

## 3. Planton spec shape (fidelity decisions)

The Planton spec flattens the upstream `RequestAuthentication` fields directly after
the namespaced envelope (`target_cluster`, `namespace`). There is no nested
`request_authentication` sub-message. Fields are renumbered sequentially from
1 -- the Planton proto is its own wire contract and does not preserve upstream's
field-number gaps (e.g. upstream `jwks` is field 10, `timeout` 13).

| Upstream | Planton | Notes |
|----------|---------|-------|
| `selector` (`type.v1beta1.WorkloadSelector`) | `selector` (`KubernetesIstioApiWorkloadSelector`) | Shared type in `istio_api.proto`. |
| `targetRefs` (`type.v1beta1.PolicyTargetReference`) | `target_refs` (`KubernetesIstioApiPolicyTargetReference`) | New shared type in `istio_api.proto` (this component is its first consumer; AuthorizationPolicy reuses it). |
| `targetRef` (singular, `$hide_from_docs`) | omitted | Upstream hides it; only the public `targetRefs` list is modeled. |
| `jwtRules` (`JWTRule`) | `jwt_rules` (`KubernetesRequestAuthenticationJwtRule`) | Repeated; nested messages below. |
| `JWTHeader` | `KubernetesRequestAuthenticationJwtHeader` | `name`, `prefix`. |
| `ClaimToHeader` | `KubernetesRequestAuthenticationClaimToHeader` | `header`, `claim`. |
| `JWTRule.timeout` (`google.protobuf.Duration`) | `timeout` (`string`) | Durations are modeled as strings. |

### selector vs target_refs (attachment fork)

Upstream permits **at most one** of `selector`, `targetRef`, `targetRefs`; all
omitted means "match every workload in the namespace", which is valid. So this is an
"at most one" rule, not an "exactly one" biconditional. Modeled as a message-level CEL
(`!(has(this.selector) && size(this.target_refs) > 0)`) over the two public fields.
The hidden singular `targetRef` is not modeled.

### Durations are strings

`JWTRule.timeout` is a `google.protobuf.Duration` upstream. It is modeled as an
`optional string` validated by the upstream CEL `duration(this) >= duration('1ms')`
(cel-go's `duration()` is available under protovalidate). A duration *regex* was
deliberately NOT used: it would reject valid `google.protobuf.Duration` strings such
as `1.5s` or `2h45m`. Upstream's 5s default is not baked into Planton -- it flows
through from istiod.

### CEL portability

Upstream CRD `XValidation` rules use the Kubernetes apiserver CEL extension library
(`url()`, `oneof()`), which `buf`/`protovalidate`/`cel-go` do not provide. Each rule
is re-expressed with portable CEL primitives that preserve the validated outcome:

| Upstream intent | Planton CEL |
|-----------------|-------------|
| `oneof(selector, targetRef, targetRefs)` | `!(has(this.selector) && size(this.target_refs) > 0)` |
| `oneof(jwksUri, jwks)` | `!(has(this.jwks_uri) && has(this.jwks))` |
| `url(self).getScheme() in ['http','https']` | `this == '' \|\| this.startsWith('http://') \|\| this.startsWith('https://')` |
| `duration(self) >= duration('1ms')` | same (`duration()` is supported) |

### Enum case / external standard

No closed-enum fields here (unlike PeerAuthentication's `mtls.mode`); the only
constrained scalars are issuer/header patterns and the jwks_uri scheme. The shared
`KubernetesIstioApiPolicyTargetReference` carries upstream's group/kind/name patterns
and the "cross-namespace not supported" CEL (`namespace` must be empty in 1.26).

### Validation rules (each has accept + reject coverage in `spec_test.go`)

| Rule | Origin |
|------|--------|
| `request_authentication.selector_xor_target_refs` (message-level) | Upstream selector/targetRefs XValidation. |
| `request_authentication_jwt_rule.jwks_uri_xor_jwks` (message-level) | Upstream jwks_uri/jwks XValidation. |
| `request_authentication_jwt_rule.jwks_uri_scheme` | Upstream jwks_uri URL-scheme XValidation. |
| `request_authentication_jwt_rule.timeout_min` | Upstream timeout `>= 1ms` XValidation. |
| `issuer` non-empty, `from_headers[].name` non-empty, `output_claim_to_headers[].{header,claim}` (header pattern `^[-_A-Za-z0-9]+$`) | Upstream field constraints. |
| `target_refs` max 16; `KubernetesIstioApiPolicyTargetReference` group/kind/name patterns + no-cross-namespace | Upstream PolicyTargetReference + kubebuilder MaxItems. |
| selector `match_labels` non-empty keys / no-wildcard keys / no-wildcard values | Shared `istio_api.proto` WorkloadSelector. |

## 4. Composability

- **`namespace`** is the one true foreign key: `StringValueOrRef` ->
  `KubernetesNamespace` (`spec.name`). Literal or `valueFrom`; creates a real DAG
  edge.
- **`selector.match_labels`** and **`target_refs`** are plain runtime references, NOT
  foreign keys. istiod resolves them against the cluster at runtime; neither creates
  an automatic DAG edge. Wrapping them in `StringValueOrRef` would break upstream
  fidelity (a label map / a multi-field reference, not a scalar) and distort the
  typed CRD shape. An infra-chart author who needs ordering must declare it on
  `metadata.relationships` (`uses`/`depends_on` -> KubernetesDeployment /
  KubernetesGateway / KubernetesService / KubernetesServiceEntry). See the
  component README's "Composing in Infra Charts" section.

## 5. IaC implementation

Both engines are feature-equal and emit the same `security.istio.io/v1`
`RequestAuthentication`:

- **Pulumi** uses the typed crd2pulumi resource
  `istiosecurityv1.NewRequestAuthentication`. The `Spec` field is a
  `PtrInput` satisfied by the `RequestAuthenticationSpecArgs` **value** (the
  `*Ptr()` wrapper marshals to the wrong element type and panics at apply).
  `jwt_rules`, nested `from_headers` /
  `output_claim_to_headers`, `selector`, and `target_refs` are only attached when
  present; optional scalar fields use the proto3 `optional` pointer to distinguish
  unset from empty.
- **Terraform** uses `kubernetes_manifest` with a fully-typed `variable "spec"` and a
  null-pruned `locals.tf`. Each list element is assembled with `merge()` so only the
  fields the user set reach the manifest; nested optionals are read through `?:`
  guards (HCL `&&` does not short-circuit). Snake_case spec fields map to the CRD's
  camelCase (`jwksUri`, `fromHeaders`, `outputClaimToHeaders`, `targetRefs`).

## 6. E2E

- **Prerequisite**: `KubernetesIstioBaseCrds` (kind 868), declared in the registry
  (`prerequisites: [KubernetesIstioBaseCrds]`). The harness installs the istio/base
  CRDs before the scenario applies -- no full istiod is needed to prove the CR is
  accepted server-side and cleaned up on destroy.
- **Tier 4**, validated on both Pulumi and Terraform. Verification asserts the
  RequestAuthentication CR exists after apply and is gone after destroy
  (`requestauthentications.security.istio.io`), via the `ResourceExistenceVerifier`
  dispatched from `aa_e2e/verify/verifier.go`.

## 7. Future experience surface (kept open, not built here)

The spec is intentionally lossless so future Planton console experiences stay
possible without a schema change: an "identity providers" panel that lists accepted
issuers per namespace/workload, JWKS reachability checks, a "require authentication"
toggle that co-authors a paired AuthorizationPolicy, and claim-to-header mapping
editors all derive from these exact fields. The console integration itself is a
separate follow-up project.

## References

- [Istio RequestAuthentication reference](https://istio.io/latest/docs/reference/config/security/request_authentication/)
- [Istio JWT / end-user authentication task](https://istio.io/latest/docs/tasks/security/authentication/authn-policy/#end-user-authentication)
- Upstream proto: `istio.io/api` `security/v1beta1/request_authentication.proto` @ `1.26.8`
