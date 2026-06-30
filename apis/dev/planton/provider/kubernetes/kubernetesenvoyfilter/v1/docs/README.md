# KubernetesEnvoyFilter -- Design & Research

This document captures the deployment landscape, the modeling rationale, and the fidelity
decisions behind the `KubernetesEnvoyFilter` Planton component (kind 867). It is the deepest
Istio component by nesting (twelve component-local match/patch messages) and the only one
still served at `networking/v1alpha3`.

## 1. What EnvoyFilter is

`EnvoyFilter` is an Istio networking CRD (`networking.istio.io`, served **only** at
`v1alpha3` -- it has not graduated to `v1` like DestinationRule/ServiceEntry) that lets an
operator **patch the raw Envoy xDS configuration istiod generates** for selected proxies. Each
patch declares *where* to apply (`apply_to`), *which* object to match (`match`), and *how* to
change it (`patch`: an operation plus a free-form Envoy config fragment).

It is the lowest-level, highest-blast-radius surface Istio exposes. The patch body
(`config_patches[].patch.value`) is an arbitrary `google.protobuf.Struct` that istiod
proto-merges into generated config with no schema validation, so a malformed value can silently
break a workload's traffic. Istio itself positions EnvoyFilter as a last resort and is actively
graduating common uses (CORS, ext_authz, rate limiting) onto first-class typed APIs.

## 2. Source of truth and version

Translated proto-to-proto from the upstream `istio.io/api` clone pinned to tag **1.26.8**
(`networking/v1alpha3/envoy_filter.proto`). The local clone is authoritative; no specs are
pulled from the internet. The crd2pulumi typed SDK and the CRDs installed by
`KubernetesIstioBaseCrds` are likewise generated from Istio `release-1.26`, so the proto, the
typed Pulumi resource (package `networking/v1alpha3`), and the cluster CRD all agree.

**Version distinction (do not conflate):** the Planton kind's own API version is `v1`
(`cloud_resource_kind.proto` `version: v1`, package directory `kubernetesenvoyfilter/v1/`),
identical for every Planton kind. The **Istio CRD** is served at `networking.istio.io/v1alpha3`
-- the value emitted in the manifest's `apiVersion`. EnvoyFilter has no `v1` served version.

## 3. Planton spec shape (fidelity decisions)

The Planton spec flattens the upstream `EnvoyFilter` fields directly after the namespaced
envelope (`target_cluster`, `namespace`) -- no nested `envoy_filter` sub-message.
Fields are numbered sequentially from 1 (the Planton proto is its own wire contract; it does not
preserve upstream field-number gaps such as the upstream `targetRefs = 6` / `config_patches = 4`
ordering).

| Upstream | Planton | Notes |
|----------|---------|-------|
| `workloadSelector` (`networking.v1alpha3.WorkloadSelector`) | `workload_selector` (`KubernetesIstioApiNetworkingWorkloadSelector`) | Shared type -- same upstream message ServiceEntry uses; identical CRD constraints. |
| `targetRefs` (`type.v1beta1.PolicyTargetReference`) | `target_refs` (`KubernetesIstioApiPolicyTargetReference`) | Shared type -- same upstream message RequestAuthentication uses; max 16. |
| `configPatches` (`EnvoyConfigObjectPatch`) | `config_patches` (`KubernetesEnvoyFilterConfigPatch`) | 12 nested component-local messages; see below. |
| `priority` (`int32`) | `priority` (`optional int32`) | Unset => upstream default 0. |
| `EnvoyConfigObjectMatch.object_types` (proto `oneof`) | three sibling fields + at-most-one CEL | Flattened; see below. |
| `Patch.value` (`google.protobuf.Struct`) | `value` (`google.protobuf.Struct`) | Free-form, kept as Struct. |
| `ApplyTo`/`PatchContext`/`Operation`/`FilterClass`/`Action` enums | closed-set `optional string` | UPPERCASE string + `in`, not proto enums. |

### The two shared selectors reused (not redefined)

EnvoyFilter reuses both shared Istio types rather than redefining them:

- `KubernetesIstioApiNetworkingWorkloadSelector` (field `labels`) for `workload_selector`.
  EnvoyFilter's upstream `WorkloadSelector` is the **same** `networking/v1alpha3` message
  ServiceEntry uses, and its CRD applies the identical constraints (max 256 labels, value <= 63
  chars, no-wildcard values, no key rules), so the shared type is faithful without change.
