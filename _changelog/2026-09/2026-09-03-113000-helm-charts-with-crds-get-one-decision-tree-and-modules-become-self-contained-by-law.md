# Helm charts with CRDs get ONE decision tree, and a module becomes its directory by law — the rules, two guards, an honest e2e harness, and the proof that derivation works

## What changed

- **The forge and update rules carry the CRD decision tree.** Every chart
  that carries custom resource definitions sits on exactly one of three
  branches: the chart owns its CRD lifecycle (map
  `crds { install, keep_on_uninstall }` into the chart's own values and
  nothing else); upstream publishes a separate CRD chart (a second kept
  release ahead of the operator release); or the module DERIVES the CRD
  set from the pinned version — render the pinned chart with its CRD
  switch on, or fetch the pinned upstream bundle — and applies it kept
  ahead of a release that skips CRDs. There is no fourth branch. The
  guidance that told a kind to stage rendered CRD files beside its module
  and read them through `../crds` is gone, replaced with the two
  invariants (a module is its directory and reads nothing outside it; a
  module derives, never copies, what a pinned artifact carries), the
  verified facts the derive branch rests on, and the failure standard:
  every precondition, error, and refusal names what was observed, what it
  means, and the next step — mechanism-only text is a defect.
  (`_rules/catalog-kind/forge/forge-catalog-kind.mdc`,
  `_rules/catalog-kind/forge/README.md`,
  `_rules/catalog-kind/update/update-catalog-kind.mdc`.)
- **Two guards make the retired shape unshippable.** The anatomy gate
  (`pkg/anatomy`) no longer admits `iac/crds/`; it is an unexpected entry
  with the remedy in its message, and the four kinds that still carry it
  sit in `baseline.yaml` under the baseline's own must-shrink contract. A
  new static guard, `hack/guards/ensure_modules_are_self_contained.sh`,
  fails any Terraform module that reads `${path.module}/..` or a `"../`
  literal and any Pulumi module carrying a parent-path string literal; it
  carries a `known_violations` table (the eight modules of the same four
  kinds) that FAILS when a listed module stops violating, so the table
  can only shrink, and `SELF_CONTAINED_GUARD_IGNORE_KNOWN=1` shows the
  raw result. Wired as a third invariant of
  `.github/workflows/lint.release-packaging.yaml`, because "a module is
  what its zip contains" is a packaging truth spanning both engines.
- **The e2e harness runs Terraform modules the way releases ship them.**
  `PrepareWorkDir` copies the module directory and nothing beside it — the
  exact file set `module.zip` carries. The parent-tree copy that let a lane
  pass on a layout no release ever had is deleted, comment included.
  Proven on Kind: `TestKubernetesKeda_Terraform` (three scenarios, 249 s)
  green from the plain copy.
- **Four e2e profiles tell the truth.** `kubernetessolroperator`,
  `kubernetesoteloperator`, `kubernetesopensearchoperator`,
  `kubernetesplantonoperator` move from `status: green` to
  `status: deferred` with the reason: both engines read `../crds`, which
  no published form of the module carries, so no fresh install from a
  release has ever converged and a lane passing from the source tree proves
  nothing about what ships. The deferral lifts when each module derives its
  CRDs and leaves the guard's table. `planton e2e discover`, the
  fixture-integrity walk, and the tier-wiring guard all accept the change.
- **The break-glass docker-config read is documented where it lives.**
  `kubernetesannotationkeys.DockerConfigJsonFileAnnotationKey` and the
  `loadDockerConfigFromFile` helper in the five workload kinds (Deployment,
  StatefulSet, DaemonSet, Job, CronJob) now say what they are: an expert
  applying a module from their own laptop names a file on their own disk.
  The read is of the operator's machine, never the module's directory, and
  sits deliberately outside the self-containment invariant.

## Why

The catalog answered "how does a kind install a chart that carries CRDs"
five different ways, and the four that vendored a CRD copy beside the module
had never once worked from a published release: release packaging zips
exactly `iac/tf` or `iac/pulumi`, the Pulumi binary lane runs from a
generated workspace with no files at all, and `fileset()` over a missing
`../crds` silently plans zero resources. The tests passed because the
harness copied a tree releases never ship. Fixing the packaging would have
preserved the deeper flaw: a vendored CRD is frozen at one version while
`chart_version` moves with the user, so the schema silently stops matching
the operator. Helm leaves this hole and every wrapper inherits it; Flux is
the only engine that closed it first-class. This change closes it once for
the catalog, as law the forge teaches to every future kind, with the
mechanics (one shared Pulumi package, one canonical Terraform block) to
follow on the ground proven below.

