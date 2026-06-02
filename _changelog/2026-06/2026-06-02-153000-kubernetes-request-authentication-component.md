# KubernetesRequestAuthentication deployment component

**Date**: June 2, 2026
**Type**: Feature
**Components**: API Definitions, Kubernetes Provider, IaC Modules (Pulumi + Terraform), E2E

## Summary

Forged `KubernetesRequestAuthentication` (kind 864), the second typed Istio API
deployment component, at 100% upstream fidelity to Istio 1.26.8. It provisions a
namespaced `security.istio.io/v1` `RequestAuthentication` -- the mesh policy that
defines which JSON Web Tokens (JWTs) are accepted on selected workloads, extracting
the authenticated identity for downstream authorization. Proven green on a live
`kind` cluster across both IaC engines on the first run.

## What's new

- **API**: `kubernetesrequestauthentication/v1` -- `spec.proto`, `api.proto`,
  `stack_input.proto`, `stack_outputs.proto`, plus generated stubs. The spec
  flattens the upstream fields after the namespaced envelope:
  - `namespace` -- `StringValueOrRef` foreign key to `KubernetesNamespace`.
  - `selector` -- reuses the shared `KubernetesIstioApiWorkloadSelector`.
  - `target_refs` -- repeated, reuses the new shared
    `KubernetesIstioApiPolicyTargetReference` (max 16). Mutually exclusive with
    `selector` via a message-level CEL (at most one attachment mechanism).
  - `jwt_rules` -- repeated `KubernetesRequestAuthenticationJwtRule` (issuer,
    audiences, `jwks_uri`/`jwks` (mutually exclusive), `from_headers`,
    `from_params`, `from_cookies`, `output_payload_to_header`,
    `forward_original_token`, `output_claim_to_headers`, `timeout`), with nested
    `KubernetesRequestAuthenticationJwtHeader` and
    `KubernetesRequestAuthenticationClaimToHeader` messages.
- **Shared type**: added `KubernetesIstioApiPolicyTargetReference` to
  `istio_api.proto` (group/kind/name with upstream patterns, plus the upstream
  "cross-namespace not supported" rule on `namespace`). RequestAuthentication is its
  first consumer; AuthorizationPolicy (kind 865) will reuse it.
- **IaC**: typed Pulumi module (`istiosecurityv1.NewRequestAuthentication`) and a
  null-pruned Terraform module (`kubernetes_manifest`), feature-equal across both
  engines. Each list element (jwt rules, from-headers, claim-to-headers,
  target-refs) is assembled with per-element null-pruning.
- **E2E**: tier-4 profile (green, pulumi + terraform). Declares
  `prerequisites: [KubernetesIstioBaseCrds]`; verification dispatches through the
  `istioApiKinds` map, asserting the `requestauthentications.security.istio.io` CR
  exists after apply and is gone after destroy.

## Validation

Full ladder, all green:

- Static: `make protos` (incl. Java gate), `make generate-cloud-resource-kind-map`,
  `go build`, `go vet`, `terraform fmt`/`validate`, bazel build + nogo.
- Codified: protovalidate spec tests (accept + reject for every CEL rule, including
  the selector/target_refs and jwks_uri/jwks exclusivity rules, the timeout duration
  minimum, the jwks_uri scheme, the claim-header pattern, and the shared
  PolicyTargetReference constraints).
- **Live E2E on a `kind` cluster** for both Pulumi and Terraform -- all eight
  lifecycle phases pass (dependency install of `KubernetesIstioBaseCrds` -> apply ->
  verify exists -> destroy -> verify gone). No runtime bugs: the T02 breadcrumbs
  (assign the typed `Spec` Args **value** not the `*Ptr()` wrapper; guard nested
  Terraform optionals with `?:` not `&&`) were applied up front.

## Notes

- **CEL portability**: upstream CRD `XValidation` rules use Kubernetes apiserver CEL
  extensions (`url()`, `oneof()`) that `protovalidate`/`cel-go` do not provide. Each
  was re-expressed with portable primitives that preserve the validated outcome
  (`!(has(a) && has(b))` for "at most one"; `startsWith('http://')` for the URL
  scheme). `duration()` is supported, so the timeout rule mirrors upstream exactly.
- **Duration fidelity**: `JWTRule.timeout` is modeled as `optional string` validated
  by `duration(this) >= duration('1ms')` (DD-002), NOT the Gateway API duration
  regex -- which would wrongly reject valid `google.protobuf.Duration` values like
  `1.5s` or `2h45m`.
- The hidden upstream singular `targetRef` (field 3, `$hide_from_docs`) is
  intentionally not modeled; only the public `target_refs` list is surfaced.
