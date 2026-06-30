# KubernetesDestinationRule -- Design & Research

This document captures the modeling rationale and fidelity decisions behind the
`KubernetesDestinationRule` Planton component (kind 861). It is the largest Istio component:
it models several Istio `oneof` unions, reuses one traffic-policy shape at four nested paths,
and carries the family's only TLS `credential_name`.

## 1. What DestinationRule is

`DestinationRule` is an Istio networking CRD (`networking.istio.io`, served at `v1`,
`v1beta1`, and `v1alpha3` -- identical spec) that configures **how** the mesh talks to a
destination *after* routing has chosen it: the load balancing algorithm, connection-pool
sizing, circuit breaking (outlier detection), and the client TLS the sidecar originates
upstream. It can also define named `subsets` of a service that route rules target.

It does not add destinations (that is `ServiceEntry`) and it does not route (that is
`VirtualService`); it tunes the client-side behavior toward a host already in the registry.

## 2. Source of truth and version

Translated proto-to-proto from the upstream `istio.io/api` clone pinned to tag **1.26.8**
(`networking/v1alpha3/destination_rule.proto`; the `v1` CRD is a type alias of `v1alpha3`).
The local clone is authoritative. The crd2pulumi typed SDK and the CRDs
installed by `KubernetesIstioBaseCrds` are generated from Istio `release-1.26`, so the
proto, the typed Pulumi resource, and the cluster CRD agree on the schema.

DestinationRule is the family's heaviest user of `google.protobuf.Duration`, wrapper types
(`UInt32Value`, `BoolValue`, `DoubleValue`), and proto `oneof`s -- the modeling of all
three is described below.

## 3. The central modeling decision: Istio `oneof` -> optional siblings + "at most one" CEL

DestinationRule has three `oneof`s -- `LoadBalancerSettings.lb_policy {simple,
consistent_hash}`, `ConsistentHashLB.hash_key {http_header_name, http_cookie,
use_source_ip, http_query_parameter_name}`, and `ConsistentHashLB.hash_algorithm
{ring_hash, maglev}` -- plus two documented "only one of" constraints
(`warmup`/`warmup_duration_secs`; locality `distribute`/`failover`/`failover_priority`).

The rule is "never carry a proto `oneof`; make the union explicit and validated." There are
**two faithful ways** to realize that, and the right one depends on the union's shape:

- **(a) Same-type, collapsible members -> discriminator + single `value`.** Used by the
  shared `KubernetesIstioApiStringMatch` (`exact`/`prefix`/`regex` are all strings, so they
  collapse to a `match_type` discriminator + one `value`). This is the gateway-family
  `HTTPRouteFilter` shape -- but note the gateway `type` discriminator is faithful only
  because the Gateway API upstream *natively has* a `type` field.
- **(b) Distinctly-named members -> optional sibling fields + an "at most one" message CEL.**
  Used by ServiceEntry (`workload_selector` vs `endpoints`) and EnvoyFilter
  (`workload_selector` vs `target_refs`).

DestinationRule's unions are entirely case (b): each member maps 1:1 to a distinctly-named
CRD JSON key, and Istio's proto has **no** native discriminator. Adding an `lb_policy_type`
discriminator would invent a required input the CRD does not have (violating 100% upstream
fidelity) and would reject a legitimately-valid `loadBalancer` that sets only `localityLbSetting`/`warmup`
and neither arm. So each union is modeled as optional sibling fields constrained by a
message-level CEL:

| Union | Planton CEL (id) |
|-------|------------------|
| `lb_policy {simple, consistent_hash}` | `destination_rule_load_balancer.lb_policy_at_most_one` |
| `warmup` vs `warmup_duration_secs` (upstream `oneof(...)` XValidation) | `destination_rule_load_balancer.warmup_at_most_one` |
| `hash_key {4 fields}` | `destination_rule_consistent_hash.hash_key_at_most_one` |
| `hash_algorithm {ring_hash, maglev}` | `destination_rule_consistent_hash.hash_algorithm_at_most_one` |
| locality `distribute` / `failover` / `failover_priority` | `destination_rule_locality.distribution_at_most_one` |

