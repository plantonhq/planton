# KubernetesAuthorizationPolicy -- Design & Research

This document captures the deployment landscape, the modeling rationale, and the
fidelity decisions behind the `KubernetesAuthorizationPolicy` OpenMCF component
(kind 865). It is the fifth typed Istio API component and completes the mesh security
triad alongside `KubernetesPeerAuthentication` (863) and
`KubernetesRequestAuthentication` (864). It reuses the shared selector / target-ref
types those siblings introduced and adds the family's richest rule model (sources,
operations, conditions) plus a closed-set `action` enum.

## 1. What AuthorizationPolicy is

`AuthorizationPolicy` is an Istio security CRD (`security.istio.io`, served at both
`v1` and `v1beta1` -- identical schema) that enforces **access control** on the
workloads it selects. It answers: *should this request be allowed, denied, or
audited?*

Istio evaluates the actions in a fixed order -- CUSTOM, then DENY, then ALLOW:

1. If a CUSTOM policy matches, the external authorizer's decision is honored (and can
   only further restrict, never bypass, ALLOW/DENY).
2. If a DENY policy matches, the request is denied.
3. If there are no ALLOW policies for the workload, the request is allowed.
4. If an ALLOW policy matches, the request is allowed; otherwise it is denied.

AUDIT is orthogonal: it flags matching requests for logging by an audit plugin without
changing the allow/deny outcome.

It is distinct from `RequestAuthentication` (validates JWTs but does not require one)
and `PeerAuthentication` (transport-level mTLS). The common "require login" pattern is
a RequestAuthentication that validates tokens plus an AuthorizationPolicy that requires
`request_principals`.

## 2. Source of truth and version

Translated proto-to-proto from the upstream `istio.io/api` clone pinned to tag
**1.26.8** (`security/v1beta1/authorization_policy.proto`). Per DD-001, the local clone
is authoritative; no specs are pulled from the internet. The crd2pulumi typed SDK and
the CRDs installed by `KubernetesIstioBaseCrds` are likewise on Istio `release-1.26`, so
the proto, the typed Pulumi resource, and the cluster CRD all agree on the schema.

## 3. OpenMCF spec shape (fidelity decisions)

The OpenMCF spec flattens the upstream `AuthorizationPolicy` fields directly after the
namespaced envelope (`target_cluster`, `namespace`), per DD-002. There is no nested
`authorization_policy` sub-message. Fields are renumbered sequentially from 1 -- the
OpenMCF proto is its own wire contract and does not preserve upstream's field-number
gaps (e.g. upstream `Source.service_accounts` is field 11).

| Upstream | OpenMCF | Notes |
|----------|---------|-------|
| `selector` (`type.v1beta1.WorkloadSelector`) | `selector` (`KubernetesIstioApiWorkloadSelector`) | Shared type in `istio_api.proto` (field `match_labels`). |
| `targetRefs` (`type.v1beta1.PolicyTargetReference`) | `target_refs` (`KubernetesIstioApiPolicyTargetReference`) | Shared type (introduced by RequestAuthentication, reused here). Max 16. |
| `targetRef` (singular, `$hide_from_docs`) | omitted | Upstream hides it; only the public `targetRefs` list is modeled (DD-001). |
| `rules` (`Rule`) | `rules` (`KubernetesAuthorizationPolicyRule`) | Repeated, max 512; nested messages below. |
| `Rule.From` | `KubernetesAuthorizationPolicyRuleFrom` | Wrapper `{ source }` (preserves CRD JSON `from: [{source}]`). Max 512. |
| `Rule.To` | `KubernetesAuthorizationPolicyRuleTo` | Wrapper `{ operation }`. |
| `Source` | `KubernetesAuthorizationPolicySource` | 12 identity/IP match lists + exclusivity CEL. |
| `Operation` | `KubernetesAuthorizationPolicyOperation` | 8 host/port/method/path match lists. |
| `Condition` | `KubernetesAuthorizationPolicyCondition` | `key` (required), `values`, `not_values`. |
| `action` (`Action` enum) | `action` (`string`) | DD-008 closed set: ALLOW/DENY/AUDIT/CUSTOM. |
| `action_detail` oneof -> `provider` (`ExtensionProvider`) | `provider` (`KubernetesAuthorizationPolicyExtensionProvider`) | Single-member oneof modeled as a plain optional message (no discriminator). |

