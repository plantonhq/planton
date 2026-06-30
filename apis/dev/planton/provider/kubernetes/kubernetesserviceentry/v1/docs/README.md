# KubernetesServiceEntry -- Design & Research

This document captures the deployment landscape, the modeling rationale, and the
fidelity decisions behind the `KubernetesServiceEntry` Planton component (kind 862).
It uses the `networking/v1alpha3` workload selector, the location/resolution enums,
the list-map port shape, and inline `WorkloadEntry` endpoints.

## 1. What ServiceEntry is

`ServiceEntry` is an Istio networking CRD (`networking.istio.io`, served at `v1`,
`v1beta1`, and `v1alpha3` -- identical spec) that **adds an entry to Istio's internal
service registry**. It makes a destination the mesh would not otherwise know about
(an external API, a SaaS endpoint, a VM, a legacy service) routable and
policy-addressable: once registered, VirtualService / DestinationRule / Telemetry /
AuthorizationPolicy can all target its `hosts`.

It is distinct from `DestinationRule` (which configures *how* to talk to a destination
already in the registry) and `Sidecar` (which limits *which* registry entries a
workload sees). ServiceEntry is the only one of the three that *adds* destinations.

## 2. Source of truth and version

Translated proto-to-proto from the upstream `istio.io/api` clone pinned to tag
**1.26.8** (`networking/v1alpha3/service_entry.proto`; the `v1` CRD is a type alias of
`v1alpha3`). The local clone is authoritative; no specs are pulled from the internet.
The crd2pulumi typed SDK and the CRDs installed by `KubernetesIstioBaseCrds` are
likewise generated from Istio `release-1.26`, so the proto, the typed Pulumi resource,
and the cluster CRD all agree on the schema.

ServiceEntry has **no** `google.protobuf.Duration`, no wrapper types, and **no proto
`oneof`**. It also has **no `credential_name`/TLS field** in 1.26.8 (that lives on
DestinationRule).

## 3. Planton spec shape (fidelity decisions)

The Planton spec flattens the upstream `ServiceEntry` fields directly after the
namespaced envelope (`target_cluster`, `namespace`). There is no nested
`service_entry` sub-message. Fields are numbered sequentially from 1 -- the Planton
proto is its own wire contract and does not preserve upstream field-number gaps.

| Upstream | Planton | Notes |
|----------|---------|-------|
| `hosts` / `addresses` / `exportTo` / `subjectAltNames` | same (`repeated string`) | Plain lists; not foreign keys. |
| `ports` (`ServicePort`) | `ports` (`KubernetesServiceEntryPort`) | List-map keyed by `name`; `number`+`name` unique. |
| `location` (enum) | `location` (`optional string`, closed set) | UPPERCASE string + `in`, not a proto enum. |
| `resolution` (enum) | `resolution` (`optional string`, closed set) | UPPERCASE string + `in`, not a proto enum. |
| `endpoints` (`WorkloadEntry`) | `endpoints` (`KubernetesServiceEntryEndpoint`) | Inline subset; `ports` is `map<string,uint32>`. |
| `workloadSelector` (`networking.v1alpha3.WorkloadSelector`) | `workload_selector` (`KubernetesIstioApiNetworkingWorkloadSelector`) | New shared type (see below). |

### The two workload selectors (a deliberate distinction)

Istio has **two** different `WorkloadSelector` messages, and ServiceEntry uses the one
the policy CRDs do not:

- `istio.type.v1beta1.WorkloadSelector` -- field `match_labels` (JSON `matchLabels`).
  Used by PeerAuthentication / RequestAuthentication / AuthorizationPolicy. Modeled by
  the shared `KubernetesIstioApiWorkloadSelector`.
- `istio.networking.v1alpha3.WorkloadSelector` -- field `labels` (JSON `labels`). Used
  by ServiceEntry, Sidecar, Gateway, and EnvoyFilter. Modeled by the shared
  `KubernetesIstioApiNetworkingWorkloadSelector`. (Note: DestinationRule uses the
  type/v1beta1 `match_labels` selector instead, despite being a networking CRD.)

They are not interchangeable -- the JSON key differs (`labels` vs `matchLabels`), so
conflating them would emit a CR the CRD rejects. The networking selector is
constrained to ServiceEntry's CRD exactly (max 256 entries, value <= 63 chars,
no-wildcard *values*); unlike the policy selector it deliberately does **not** enforce
non-empty / no-wildcard *keys*, because the ServiceEntry CRD does not -- adding those
rules would reject configurations the CRD accepts (match the validated outcome).
Each consumer must re-confirm its own CRD's selector constraints before reusing this type.

### Enum case / external standard

`location` and `resolution` are closed-set `optional string` fields validated by the
protovalidate `string.in` rule with the UPPERCASE upstream constants
(`MESH_EXTERNAL`/`MESH_INTERNAL`; `NONE`/`STATIC`/`DNS`/`DNS_ROUND_ROBIN`) -- not proto
enums (which would change CEL string-comparison semantics). `protocol` is modeled the
same way against the documented set `HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS`. Leaving any
of them unset omits the field from the CR, so istiod's own default
(`MESH_EXTERNAL`/`NONE`) applies -- the upstream defaults are not baked in.

### CEL portability

Upstream CRD `XValidation` rules use the Kubernetes apiserver CEL extension library
(`oneof()`, `default()`), which `buf`/`protovalidate`/`cel-go` do not provide. Each
rule is re-expressed with portable CEL primitives -- `has()` guards over the proto3
`optional` fields (the exact idiom `gateway_api.proto` uses) -- that preserve the
validated outcome. Unset `resolution` is treated as its upstream default `NONE`.

