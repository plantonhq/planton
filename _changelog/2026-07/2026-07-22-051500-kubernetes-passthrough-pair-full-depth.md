# Kubernetes passthrough pair rebuilt: raw manifests and Helm releases at full depth on both engines

**Date**: 2026-07-22
**Scope**: `apis/dev/planton/provider/kubernetes` (kubernetesmanifest, kuberneteshelmrelease — both rebuilt), `aa_import`, `aa_e2e/verify`, `e2e` (incl. `framework/runner/import_roundtrip.go`), `pkg/outputs`, `pkg/iac/importmap`, `pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider`, Makefile E2E tiers, site catalog, `_rules/catalog-kind/update`, `go.mod` (helm.sh/helm/v3 for strvals)

## What changed

The catalog's two escape-hatch kinds — deploy any raw Kubernetes YAML, and
install any upstream Helm chart — rebuilt to the same bar as every typed
kind, with dual-engine parity and live E2E on both engines.

### KubernetesManifest (rebuilt)

- **Terraform module forged (net-new)** — the kind was previously
  Pulumi-only. One `kubectl_manifest` per document (identity-keyed
  `for_each`, so reordering documents never churns state addresses),
  server-side apply, CRD-before-CR in one pass.
- **Namespace anchoring made real and identical on both engines**: documents
  with an explicit `metadata.namespace` keep it; namespaced documents
  without one land in `spec.namespace` (Pulumi: provider-level default
  namespace via a new shared-seam helper; Terraform: per-document
  `override_namespace` only when the document declares none); cluster-scoped
  documents pass through untouched. Previously the spec comment promised
  this and no engine implemented it.
- **`skip_await` knob** (default: await readiness on both engines).
- **Outputs upgraded**: anchor namespace + an applied-resource inventory
  ("apiVersion/Kind/name" per document) derived by parsing the input YAML
  identically on both engines.
- Import map deliberately not applicable (arbitrary user YAML has no
  per-kind schema for a blind round-trip oracle); recorded in the importmap
  README so absence is never mistaken for an oversight.

### KubernetesHelmRelease (rebuilt) — the sole intentional passthrough

- **Pulumi module now creates a REAL Helm release** (`helm.v3.Release`):
  hooks run, the release secret is written, `helm list` sees it. The
  previous module used `helm.v3.Chart` — a client-side template render that
  never created a release, silently skipped hooks, and was invisible to Helm
  tooling: a true cross-engine semantic divergence, now gone.
- **Terraform values wiring fixed**: the module's `values` was commented
  out — user-supplied values were silently ignored on that engine.
- **Spec rebuilt to the full passthrough surface**: chart identity
  (`repo` incl. `oci://` registries, `chart`, pinned `version`,
  `release_name` override), the Helm-native values model — `values_yaml`
  plus `set` / `set_string` / `set_sensitive` overrides with one documented
  precedence, merged through Helm's own strvals parser on BOTH engines
  (byte-identical releases) — private-repo credentials (secret-by-default),
  and the intersection lifecycle surface: atomic, cleanup_on_fail,
  skip_await, wait_for_jobs, timeout, skip_crds, dependency_update,
  max_history, replace, force_update, reuse/reset_values, disable_webhooks,
  disable_openapi_validation, take_ownership, description. Cross-field CEL
  rejects atomic+skip_await, reuse+reset, and half-set repo credentials.
- **One parity exception, loud**: `take_ownership` is Pulumi-inexpressible
  at the pinned SDK (the argument landed in a later pulumi-kubernetes
  release); the Pulumi module rejects a set field with an error naming the
  working engine — negative-proofed offline. Terraform implements it
  (provider floor raised to `>= 3.1`, the release that added the flag).
- **Outputs upgraded** from namespace-only to the release's observable
  handles: release_name, chart version, app_version, status, revision —
  read from what Helm actually recorded, both engines.
- **Import recipes shipped and proven**: `helm_release` catalog entry
  (`{namespace}/{name}` id; install-time attributes declared config-only —
  Helm does not persist how a release was installed), catalog kind import map,
  blind round-trip green on all three scenarios.

### Framework and satellites

- **Import round-trip oracle: plan-declared after-unknown pruning.**
  Plugin-framework providers (helm_release is the first in the catalog)
  mark computed attributes (`id`, `metadata`) wholly unknown on every
  in-place update; the oracle now prunes exactly the attributes the plan
  itself declares after-unknown — partially-unknown objects and all known
  drift still fail. Regression-proven against a previously-green kind.
- New shared seam: `GetWithKubernetesProviderConfigAndNamespace` (Pulumi
  provider with a scope-aware default namespace) for modules applying
  user-authored manifests.
- Both kinds: settled module anatomy (per-kind Pulumi project names,
  iac-input.yaml entrypoints, hack manifests at `iac/hack/` exercising
  the full surface — the HelmRelease hack manifest previously failed its
  own spec validation), `planton.ai/*` identity labels, three presets each
  (machine-validated), rewritten docs/README/catalog pages leading with
  when-NOT-to-use-this boundary language, outputs conformance cases,
  E2E registration (HelmRelease joins Tier 1; profile un-deferred), site
  catalog regenerated.
- E2E verifier: `kuberneteshelmrelease` dispatches to the Helm kind
  verifier (namespace + running pods + service).
- Update rule amended (timeless): the loud-failure discipline for
  engine-inexpressible fields now covers the Pulumi-side direction, with
  the pinned-SDK verification mechanics.
- Makefile: stale retired-kind residue removed from the Tier-2 E2E regex.

## Validation

- Spec tests green for both kinds (valid/invalid matrices incl. every new
  CEL rule); per-kind + release-entrypoint builds; `make build-go`;
  secret-coverage; validate-refs; import-map conformance (one pre-existing
  aws/awsecrrepo failure is the AWS surface's, unrelated); outputs
  conformance with two new cases; kind map + e2e matrix regenerated.
- Offline proofs: tofu validate + plan proofs from the hack manifests (per-
  document namespace override visible in the plan), Pulumi preview proofs,
  and the take_ownership negative proof (loud rejection observed).
- Live E2E on the persistent kind cluster, both engines: Manifest 3
  scenarios × 2 (incl. namespace-defaulting with a cluster-scoped document;
  first-ever Terraform lanes for the kind), HelmRelease 3 scenarios × 2
  (HTTPS repo, OCI registry, values-merge proof — real releases verified).
  Blind import round-trip green for all three HelmRelease scenarios.
  Zero orphan namespaces or cluster-scoped residue after the lanes.
