# The Helm/CRD primitive lands in both engines, and the OpenTelemetry operator is the first kind on it — derive from the pinned chart, keep, re-adopt, refuse a downgrade, explain every failure

## What changed

- **One engine-neutral core, `pkg/kubernetes/helmcrds`.** Given a `Source`
  (a chart at a pinned version with the release's OWN values and one CRD
  switch override, or an upstream bundle URL template), `Derive` renders
  the chart client-side through the Helm SDK (or fetches the bundle),
  splits on separator lines, keeps the CustomResourceDefinition documents,
  and stamps each with where it came from: annotations
  `planton.ai/crd-source-chart` and `planton.ai/crd-source-version`, label
  `planton.ai/crd-source=<chart>`. `CheckNoDowngrade` orders versions with
  Helm's own semver and refuses a lower one. Every error is a `Failure`
  with three stable lines: `observed:`, `meaning:`, `next step:`. The raw
  Helm text stays inside the observation, so nothing is hidden from a
  reader who knows Helm. Offline unit tests on fixture charts (templated
  CRDs behind a switch with release-derived webhook values and a mid-line
  `---` in a description; a `crds/` directory; no CRDs; an upstream bundle
  with a doubled separator and a non-CRD document) and on the classified
  Helm errors.
- **The Pulumi half, `pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds`.**
  `Apply` reads the CRDs this source has already stamped on the cluster
  (client-go, through a new `kubeconfig.RESTConfig` that mirrors the
  provider getter's two branches so the read and the provider address the
  same cluster), refuses a downgrade before anything registers, derives,
  and registers one classic-yaml ConfigGroup per CRD keyed by its own name
  on the upsert provider, with the `retainOnDelete` transformation when
  `KeepOnUninstall`. During `pulumi destroy --run-program` it registers
  nothing: a destroy re-runs the program only for delete hooks, and a
  derive failure (a manifest currently pinning an unpublished version)
  must never stand between a user and deleting their stack. The signal is
  `PLANTON_IAC_OPERATION=destroy`, set by the platform's Pulumi runner and
  the e2e harness alike (`iacinput.OperationEnvVar`, `iacinput.IsDestroy`).
