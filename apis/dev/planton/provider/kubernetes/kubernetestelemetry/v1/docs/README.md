# KubernetesTelemetry -- Design & Research

This document captures the modeling rationale and fidelity decisions behind the
`KubernetesTelemetry` Planton component (kind 866). Telemetry is the mesh's observability
configuration: it controls how traces, metrics, and access logs are generated for the
workloads it selects. It is also the one Istio component whose Pulumi implementation cannot
use the typed crd2pulumi SDK -- see section 5.

## 1. What Telemetry is

`Telemetry` is an Istio CRD (`telemetry.istio.io`, served at `v1`; storage `v1alpha1`) that
tunes signal generation for selected workloads: trace sampling/tags, metric dimensions and
toggles, and access-log providers/filters. Configuration is hierarchical -- a
selector-less resource in the mesh root namespace is the mesh default; namespace- and
workload-scoped resources override it. Providers referenced here are extension providers
declared in the mesh's `MeshConfig`; Telemetry only selects and tailors them.

## 2. Source of truth and version

Translated proto-to-proto from the upstream `istio.io/api` clone pinned to tag **1.26.8**
(`telemetry/v1alpha1/telemetry.proto`; the `v1` CRD is a type alias). The local clone is
authoritative. The CRDs installed by `KubernetesIstioBaseCrds` are generated from Istio
`release-1.26`, so the proto and the cluster CRD agree on the schema.

## 3. Modeling decisions

### Envelope and shared types

Upstream spec fields are flattened after the Planton namespaced envelope (`target_cluster`,
`namespace`); there is no nested `telemetry` sub-message. `namespace` is the one foreign key
(`StringValueOrRef` -> `KubernetesNamespace`). The `selector` reuses the shared **policy**
selector `KubernetesIstioApiWorkloadSelector` (field `match_labels`, the
`istio.type.v1beta1.WorkloadSelector` upstream uses), and `target_refs` reuses
`KubernetesIstioApiPolicyTargetReference` (max 16) -- the same shared types as the security
components.

### `selector` vs `target_refs` -- "at most one"

Upstream's `oneof(self.selector, self.targetRef, self.targetRefs)` XValidation uses the
apiserver-CEL `oneof()` helper (not available under protovalidate/cel-go) and includes the
hidden singular `targetRef`. We omit `targetRef` (see below) and re-express the union as a
portable message CEL: `!(has(this.selector) && size(this.target_refs) > 0)`.

### Unions modeled as optional siblings + "at most one" CEL (shape-based)

Telemetry has two distinctly-named unions, both modeled as optional sibling fields with a
message-level "at most one" CEL (1:1 with the CRD JSON `oneOf`; no invented discriminator):

| Union | Planton CEL (id) |
|-------|------------------|
| `CustomTag {literal, environment, header}` | `telemetry_custom_tag.at_most_one_source` |
| `MetricSelector {metric, custom_metric}` | `telemetry_metric_selector.at_most_one_metric` |

### `TagOverride` operation/value coupling -- faithful `has()`-guarded CEL