## The gate: derivation proven before any primitive code

Every claim the derive branch makes was run with tools before the rules were
written, on both engines, from a clean Helm configuration:

- `data "helm_template"` under `hashicorp/helm ~> 3.0` (resolved 3.3.0),
  under both `terraform` 1.14.3 and `tofu` 1.12.5: OpenTelemetry operator
  0.120.0 with `crds.create=true` and `api_versions = ["cert-manager.io/v1"]`
  renders 4 CRDs (`instrumentations`, `opampbridges`,
  `opentelemetrycollectors`, `targetallocators` in `opentelemetry.io`);
  OpenSearch operator 2.8.0 with `installCRDs=true` renders 10 CRDs in
  `opensearch.opster.io`; `planton-operator` 0.7.1 from
  `oci://ghcr.io/plantonhq/charts` yields its one `crds/`-directory CRD
  (`plantonplatforms.planton.ai`) on the `crds` attribute. Templated CRDs
  arrive in `manifests` (map keyed by template path), directory CRDs on
  `crds` — a consumer reads both and filters by kind. OTel's collector CRD
  renders `cert-manager.io/inject-ca-from: gate-otel-ns/gate-otel-serving-cert`
  and a conversion-webhook `clientConfig.service` of `gate-otel-webhook` in
  `gate-otel-ns`: the release-derived values come out correct from the real
  release identity, no token substitution.
- The Helm SDK (`helm.sh/helm/v3 v3.20.2`, the version in `go.mod`):
  `action.NewInstall` with `DryRun`, `ClientOnly`, `IncludeCRDs`, the real
  release name and namespace, `APIVersions` and `KubeVersion` set, returns
  the identical three sets. OCI resolves through the registry client.
- Failure texts are identical across the provider and the SDK (the
  provider is a thin wrapper): version not published →
  `chart "opensearch-operator" version "99.99.99" not found in <repo>
  repository`; repository unreachable → `looks like "<url>" is not a valid
  chart repository or cannot be reached: Get "<url>/index.yaml": dial tcp:
  lookup <host>: no such host`; OCI version not published →
  `failed to perform "FetchReference" on source: <ref>:99.99.99: not found`.
  These are the raw inputs the shared mechanics will wrap in the
  three-part form.
- Solr's per-version CRD bundle: `all-with-dependencies.yaml` for v0.9.1
  carries exactly the 4 CRDs the module needs (three `solr.apache.org` plus
  `zookeeperclusters.zookeeper.pravega.io`). The URL must be the
  `archive.apache.org` form: Apache's `dist` tree answers 404 for every
  release but the current one (v0.9.0 is already gone from `dist`, present
  in `archive`), so a `dist` URL derived from `chart_version` would break
  the day upstream ships the next release.
- One environment fact for the failure matrix: the render consults the
  local Helm repository configuration, and a stale entry there (a listed
  repository whose index cache is missing) fails every render and every
  install identically with `no cached repo found (try 'helm repo update')`.
  Isolating `HELM_REPOSITORY_CONFIG`/`HELM_REPOSITORY_CACHE` reproduces the
  runner's clean state.

## Verification

- `go test ./pkg/anatomy/...` red on exactly the four kinds with the old
  baseline, green after regeneration (baseline +4, nothing else moved).
- `hack/guards/ensure_modules_are_self_contained.sh`: red on exactly eight
  modules with the table disabled, green with it, red on a deliberately
  stale entry; runs under macOS `/bin/bash` 3.2.
- `go vet` and `go test ./e2e/framework/runner/` green;
  `TestKubernetesKeda_Terraform` green on Kind (3 scenarios, 249 s);
  `TestCatalogFixtureIntegrity` and `ensure_e2e_tier_wiring.sh` green with
  the deferred profiles; `planton e2e discover --provider kubernetes` lists
  them as DEFERRED.
- `gofmt` clean; the five workload kinds and the annotation-keys package
  build.
- Nothing published; no release tag.

## What comes next

The shared mechanics of the derive branch: one Go package under
`pkg/iac/pulumi/` and one forge-generated canonical Terraform block with a
byte-identity guard, then the three third-party operator kinds and the
generic `KubernetesHelmRelease` moved onto them, each leaving the anatomy
baseline and the self-containment table as it converts.