### selector vs target_refs (attachment fork)

Upstream permits **at most one** of `selector`, `targetRef`, `targetRefs`; all omitted
means "match every workload in the namespace", which is valid. So this is an "at most
one" rule. Modeled as a message-level CEL
(`!(has(this.selector) && size(this.target_refs) > 0)`) over the two public fields. The
hidden singular `targetRef` is not modeled.

### The from/to wrappers

Upstream nests `source` inside a `From` message and `operation` inside a `To` message;
the CRD JSON is `from: [{source: {...}}]` / `to: [{operation: {...}}]`. To preserve that
shape (and match the crd2pulumi typed SDK exactly), OpenMCF keeps the wrapper messages
`KubernetesAuthorizationPolicyRuleFrom{source}` and `...RuleTo{operation}` rather than
flattening to a bare repeated Source/Operation.

### action as a closed string set (DD-008)

`Action` is an upstream proto enum (ALLOW/DENY/AUDIT/CUSTOM). Per DD-008 it is modeled
as an `optional string` with a `string.in` rule (UPPERCASE -- external standard
exception), the same form ServiceEntry uses for `location`/`resolution`. It is left
optional: unset inherits the upstream default ALLOW (pitfall #5 -- no OpenMCF default).

### provider as a single-member oneof

Upstream's `action_detail` oneof has exactly one member, `provider`. A discriminator
(DD-004) is unnecessary for a degenerate one-member union (discriminators are reserved
for value-type unions like StringMatch -- the EnvoyFilter `object_types` precedent), so
`provider` is a plain optional message. The "provider only with CUSTOM" coupling is an
istiod runtime check, not a CRD `XValidation`, so it is documented rather than enforced
(matching the validated surface; the EnvoyFilter T05 finding).

### CEL portability

Upstream CRD `XValidation` rules use the Kubernetes apiserver CEL extension library
(`oneof()`, `has()` on lists), which `buf`/`protovalidate`/`cel-go` do not provide. Each
rule is re-expressed with portable CEL primitives that preserve the validated outcome:

| Upstream intent | OpenMCF CEL |
|-----------------|-------------|
| `oneof(selector, targetRef, targetRefs)` | `!(has(this.selector) && size(this.target_refs) > 0)` |
| `Source` serviceAccounts exclusivity (`has()` on lists) | `(size(this.service_accounts) > 0 \|\| size(this.not_service_accounts) > 0) ? (size(this.principals) == 0 && size(this.not_principals) == 0 && size(this.namespaces) == 0 && size(this.not_namespaces) == 0) : true` |

### Validation rules (each has accept + reject coverage in `spec_test.go`)

| Rule | Origin |
|------|--------|
| `authorization_policy.selector_xor_target_refs` (message-level) | Upstream selector/targetRefs XValidation. |
| `authorization_policy_source.service_accounts_exclusive` (message-level) | Upstream Source serviceAccounts XValidation. |
| `action in [ALLOW, DENY, AUDIT, CUSTOM]` | Upstream Action enum (DD-008). |
| `target_refs` max 16; `KubernetesIstioApiPolicyTargetReference` group/kind/name patterns + no-cross-namespace | Upstream PolicyTargetReference + kubebuilder MaxItems. |
| `rules` max 512; `Rule.from` max 512 | Upstream kubebuilder MaxItems. |
| `Source.service_accounts` / `not_service_accounts` max 16, item max 320 chars | Upstream kubebuilder MaxItems + list-value MaxLength. |
| `Condition.key` non-empty (required) | Upstream `field_behavior=REQUIRED`. |
| `ExtensionProvider.name` non-empty | An extension provider must name a MeshConfig provider. |
| selector `match_labels` non-empty keys / no-wildcard keys / no-wildcard values | Shared `istio_api.proto` WorkloadSelector. |

The upstream doc note "a Condition must set at least one of `values`/`not_values`" is
not a declared CRD `XValidation`, so it is intentionally not enforced (DD-001: match the
validated surface; enforcing it would reject configs the CRD accepts).

## 4. Composability (DD-009)

- **`namespace`** is the one true foreign key: `StringValueOrRef` ->
  `KubernetesNamespace` (`spec.name`). Literal or `valueFrom`; creates a real DAG edge.
- **`selector.match_labels`** and **`target_refs`** are plain runtime references, NOT
  foreign keys. istiod resolves them against the cluster at runtime; neither creates an
  automatic DAG edge. Wrapping them in `StringValueOrRef` would break upstream fidelity
  (a label map / a multi-field reference, not a scalar) and distort the typed CRD shape.
  An infra-chart author who needs ordering must declare it on `metadata.relationships`
  (`uses`/`depends_on` -> KubernetesDeployment / KubernetesGateway / KubernetesService /
  KubernetesServiceEntry). See the component README's "Composing in Infra Charts"
  section.

## 5. IaC implementation

Both engines are feature-equal and emit the same `security.istio.io/v1`
`AuthorizationPolicy`:

- **Pulumi** uses the typed crd2pulumi resource `istiosecurityv1.NewAuthorizationPolicy`
  (DD-005). The `Spec` field is a `PtrInput` satisfied by the
  `AuthorizationPolicySpecArgs` **value** (the `*Ptr()` wrapper marshals to the wrong
  element type and panics at apply -- a bug the PeerAuthentication forge caught live).
  `selector`, `target_refs`, `rules` (with the `from.source` / `to.operation` / `when`
  builders), `action`, and `provider` are only attached when present; `action` uses the
  proto3 `optional` pointer to distinguish unset from empty.
- **Terraform** uses `kubernetes_manifest` with a fully-typed `variable "spec"` and a
  deeply null-pruned `locals.tf`. Each list element is assembled with `merge()` so only
  the fields the user set reach the manifest; nested optionals are read through `?:`
  guards (HCL `&&` does not short-circuit). Snake_case spec fields map to the CRD's
  camelCase (`targetRefs`, `requestPrincipals`, `serviceAccounts`, `notServiceAccounts`,
  `ipBlocks`, `remoteIpBlocks`, `notValues`, ...).

## 6. E2E

- **Prerequisite**: `KubernetesIstioBaseCrds` (kind 868), declared in the registry
  (`prerequisites: [KubernetesIstioBaseCrds]`). The harness installs the istio/base CRDs
  before the scenario applies -- no full istiod is needed to prove the CR is accepted
  server-side and cleaned up on destroy.
- **Tier 4**, validated on both Pulumi and Terraform. Verification asserts the
  AuthorizationPolicy CR exists after apply and is gone after destroy
  (`authorizationpolicies.security.istio.io`), via the `ResourceExistenceVerifier`
  dispatched from `aa_e2e/verify/verifier.go`.

## 7. Future experience surface (kept open, not built here)

The spec is intentionally lossless so future Planton console experiences stay possible
without a schema change: a visual rule builder (source identities x operations x
conditions), an action-precedence simulator that shows how CUSTOM/DENY/ALLOW policies
combine on a workload, a "require login" toggle that co-authors a paired
RequestAuthentication + ALLOW-on-request_principals policy, and an ext-authz provider
picker sourced from MeshConfig all derive from these exact fields. The console
integration itself is a separate follow-up project.

## References

- [Istio AuthorizationPolicy reference](https://istio.io/latest/docs/reference/config/security/authorization-policy/)
- [Istio authorization concepts](https://istio.io/latest/docs/concepts/security/#authorization)
- Upstream proto: `istio.io/api` `security/v1beta1/authorization_policy.proto` @ `1.26.8`