Upstream's two `TagOverride` XValidations are `((has(self.operation) ? self.operation : "")
== "UPSERT") ? (self.value != "") : true` and the REMOVE analogue. Crucially, an **unset**
operation is treated as neither UPSERT nor REMOVE (no value constraint). To preserve that
exactly, `operation` and `value` are both `optional` (proto3 presence), and the rules are
re-expressed with `has()` guards:

- `!(has(this.operation) && this.operation == 'UPSERT') || (has(this.value) && this.value != '')`
- `!(has(this.operation) && this.operation == 'REMOVE') || !has(this.value)`

### Durations, wrappers, closed-set strings

`reporting_interval` is an `optional string` validated `this == '' || duration(this) >=
duration('1ms')`, mirroring the upstream CRD XValidation. All `google.protobuf.DoubleValue`
/ `BoolValue` wrappers become the matching `optional` scalar (`random_sampling_percentage`
-> `optional double` with the upstream 0-100 bounds; `disable_span_reporting`,
`enable_istio_tags`, metrics/access-logging `disabled` -> `optional bool`). `WorkloadMode`,
`IstioMetric`, and `Operation` are `optional string` validated by `string.in [...]` against
the UPPERCASE upstream constants (not proto enums).

### Deprecated / hidden fields

The hidden singular `targetRef` (`$hide_from_docs`) is omitted -- it is the legacy form of
`target_refs`. `Tracing.use_request_id_for_trace_sampling` is `$hide_from_docs` but **not**
deprecated -- it is a functional advanced knob the CRD accepts -- so it is **kept** (as
`optional bool`) for full fidelity, with a comment marking it advanced/hidden. (Hidden-only
fields are retained; only deprecated-and-hidden fields are dropped.)

## 4. The free-form access-log filter

`AccessLogging.Filter.expression` is a CEL string the CRD does not schema-validate; it is
carried as a plain `string` (no Planton validation), faithful to upstream.

## 5. IaC implementation

Both engines emit the same `telemetry.istio.io/v1` `Telemetry`.

### Pulumi uses an untyped CustomResource (the one Istio exception) -- and why

Every other typed Istio component uses its crd2pulumi-generated `New<Kind>` resource.
Telemetry **cannot**, because of a concrete crd2pulumi limitation: the CRD's
`spec.tracing[].customTags` is a map whose values are nested objects with a `oneOf` over
`{literal, environment, header}`. crd2pulumi (confirmed through v1.6.0 by regenerating the
SDK) degrades that map to `map[string]map[string]string`, which **structurally cannot
carry** the nested `{literal: {value: "..."}}` object -- and it generates no `CustomTag`
struct at all. Using the typed `NewTelemetry` would make a valid, upstream-supported custom
tag impossible to express (silent data loss), violating 100% fidelity. Mixing a typed
resource with one untyped field is not possible (the `Spec` and `customTags` fields are
strongly typed), so the whole resource is built via `apiextensions.NewCustomResource` with
the `spec` assembled as a map **from the strongly-typed proto getters** -- the input side
stays type-safe; only the Pulumi resource wrapper is generic, and the nested `customTags`
shape is preserved exactly. (`tagOverrides` is also degraded to `map[string]map[string]string`
upstream, but there it is faithful since `{operation, value}` are both strings -- it is not
the reason for the untyped resource.) If a future crd2pulumi types object-valued
`additionalProperties` maps, Telemetry can move to the typed SDK like its siblings.

### Terraform

`kubernetes_manifest` with a fully-typed `variable "spec"` and a `locals.tf` that prunes
unset fields. Two shapes are handled differently:

- **Object-typed `oneOf` members -> `merge()`-prune** (omit the non-chosen member entirely).
  `custom_tags`' `{literal|environment|header}` are objects with required subfields; emitting
  a non-chosen member as null would be sent as an empty `{}`, which both violates its required
  subfields and matches a second `oneOf` arm. Only the chosen member is emitted (its required
  subfields seeded as the merge base).
- **Scalar `oneOf` members and uniform-type leaves -> object constructor (value-or-null).**
  The metrics override `match` (`metric`/`custom_metric`/`mode` -- all strings, with a
  metric-vs-custom_metric `oneOf`) and `tag_overrides` values (`{operation, value}` -- both
  strings) are built this way. `kubernetes_manifest` prunes scalar nulls before the API
  server evaluates the `oneOf`, so only the field that was set is sent; an all-conditional
  `merge()` of these uniform-string fields would instead collapse to a `map(string)` the
  provider cannot morph into an object.

## 6. Composability

- **`namespace`** is the one true foreign key: `StringValueOrRef` -> `KubernetesNamespace`
  (`spec.name`); creates a real DAG edge.
- **`selector.match_labels`** and **`target_refs`** are plain runtime references, NOT foreign
  keys -- istiod resolves them at runtime, so they create no automatic DAG edge. An
  infra-chart author who needs ordering declares it on `metadata.relationships`
  (`depends_on` -> the observed workload/gateway). Provider names refer to MeshConfig
  extension providers, not Planton resources. See the README's "Composing in Infra
  Charts" section for a worked example.
- **Outputs** export the honest `telemetry_name` + `namespace`; Telemetry is a config
  resource with no controller-reconciled status worth surfacing.

## 7. Validation (each rule has accept + reject coverage in `spec_test.go`)

Static (`make protos` incl. the Java gate, `go build`/`vet`, `terraform validate`, bazel +
nogo) -> protovalidate spec tests -> live E2E on `kind` (both engines, full lifecycle). The
spec tests cover: the two union "at most one" CELs (accept zero/one, reject two), the
`TagOverride` UPSERT-needs-value / REMOVE-forbids-value rules (incl. the unset-operation
accept), `random_sampling_percentage` bounds (0/100 accept, -1/100.01 reject),
`reporting_interval` (>= 1ms, malformed reject), every closed enum set
(`WorkloadMode`/`IstioMetric`/`Operation`), required strings
(`ProviderRef.name`/`Literal.value`/`Environment.name`/`RequestHeader.name`), `target_refs`
max 16, the selector/target_refs exclusivity, the selector wildcard rule, and `namespace`
required.

## 8. E2E

- **Prerequisite**: `KubernetesIstioBaseCrds` (kind 868), declared in the registry
  (`prerequisites: [KubernetesIstioBaseCrds]`). The harness installs the istio/base CRDs --
  including `telemetries.telemetry.istio.io` -- before the scenario applies; no full istiod
  is needed to prove the CR is accepted server-side and cleaned up on destroy.
- **Tier 4**, validated on both Pulumi and Terraform via the `ResourceExistenceVerifier`
  (`telemetries.telemetry.istio.io`) dispatched from `aa_e2e/verify/verifier.go`. The
  scenario exercises a nested `custom_tags` (the untyped-Pulumi path), `tag_overrides`
  (UPSERT + REMOVE), the metric `oneOf`, and an access-log filter so both engines'
  transformations are proven against the real CRD schema.

## 9. Future experience surface (kept open, not built here)

The spec is intentionally lossless so future Planton console experiences stay possible
without a schema change: a guided trace-sampling/customisation wizard, a metrics-dimension
editor with provider awareness, and an access-logging policy view all derive from these
exact fields. The console integration is a separate follow-up project.

## References

- [Istio Telemetry reference](https://istio.io/latest/docs/reference/config/telemetry/)
- Upstream proto: `istio.io/api` `telemetry/v1alpha1/telemetry.proto` @ `1.26.8`
