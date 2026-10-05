# Kubernetes Istio family at full depth: 1.30.3 rebuild, ambient-first mesh kind, live routing and enforcement proofs

**Date**: 2026-07-22
**Scope**: `apis/dev/planton/provider/kubernetes` (istio_api.proto shared types; kubernetesistio + kubernetesistiobasecrds rebuilt; kubernetesdestinationrule, kubernetesserviceentry, kubernetespeerauthentication, kubernetesrequestauthentication, kubernetesauthorizationpolicy, kubernetestelemetry, kubernetesenvoyfilter reconciled), `pkg/kubernetes/kubernetestypes` (istio pin release-1.26 → 1.30.3), `pkg/iac/importmap` + `apis/dev/planton/iac/catalogkindimportmap` (tofu_resource_name scoping), `pkg/iac/tofu/generators` (kubectl_manifest module generation), eight catalog kind import maps, `aa_e2e/verify` (three new verifiers), `e2e` + `e2e/framework/runner` (per-manifest dependency stacks) + Makefile tiers, `pkg/outputs`, site catalog, `_rules/catalog-kind/update`

## What changed

The complete Istio family — the mesh control-plane kind, the CRDs-only
installer, and the seven typed mesh-configuration kinds — rebuilt or
reconciled to full configuration depth against Istio 1.30.3 with dual-engine
parity, live kind-cluster E2E on both engines, and blind state-import
round-trips proven for every mappable kind. The session also landed the
catalog's first BEHAVIORAL Gateway API proofs: a live request routed through
an istiod-provisioned gateway, and live authorization-policy enforcement.

### Pin reconciliation (release-1.26 → 1.30.3)

- The crd2pulumi generator, the CRD-installer modules, and every spec comment
  now pin the exact release tag 1.30.3. The CRD installer previously fetched
  from the MOVING `release-1.26` branch ref — the same deployed resource
  would install different CRD schemas as upstream patched; tag pinning keeps
  installs reproducible and matched to the generated SDK.
- Surface additions verified field-by-field against the 1.30.3 CRDs (and,
  where the schema is silent, the istiod validating-webhook source):
  DestinationRule `retry_budget` and consistent-hash cookie `attributes`;
  ServiceEntry `DYNAMIC_DNS` resolution with four webhook-mirroring
  validation rules; RequestAuthentication optional `issuer` (with an
  issuer-or-jwks_uri rule matching the webhook's own rejection) and
  `space_delimited_claims`; AuthorizationPolicy `trust_domains` /
  `not_trust_domains`; Telemetry `disable_context_propagation` and the
  `formatter` custom-tag source; EnvoyFilter `WAYPOINT` patch context and
  the `waypoint` match object for ambient waypoint proxies.

### KubernetesIstio: the control-plane kind, rebuilt

- Typed surface over the official base/istiod/cni/ztunnel charts (one pinned
  version drives all releases): control-plane revision, first-class
  `dataplane_mode` (sidecar or ambient — ambient installs the istio-cni node
  agent and the ztunnel per-node proxy), istiod sizing with the
  chart-managed HPA (replicas and autoscaling made mutually exclusive at
  validation), MeshConfig essentials (trust domain, REGISTRY_ONLY egress
  lockdown, access logging, multi-cluster identity), proxy defaults,
  sidecar-injection policy, gateway-class service-type defaults, image
  source overrides, and per-release `helm_values` escape hatches.
- Deliberately NO bundled gateway release: istiod implements the Kubernetes
  Gateway API, so north-south exposure composes from KubernetesGateway
  resources (`gateway_class_name: istio`) and istiod provisions gateway
  deployments itself. Outputs export the composition handles
  (`gateway_class_name`, `istiod_service_name`, `revision`, `trust_domain`,
  `dataplane_mode`).
- CRDs are module-owned (server-side apply, outside the Helm release, with
  the chart's copies excluded via `base.excludedCRDs`): Helm refuses to
  adopt CRDs that already exist without its ownership metadata, so a cluster
  running the CRDs-only installer could never have upgraded to the full mesh
  if the chart owned them. Both kinds now co-own the CRDs and that migration
  is a plain redeploy — caught live by the E2E lanes.

### Live behavioral proofs (both engines)

- **Gateway API routing**: a scenario chains the mesh as a fixture, creates
  a Gateway on the `istio` class plus an HTTP backend, and the verifier
  asserts the route Accepted, the Gateway Programmed, and a live request
  through the auto-provisioned gateway returning the backend's response.
- **AuthorizationPolicy enforcement**: a meshed client's request to a meshed
  backend is denied (403) while a DENY policy exists and succeeds again
  after the policy is destroyed — the full enforce-and-release cycle.
- New verifiers: Istio control-plane install (istiod Available + CRDs
  Established + ambient DaemonSets fully ready), gateway routing, and authz
  enforcement — dispatched when a scenario names KubernetesIstio as a
  prerequisite.

### Import framework: multi-release modules and name-keyed bundles

- `CatalogKindImportMap` value declarations can be scoped to one Terraform
  logical resource via `tofu_resource_name` — needed the first time a module
  declared several resources of one type whose ID placeholders differ (the
  mesh's four Helm releases). Scoped declarations win, unscoped are the
  fallback, and the offline conformance guard rejects scopes that name no
  real module resource.
- Multi-document CRD installers key each `kubectl_manifest` by the CRD's own
  name (never the split index), so `from_address_key` derives the composed
  import ID blind and state addresses survive bundle reorderings.
- Eight import maps authored and blind-round-trip proven live, including the
  mesh kind's 20-resource import (four scoped releases, fifteen name-keyed
  CRDs, the anchor namespace).

### Module generation and framework fixes

- `planton tofu generate-module` now emits the settled kubectl_manifest
  anatomy for projection kinds: the alekc/kubectl provider (plannable before
  CRDs exist), `planton.ai/*` identity labels, `backend.tf`, and a typed
  metadata variable with defaulted optional attributes (an `any`-typed
  metadata breaks identity-label reads when the tfvars carry only `name` —
  HCL's `&&` evaluates both operands even on missing attributes).
- E2E dependency stacks are named per MANIFEST, not per kind: a scenario
  chaining two extra-instance fixtures of one kind previously landed both on
  a single stack, the second silently replacing the first.
- The update rule gained the CRD-ownership lesson, the tofu_resource_name /
  name-keyed-bundle mechanics, and both are recorded in the importmap
  README.

### Registration and satellites

Outputs-conformance cases for all nine kinds (the family had none); the
whole family moved to Tier 1 in the E2E test lists and Makefile regexes
(including the mesh kind's first Terraform lanes — its old Pulumi-only
deferral dissolved with the gateway-release removal); presets refreshed and
machine-validated (ambient and production mesh presets, a dynamic-DNS
egress preset); docs, module READMEs, and catalog pages rebuilt; site
catalog regenerated.