- `KubernetesIstioApiPolicyTargetReference` for `target_refs` -- the **same**
  `type/v1beta1.PolicyTargetReference` RequestAuthentication uses (group/kind/name + the
  no-cross-namespace CEL). EnvoyFilter additionally caps the list at 16.

### The `object_types` oneof (modeled without a discriminator)

`EnvoyConfigObjectMatch` carries a proto `oneof object_types { listener | route_configuration |
cluster }`. The oneof is flattened to sibling message fields. It is modeled as three **optional
sibling fields with no `match_type` discriminator** -- the same shape the crd2pulumi SDK uses,
and the same choice made for structural attachment forks (selector vs target_refs,
workload_selector vs endpoints); discriminators are reserved for same-type value unions (e.g.
`KubernetesIstioApiStringMatch`'s EXACT/PREFIX/REGEX). Because flattening loses the proto oneof's
implicit at-most-one guarantee -- and the generated CRD does **not** re-encode it as an
XValidation -- the rule is restored as a message-level CEL. This preserves upstream semantics; it
is not extra validation.

### Enum case / external standard

`apply_to`, `match.context`, `route.action`, `patch.operation`, and `patch.filter_class` are
closed-set `optional string` fields validated by the protovalidate `string.in` rule with the
UPPERCASE upstream constants -- not proto enums. The zero-value sentinels (`INVALID`, and
`FilterClass.UNSPECIFIED`) are excluded from the sets: unset already means "unspecified",
selecting an explicit `INVALID` is never valid. `apply_to` keeps `BOOTSTRAP` for fidelity (the
CRD still accepts it) with an inline note that it is deprecated upstream. Leaving any enum unset
omits it from the CR, so istiod's own default applies.

### Deliberate omission

`ListenerMatch.port_name` is **not** surfaced: upstream marks it "Not implemented" /
`$hide_from_docs`, so it carries no behavior. Omitting a non-functional field (rather than
exposing a dead knob) keeps 100% of the *documented* surface.

### CEL portability

The upstream CRD declares exactly one `XValidation`
(`oneof(self.workloadSelector, self.targetRefs)`), which uses the apiserver-CEL `oneof()` helper
that `buf`/`protovalidate`/`cel-go` do not provide. It is re-expressed with a portable `has()`
guard. The object_types rule is added to restore the flattened proto oneof (above). No
istiod-runtime cross-field rules (e.g. `ROUTE_CONFIGURATION` allows only `MERGE`,
`HTTP_FILTER`/`NETWORK_FILTER` for `REPLACE`) are encoded -- those are istiod webhook semantics,
not CRD `XValidation`, and the fidelity contract is the CRD's validated surface.

| Upstream intent | Planton CEL | Origin |
|-----------------|-------------|--------|
| `oneof(workloadSelector, targetRefs)` | `!(has(this.workload_selector) && size(this.target_refs) > 0)` | CRD XValidation. |
| object_types at most one | three pairwise `!(has(a) && has(b))` clauses over listener/route_configuration/cluster | Restores the flattened proto `oneof`. |
| `targetRefs` <= 16 | `repeated.max_items = 16` | Upstream `+kubebuilder:validation:MaxItems`. |
| port numbers 1-65535 | `uint32 {gte:1, lte:65535}` on `port_number`/`destination_port` | Upstream port semantics. |
| selector value no-wildcard; target_ref no cross-namespace | inherited from the shared types | Upstream CRD CEL. |

### Validation rules (each has accept + reject coverage in `spec_test.go`)

| Rule | Origin |
|------|--------|
| `envoy_filter.workload_selector_xor_target_refs` (message-level) | Upstream selector/targetRefs XValidation. |
| `envoy_filter_match.object_types_at_most_one` (message-level) | Restores the flattened proto oneof. |
| `target_refs` max 16; PolicyTargetReference kind/name required + no cross-namespace | Upstream markers + shared type. |
| `apply_to` / `context` / `operation` / `filter_class` / route `action` closed sets | Upstream enums (BOOTSTRAP kept). |
| cluster `port_number` / listener `destination_port` 1-65535 | Upstream port semantics. |
| `workload_selector.labels` value no-wildcard | Networking WorkloadSelector CRD (shared type). |

## 4. Composability

- **`namespace`** is the one true foreign key: `StringValueOrRef` -> `KubernetesNamespace`
  (`spec.name`). Literal or `valueFrom`; creates a real DAG edge.
- **`workload_selector.labels`** and **`target_refs`** are plain runtime references, NOT foreign
  keys. istiod matches selectors against pod labels and resolves target refs against the cluster
  at runtime; neither creates an automatic DAG edge. An infra-chart author who needs ordering
  (e.g. an EnvoyFilter that patches a Gateway's listeners) declares it on
  `metadata.relationships` (`depends_on` -> KubernetesGateway / KubernetesDeployment /
  KubernetesService). See the component README's "Composing in Infra Charts" section.
- **Outputs** export the honest `envoy_filter_name` + `namespace`; an EnvoyFilter is a
  config-patch resource with no controller-reconciled status worth surfacing.

## 5. IaC implementation

Both engines are feature-equal and emit the same `networking.istio.io/v1alpha3` `EnvoyFilter`:

- **Pulumi** uses the typed crd2pulumi resource `istionetworkingv1alpha3.NewEnvoyFilter`.
  The `Spec` field is a `PtrInput` satisfied by the `EnvoyFilterSpecArgs` **value**
  (the `*Ptr()` wrapper marshals to the wrong element type and panics at apply).
  Per-nested-message builder helpers attach every
  block only when present; proto `uint32` ports are cast to the SDK's `int`. The free-form
  `patch.value` Struct is converted to a `pulumi.Map` by a small recursive `structToPulumiMap`
  helper (`map`->`pulumi.Map`, slice->`pulumi.Array`, scalars->typed inputs) -- the SDK types
  `value` as `pulumi.MapInput`, and the repo has no generic `pulumi.ToMap`, so the recursive
  converter is the robust path and preserves arbitrary nesting.
- **Terraform** uses `kubernetes_manifest` with a fully-typed `variable "spec"` and a deeply
  null-pruned `locals.tf`. Each list element and nested block is assembled with `merge()` and
  read through a `?:` guard (HCL `&&` does not short-circuit), so only the fields the user set
  reach the manifest. Snake_case spec fields map to the CRD's camelCase (`applyTo`,
  `routeConfiguration`, `filterChain`, `subFilter`, `proxyVersion`, `portNumber`,
  `transportProtocol`, `applicationProtocols`, `destinationPort`, `domainName`, `filterClass`,
  `targetRefs`, `workloadSelector`). The free-form `patch.value` is typed `optional(any)` and
  passes through unmodified.

## 6. E2E

- **Prerequisite**: `KubernetesIstioBaseCrds` (kind 868), declared in the registry
  (`prerequisites: [KubernetesIstioBaseCrds]`). The harness installs the istio/base CRDs before
  the scenario applies -- no full istiod is needed to prove the CR is accepted server-side and
  cleaned up on destroy.
- **Tier 4**, validated on both Pulumi and Terraform. Verification asserts the EnvoyFilter CR
  exists after apply and is gone after destroy (`envoyfilters.networking.istio.io`), via the
  `ResourceExistenceVerifier` dispatched from `aa_e2e/verify/verifier.go`.

## 7. Future experience surface (kept open, not built here)

The spec is intentionally lossless so future Planton console experiences stay possible without a
schema change. The product north star is to treat EnvoyFilter as a **guarded escape hatch with a
path off it**: recognize common patterns (gRPC-Web CORS, ext_authz, rate limiting) authored as
EnvoyFilters and offer first-class typed alternatives so users graduate off raw xDS patching --
exactly the direction Istio itself is taking. A "what does this patch?" view can resolve the
`workload_selector` / `target_refs` against matched workloads, and a safety analyzer can warn on
high-risk operations (`REPLACE` on a core filter, `BOOTSTRAP`). The console integration is a
separate follow-up project.

## References

- [Istio EnvoyFilter reference](https://istio.io/latest/docs/reference/config/networking/envoy-filter/)
- Upstream proto: `istio.io/api` `networking/v1alpha3/envoy_filter.proto` @ `1.26.8`