Each CEL is `(has(a)?1:0) + (has(b)?1:0) + ... <= 1` (repeated fields use `size(x) > 0`
because cel-go rejects `has()` on repeated fields). This preserves the upstream JSON shape
exactly and loses nothing for a future console wizard (a radio group sets the chosen field
and leaves the siblings null; the selected option is inferred from which field is non-null).

## 4. Other fidelity decisions

### Workload selector -- the POLICY selector, not the networking one

The upstream `destination_rule.proto` declares `istio.type.v1beta1.WorkloadSelector
workload_selector` (field `match_labels`, JSON `matchLabels`) -- the **same** selector
PeerAuthentication / AuthorizationPolicy use, NOT the networking `labels` selector that
ServiceEntry / EnvoyFilter use. DestinationRule therefore reuses the existing shared
`KubernetesIstioApiWorkloadSelector`; its CRD constraints (max 4096 entries, value <= 63,
non-empty / no-wildcard keys, no-wildcard values) match that type exactly.

### Durations and wrappers

Every `google.protobuf.Duration` is an `optional string`. Where upstream documents
">= 1ms" (`connect_timeout`, `max_connection_duration`, outlier `interval`,
`base_ejection_time`) the CEL is `this == '' || duration(this) >= duration('1ms')`; where
upstream sets `duration-validation:none` (cookie `ttl`, the idle timeouts, keepalive
`time`/`interval`, `warmup_duration_secs`) the CEL is `>= duration('0s')` (valid,
non-negative). Wrapper types become the matching `optional` scalar: outlier `consecutive_*`
(`UInt32Value` -> `uint32`), `insecure_skip_verify` / locality `enabled` (`BoolValue` ->
`bool`), warmup `minimum_percent` (0-100) and `aggression` (>= 1) (`DoubleValue` ->
`double`).

### Closed-set strings

`simple`, `h2_upgrade_policy`, TLS `mode`, tunnel `protocol`, and proxy `version` are
`optional string` validated by `string.in [...]` against the UPPERCASE upstream constants
(not proto enums). `tunnel.protocol` and `simple` track the documented sets; the CRD leaves
`tunnel.protocol` free-form, so constraining it to `CONNECT`/`POST` is a deliberate,
documented tightening (the only place this component is stricter than the raw CRD).

### Deprecated / hidden fields

The deprecated-and-hidden `OutlierDetection.consecutive_errors` (`$hide_from_docs`) is
omitted. The deprecated-but-visible fields are kept for fidelity (the CRD still accepts
them), each labeled in its comment: `warmup_duration_secs` (prefer `warmup`),
`ConsistentHashLB.minimum_ring_size` (prefer `ring_hash`), and the `LEAST_CONN` value in the
`simple` set (prefer `LEAST_REQUEST`).

## 5. IaC implementation

Both engines emit the same `networking.istio.io/v1` `DestinationRule`.

### The crd2pulumi path-typed duplication (Pulumi)

The upstream CRD reuses one `TrafficPolicy` (and one `PortTrafficPolicy` / `LoadBalancer` /
`ConnectionPool` / `OutlierDetection` / `ClientTLSSettings`) by reference at four reachable
paths: `spec.trafficPolicy`, `spec.subsets[].trafficPolicy`, and the `portLevelSettings[]`
under each. **crd2pulumi does not share a Go type across reference paths** -- it emits a
distinct, path-named struct at every path (e.g.
`DestinationRuleSpecTrafficPolicyLoadBalancerArgs` vs
`DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerArgs`). These generated structs share no
common settable interface, so the builders cannot be collapsed into one generic function.
The proto stays DRY (one shared message per shape); only the Pulumi adapter is duplicated --
the price of compile-time-typed CRD args. It is isolated in `traffic_policy.go`
with a file header explaining why, and all leaf scalar mapping goes through shared
`opt*`/`strArr`/`u32IntMap` helpers so each per-path builder is a one-statement struct
literal. Unions are mapped by which proto field is set (no discriminator). The `Spec` field
is a `PtrInput` satisfied by the `DestinationRuleSpecArgs` value (the `*Ptr()` wrapper
marshals to the wrong element type and panics at apply).