| Upstream intent | Planton CEL |
|-----------------|-------------|
| `oneof(workloadSelector, endpoints)` | `!(has(this.workload_selector) && size(this.endpoints) > 0)` |
| CIDR addresses require NONE/STATIC | `(!has(this.resolution) \|\| this.resolution in ['NONE','STATIC']) \|\| !this.addresses.exists(a, a.contains('/'))` |
| NONE forbids endpoints | `!((!has(this.resolution) \|\| this.resolution == 'NONE') && size(this.endpoints) > 0)` |
| DNS_ROUND_ROBIN <= 1 endpoint | `!(has(this.resolution) && this.resolution == 'DNS_ROUND_ROBIN' && size(this.endpoints) > 1)` |
| ports: unique `number` / unique `name` | `this.ports.all(p1, this.ports.exists_one(p2, p1.number == p2.number))` (and `.name`) |
| endpoint: address-or-network; UDS no ports | message-level CEL on the endpoint; UDS path/dir shape as field CEL on `address` |

### Validation rules (each has accept + reject coverage in `spec_test.go`)

| Rule | Origin |
|------|--------|
| `service_entry.workload_selector_xor_endpoints` (message-level) | Upstream selector/endpoints XValidation. |
| `service_entry.cidr_addresses_require_none_or_static` (message-level) | Upstream CIDR/resolution XValidation. |
| `service_entry.none_resolution_forbids_endpoints` (message-level) | Upstream NONE/endpoints XValidation. |
| `service_entry.dns_round_robin_single_endpoint` (message-level) | Upstream DNS_ROUND_ROBIN XValidation. |
| `service_entry.port_{numbers,names}_unique` (message-level) | Upstream list-map + duplicate-number XValidation. |
| `hosts` required / max 256 / no bare `*`; `addresses` max 256, len <= 64 | Upstream field markers. |
| port `number`/`target_port` 1-65535; `name` required; `protocol` closed set | Upstream field markers + documented protocol MUST. |
| `location`/`resolution` closed sets | Upstream enum. |
| endpoint `address`-or-`network`; UDS no-ports / absolute-or-abstract / not-a-dir; `ports` 1-65535 + key pattern | Upstream WorkloadEntry XValidations + markers. |
| `workload_selector.labels` value no-wildcard | Networking WorkloadSelector CRD. |

## 4. Composability

- **`namespace`** is the one true foreign key: `StringValueOrRef` ->
  `KubernetesNamespace` (`spec.name`). Literal or `valueFrom`; creates a real DAG edge.
- **`hosts` / `addresses` / `workload_selector.labels`** are plain runtime values, NOT
  foreign keys. A ServiceEntry's `hosts` are the registry names it *defines*, not
  references it consumes; `workload_selector` is matched against pod labels by istiod.
  None create an automatic DAG edge. An infra-chart author who needs ordering (e.g. a
  MESH_INTERNAL ServiceEntry that fronts a Deployment's pods) declares it on
  `metadata.relationships` (`depends_on` -> KubernetesDeployment). See the component
  README's "Composing in Infra Charts" section.
- **Outputs** export the honest `service_entry_name` + `namespace`; a ServiceEntry is a
  registry entry with no controller-reconciled status worth surfacing.

## 5. IaC implementation

Both engines are feature-equal and emit the same `networking.istio.io/v1`
`ServiceEntry`:

- **Pulumi** uses the typed crd2pulumi resource `istionetworkingv1.NewServiceEntry`.
  The `Spec` field is a `PtrInput` satisfied by the `ServiceEntrySpecArgs`
  **value** (the `*Ptr()` wrapper marshals to the wrong element type and panics at
  apply). `hosts` is always set
  (required); `ports`, `endpoints`, `workload_selector`, the string lists, and the
  optional `location`/`resolution` scalars are attached only when present. Proto
  `uint32` port numbers/weights are cast to the SDK's `int` inputs, and the endpoint
  `map<string,uint32>` ports become a `pulumi.IntMap`.
- **Terraform** uses `kubernetes_manifest` with a fully-typed `variable "spec"` and a
  null-pruned `locals.tf`. Each list element is assembled with `merge()` so only the
  fields the user set reach the manifest; the nested `workload_selector` is read
  through a `?:` guard (HCL `&&` does not short-circuit). Snake_case spec fields map to
  the CRD's camelCase (`targetPort`, `serviceAccount`, `subjectAltNames`, `exportTo`,
  `workloadSelector`).

## 6. E2E

- **Prerequisite**: `KubernetesIstioBaseCrds` (kind 868), declared in the registry
  (`prerequisites: [KubernetesIstioBaseCrds]`). The harness installs the istio/base
  CRDs before the scenario applies -- no full istiod is needed to prove the CR is
  accepted server-side and cleaned up on destroy.
- **Tier 4**, validated on both Pulumi and Terraform. Verification asserts the
  ServiceEntry CR exists after apply and is gone after destroy
  (`serviceentries.networking.istio.io`), via the `ResourceExistenceVerifier`
  dispatched from `aa_e2e/verify/verifier.go`.

## 7. Future experience surface (kept open, not built here)

The spec is intentionally lossless so future Planton console experiences stay possible
without a schema change: an "external services" catalog per namespace, DNS/JWKS-style
reachability checks against `hosts`, a guided "add an external API" wizard that picks
resolution from whether addresses/endpoints are supplied, and a topology view that
draws MESH_INTERNAL ServiceEntries against the workloads their `workload_selector`
matches all derive from these exact fields. The console integration itself is a
separate follow-up project.

## References

- [Istio ServiceEntry reference](https://istio.io/latest/docs/reference/config/networking/service-entry/)
- [Istio egress / accessing external services task](https://istio.io/latest/docs/tasks/traffic-management/egress/egress-control/)
- Upstream proto: `istio.io/api` `networking/v1alpha3/service_entry.proto` @ `1.26.8`
