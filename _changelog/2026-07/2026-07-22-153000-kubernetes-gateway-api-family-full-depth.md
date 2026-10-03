# Kubernetes Gateway API family at full depth: v1.6.1 rebuild, two new kinds, composed-ID state import for every typed CR

**Date**: 2026-07-22
**Scope**: `apis/dev/planton/provider/kubernetes` (gateway_api.proto shared types; kubernetesgatewayapicrds, kubernetesgatewayclass, kubernetesgateway, kuberneteshttproute, kubernetesgrpcroute, kubernetestcproute, kubernetestlsroute, kubernetesreferencegrant rebuilt; kubernetesudproute + kuberneteslistenerset forged), `catalogkind` (family renumbered 840–849; Istio 850–858; MetricsServer 859), `pkg/kubernetes/kubernetestypes` (gateway-api pin v1.5.1 → v1.6.1), `pkg/iac/importmap` + `apis/dev/planton/iac` (composed-ID import vocabulary), `aa_import/catalog.yaml` (kubectl_manifest row), fifteen catalog kind import maps, `aa_e2e/verify`, `e2e` + Makefile tiers, `pkg/outputs`, site catalog, `_rules/catalog-kind/update`

## What changed

The complete Gateway API surface — the CRD installer, GatewayClass, Gateway,
ListenerSet, all five route kinds, and ReferenceGrant — rebuilt to full
configuration depth against Gateway API v1.6.1 with dual-engine parity, live
kind-cluster E2E on both engines, and blind state-import round-trips proven
for every kind in the family. The same session landed the import-framework
uplift that typed custom resources needed, retroactively proving the six
cert-manager and external-secrets CR kinds' imports as well.

### Pin and channel reconciliation (v1.5.1 → v1.6.1)

- TCPRoute and UDPRoute graduated to GA `v1` in the Gateway API v1.6 release
  and now ship in the STANDARD channel; ListenerSet has been standard since
  v1.5. The crd2pulumi generator now pulls every type from the standard
  channel at v1.6.1 — the experimental channel is no longer needed for any
  generated type, and the E2E prerequisite fixture installs standard.
- KubernetesTcpRoute moved off `v1alpha2`: the projection, typed Pulumi
  resource, and rendered manifests all serve `gateway.networking.k8s.io/v1`.
  The experimental-only `use_default_gateways` field was removed — it has no
  deployable target in the standard channel.
- Version bumps folded in: TLSRoute hostnames now allow 1024 entries,
  Gateway infrastructure annotations 16, frontend-TLS CA references 16.

### Two new kinds

- **KubernetesUdpRoute (847)** — the GA UDPRoute: weighted datagram routing
  (DNS, syslog, game servers) behind a Gateway's UDP listener. Structural
  twin of KubernetesTcpRoute on the shared backend-ref message.
- **KubernetesListenerSet (843)** — the per-team delegation kind: merges
  additional listeners (ports, hostnames, TLS certificates) into a Gateway
  that opts in via `allowed_listeners`, so teams manage their own listeners
  without touching the shared Gateway. Listener entries reuse the exact
  shared listener building blocks the Gateway uses, so the two kinds cannot
  drift.

### Cross-resource references became real foreign keys

Route `parent_refs[].name` (→ KubernetesGateway), backend `name`s
(→ KubernetesService), listener TLS `certificate_refs[].name`
(→ KubernetesSecret — the cert-manager seam), frontend CA references
(→ KubernetesConfigMap), and the ListenerSet's `parent_ref.name`
(→ KubernetesGateway) are now `StringValueOrRef` foreign keys: infra charts
wire them with `valueFrom` and get real dependency edges instead of manual
relationship hints. ReferenceGrant's from/to fields deliberately stay plain
— they are trust assertions about kinds, not pointers to instances.

### Faithful-projection presence fix (caught live)

The Gateway API CRDs REQUIRE the `group` key on ParametersReference,
ObjectReference, and LocalObjectReference while allowing the empty value
(core API group). Plain proto3 strings dropped `group: ""` from the rendered
manifest (protojson omits empty scalars) and the API server rejected the CR —
caught by the live full-surface Terraform lanes. Those fields are now
`optional` with a presence-required rule, so the key always renders. The
same pattern ReferenceGrant already used, now uniform across the family.

### Terraform modules migrated to kubectl_manifest

All seven existing CR kinds moved from `kubernetes_manifest` (which needs a
live cluster at plan time) to alekc/kubectl's `kubectl_manifest`: routes and
gateways can now be planned before the CRDs exist — single-run infra charts
and offline plan proofs work. Identity labels converged on the `planton.ai/*`
convention in both engines. Pre-anatomy debt paid: per-kind Pulumi project
names, iac-input entrypoints, full-surface hack manifests.

### Composed-ID state import (framework uplift)

- The catalog kind import map vocabulary gained a `literal` derivation arm
  (constants of the module — a typed-CR module's apiVersion/kind).
- id_format templates gained an optional SEGMENT GROUP syntax:
  `[//{namespace}]` disappears wholesale when the placeholder does not
  resolve. This is deliberately a new syntax rather than a change to the
  existing `{name?}` semantics, whose keep-the-delimiters behavior proven
  formats depend on. The kubectl importer rejects a trailing `//`, so
  cluster-scoped CRs render the 3-part ID.
- The kubernetes provider catalog gained the `kubectl_manifest` row
  (`{api_version}//{kind}//{name}[//{namespace}]`) with its provider-side
  knobs declared config-only and `yaml_body` write-normalized (the importer
  stores the stripped live object).
- Fifteen kind maps authored and proven: the nine Gateway API CR kinds
  plus the six previously deferred kinds (ClusterIssuer, Issuer,
  Certificate, ClusterSecretStore, SecretStore, ExternalSecret). The CRD
  bundle installer is recorded as deliberately not applicable (positional
  multi-document addresses carry no per-document GVK).

## Validation

- Ten spec-test suites green (every CEL contract rejection-locked);
  importmap + outputs conformance green; secret-coverage and validate-refs
  green; `make build-go` green; site catalog regenerated.
- Offline `tofu plan` proofs for all ten kinds, full-surface AND
  optionals-absent, with rendered-manifest type-fidelity spot-checks.
- Live on the persistent kind cluster, BOTH engines, full six-phase runner:
  22 scenarios across the ten kinds — including the composed chain
  (GatewayClass fixture → Gateway → HTTPRoute with a live Service backend,
  all FK-wired) and the ListenerSet delegation proof against an opted-in
  Gateway fixture. Zero orphans.
- Blind import round-trips green for all sixteen mapped kinds (ten Gateway
  family + six previously deferred), ledgered in `pkg/iac/importmap/README.md`.