### Terraform

`kubernetes_manifest` with a fully-typed `variable "spec"` and a `locals.tf` that builds the
manifest by `merge()`-pruning conditional fragments, so UNSET fields are omitted entirely.
This is required by the DestinationRule CRD's `oneOf` groups (loadBalancer
simple-vs-consistentHash; consistentHash hashKey/hashAlgorithm arms): emitting a non-chosen
alternative as an explicit `null` makes the `required`-based `oneOf` arms match ("found 2
valid alternatives"), and a null object with required subfields (warmup.duration,
tunnel.target*, httpCookie.name) would be sent as an invalid empty object. Required subfields
are seeded as the merge base; everything else is pruned. HCL has no functions, so to transform
each reused leaf exactly once the locals gather every `LoadBalancer`/`ConnectionPool`/
`OutlierDetection`/`TLS` instance (across all four paths) into a flat map keyed by path,
transform the map, then look the built leaf back up during assembly.

## 6. Composability

- **`namespace`** is the one true foreign key: `StringValueOrRef` -> `KubernetesNamespace`
  (`spec.name`); creates a real DAG edge.
- **`host`**, **`workload_selector.match_labels`**, and TLS **`credential_name`** are plain
  runtime references, NOT foreign keys. `host` is a registry name resolved by istiod;
  `match_labels` is matched against pod labels; `credential_name` is a Secret/cert name
  resolved by Envoy. None create an automatic DAG edge. An infra-chart author who needs
  ordering declares it on `metadata.relationships` (`depends_on` -> KubernetesService;
  `uses` -> KubernetesSecret/KubernetesCertificate). See the README's "Composing in Infra
  Charts" section.
- **Outputs** export the honest `destination_rule_name` + `namespace`; a DestinationRule is
  a policy resource with no controller-reconciled status worth surfacing.

## 7. Validation (each rule has accept + reject coverage in `spec_test.go`)

Static (`make protos` incl. the Java gate, `go build`/`vet`, `terraform validate`, bazel +
nogo) -> protovalidate spec tests -> live E2E on `kind` (both engines, full lifecycle). The
spec tests cover: the three union "at most one" CELs (accept zero/one, reject two), warmup
and locality exclusivity, duration minimums (`1ms` accept, `0s`/malformed reject, `1.5s`
accept), warmup bounds (`minimum_percent` 0-100, `aggression` >= 1), every closed enum set,
`host` required, tunnel `target_host`/`target_port`, `http_cookie.name` required, subset
`name` required, and the selector wildcard rule.

## 8. E2E

- **Prerequisite**: `KubernetesIstioBaseCrds` (kind 868), declared in the registry
  (`prerequisites: [KubernetesIstioBaseCrds]`). The harness installs the istio/base CRDs
  before the scenario applies -- no full istiod is needed to prove the CR is accepted
  server-side and cleaned up on destroy.
- **Tier 4**, validated on both Pulumi and Terraform via the `ResourceExistenceVerifier`
  (`destinationrules.networking.istio.io`) dispatched from `aa_e2e/verify/verifier.go`.

## 9. Future experience surface (kept open, not built here)

The spec is intentionally lossless so future Planton console experiences stay possible
without a schema change: a guided "configure load balancing / circuit breaking" wizard, a
subset/canary editor wired to the matching `VirtualService`, a TLS-origination helper that
links `credential_name` to a `KubernetesSecret`, and a per-host policy view that overlays
outlier-detection thresholds all derive from these exact fields. The console integration is
a separate follow-up project.

## References

- [Istio DestinationRule reference](https://istio.io/latest/docs/reference/config/networking/destination-rule/)
- Upstream proto: `istio.io/api` `networking/v1alpha3/destination_rule.proto` @ `1.26.8`