- **The Terraform half, generated.** `planton tofu generate-helm-crds`
  writes the canonical `helm_crds.tf`: `data "http"` on the repository
  index (an unpublished version is explained in the module's words beside
  Helm's), `data "helm_template"` with the release's own values list plus
  the CRD switch (`include_crds`, the chart's capability API versions),
  `data "http"` for the bundle branch, `data "kubernetes_resources"` for the
  never-downgrade read by label, the split/filter/stamp locals, and
  `kubectl_manifest.helm_crds` keyed by CRD name with server-side apply,
  `force_conflicts`, and `apply_only = local.helm_crds_keep`. Postconditions
  and preconditions carry the same three-part texts. The file has zero
  per-kind content; a module supplies `helm_release_values` (the list its
  release ALSO consumes), `helm_crds_render_override`,
  `helm_crds_api_versions`, `helm_crds_bundle_url`, `helm_crds_enabled`,
  `helm_crds_keep`. `TestHelmCRDsTFDrift` walks every committed
  `helm_crds.tf` and holds it byte-identical to the generator;
  `TestHelmCRDsTFStampKeysMatchGo` holds the block's stamp literals equal
  to the Go constants. Both run in the new `tf-generated-drift` job of
  `lint.terraform-modules.yaml`, which also brings the existing
  `variables.tf` drift test under CI for the first time.
- **The OpenTelemetry operator is the first kind on the primitive**, both
  twins. Spec: `bool skip_crds = 4` is `reserved 4;` and
  `KubernetesOtelOperatorCrds crds = 14` (`install`, `keep_on_uninstall`,
  both default true) takes its place; stubs, `reference.md`, spec tests
  regenerated. Pulumi: `crds.go` deleted, `main.go` builds the release
  values once and hands them to `keptcrds.Apply`, the release sets
  `SkipCrds`. Terraform: `locals.tf` owns `helm_release_values` and the
  `helm_crds_*` contract, `main.tf` consumes the same list and sets
  `skip_crds = true`, `provider.tf` pins `hashicorp/http`, `helm_crds.tf`
  and `variables.tf` are generated. `iac/crds/` is gone; the kind left
  `pkg/anatomy/baseline.yaml` and the self-containment guard's table. The
  README, catalog page, both module READMEs, import map (rows re-keyed to
  `helm_crds`), permissions (the `delete` verb for the keep-off arm, the
  `list` verb for the version read), and controls read as if derivation
  always existed.
- **The e2e harness learned the second act.** Three scenario annotations
  extend the lifecycle: `planton.dev/e2e-upgrade-manifest` (UPGRADE ->
  VERIFY-UPGRADED against the second manifest),
  `planton.dev/e2e-expect-upgrade-failure` (UPGRADE-EXPECT-FAIL, the first
  manifest restored before destroy), `planton.dev/e2e-reinstall`
  (REINSTALL -> VERIFY-REINSTALLED -> DESTROY-AGAIN -> VERIFY-CLN-AGAIN).
  The second-act manifest carries `planton.dev/e2e-second-act` so
  discovery never runs it as a lane. `runValidate` and the lanes share one
  `bindManifest`. The Kubernetes harness implements the framework's
  `DeployFailureVerifier` and dispatches to a kind's
  `verify.DeployFailureVerifier`. The Kubernetes `TestMain` isolates
  Helm's repository configuration per run (`runner.IsolateHelmEnvironment`)
  so a laptop's stale `helm repo` entry can never fail a lane. Documented
  in `e2e/README.md`.
- **The OTel verifier proves the lifecycle.** It reads `chartVersion`,
  `crds.keepOnUninstall`, `crds.install`, and the expect-deploy-failure
  annotation from the manifest; asserts every CRD's source stamp equals the
  manifest's version (so VERIFY-UPGRADED is "the stamp moved"); asserts
  kept-or-deleted on destroy as declared; and pins the two refusal classes
  (`chart-version-not-published`, `crd-schema-downgrade`) to the three-part
  text with the engines' line-wrapping normalized, checking after a refused
  downgrade that no CRD changed. Five scenarios: `minimal` (keep, reinstall
  re-adopts), `full-surface` (keep off, cleanup), `upgrade` (0.120.0 to
  0.120.3), `version-not-published`, `downgrade-refused` (0.120.3 to
  0.119.0).
- **Convergence in passing.** `pulumikubernetesprovider` reads `KUBE_CTX`
  through the new `kubeconfig.KubeContextEnvVar` constant instead of a
  literal; `Masterminds/semver/v3` is a direct dependency; `MODULE.bazel`
  exposes `io_k8s_client_go`, `io_k8s_apimachinery`,
  `com_github_masterminds_semver_v3`.

## Why

Helm installs a chart's `crds/` directory once and never upgrades or
removes it, and templated CRDs are release-owned and die with the release.
Every IaC wrapper inherits the hole. The catalog's four kinds that answered
it by staging a CRD copy beside the module had never installed from a
published release, and even the copy would have been frozen at one version
while `chart_version` moved with the user. The decision tree became law
earlier today; this change is its third branch made real, once per engine,
and proven on the hardest kind in the catalog: a chart whose CRDs are
templated from the release's identity and carry a conversion webhook whose
trust cert-manager keeps. Two facts the build turned up are now part of the
design: the render must see the release's FULL values with only the CRD
switch flipped (the OTel CRD templates depend on `fullnameOverride`, the
cert-manager arm, and the webhook port; a minimal render points the CRDs at
the wrong webhook), and a destroy must never be blocked by a derive
failure.

## Verified live (Kind, both engines)

Terraform, five scenarios green: install and keep, reinstall re-adopting
the kept CRDs (REINSTALL, VERIFY-REINSTALLED, DESTROY-AGAIN, VERIFY-CLN-AGAIN
all PASS); keep off, four CRDs deleted on destroy; upgrade 0.120.0 to
0.120.3 with every CRD's stamp moving and the webhook proofs passing after
the bump; the unpublished version refused at plan with the three-part text
beside Helm's own; the downgrade refused at upgrade with the cluster's
0.120.3 and the manifest's 0.119.0 named, no CRD changed. Pulumi, the same
five green, with the in-process render and the client-go downgrade read;
the refused-deploy lane's destroy proceeds because the program stands
aside during destroy. The canonical block was also driven directly against
a cluster before the kind moved: idempotent re-plan, upgrade, refusal,
destroy-keeps, reinstall-adopts, and the Solr bundle branch fetching
exactly four CRDs from `archive.apache.org` with a clean 404 explanation
for an unpublished version.

Offline: `go test` for `helmcrds`, `keptcrds`, `kubeconfig`,
`pkg/iac/tofu/generators` (drift and stamp-key tests), `pkg/anatomy`,
`pkg/explain`, `pkg/protodocs`, `pkg/catalogpage`, `pkg/iac/permissions`,
`e2e/framework/runner`, the OTel spec tests; `tofu fmt -check` and
`tofu validate` on the OTel module; `bazel build` of every new and changed
Go target; the self-containment and provider-pin guards.

## What comes next

Solr (bundle branch) and OpenSearch (render branch) onto the primitive,
then a catalog release; the generic `KubernetesHelmRelease` with its
`crds` message and the templated-CRD design question; the wiki article that
carries the tree and the failure matrix.
