# Planton Helm Charts

Helm charts for running [Planton](https://planton.ai) on your own Kubernetes
cluster. All charts are published as OCI artifacts to GitHub Container
Registry and are publicly pullable — no registry login, no repo to add:

```
oci://ghcr.io/plantonhq/charts/<chart>
```

> Not to be confused with the repository's [`charts/`](../charts) directory —
> that is the **InfraChart catalog**, Planton's own blueprint format for
> deploying infrastructure *through* Planton. The charts here deploy Planton
> itself.

## Which chart do I want?

| Chart | Use it when |
|---|---|
| [`planton-operator`](planton-operator) | **Always first.** The Planton operator and the `PlantonPlatform` definition it serves; the chart owns the definition's lifecycle, so upgrading the chart upgrades the schema with the operator. One per cluster. |
| [`planton`](planton) | A `PlantonPlatform` as a Helm release with proven defaults and per-cloud values files. Install it after the operator; a few minutes later the full platform is running — console, API, sign-in, an in-cluster deploy runner — reachable over a single `kubectl port-forward` command the install prints for you. Managing the resource yourself (GitOps, custom manifests) works just as well. |
| [`planton-runner`](planton-runner) | Deploy a standalone Planton Runner into a remote cluster with a runner token — the runner enrolls itself with the control plane on first boot and receives its own identity. Normally driven by `planton runner deploy` rather than installed by hand. |

## Installing Planton

Two releases, in this order — the operator brings the definition the
platform resource needs, and Helm validates every resource of a release
against the cluster before creating any of them, so one release cannot both
define the kind and create the platform:

```bash
helm install planton-operator oci://ghcr.io/plantonhq/charts/planton-operator \
  --namespace planton \
  --create-namespace

helm install planton oci://ghcr.io/plantonhq/charts/planton \
  --namespace planton \
  --set platform.spec.version=<release>
```

One value is required on every cluster: `platform.spec.version`, the platform
release to install. The chart pins none of its own, so a chart release never
moves a running platform. The published releases are the
[control-plane image's tags](https://github.com/orgs/plantonhq/packages/container/package/planton%2Fcontrol-plane),
and the operator refuses a release below its floor, naming the floor. The
install output tells you what to do next: watch the platform reach `Ready`,
run the printed port-forward command, and open the console — the first
visitor becomes the administrator. The Planton desktop's guided install does
the same two steps for you: the operator chart, then the platform declared
directly, at the release the desktop shipped with.

### Per-cloud values files

To publish Planton at your own URL and give the runner a cloud identity, use
the recipe values files shipped beside the `planton` chart — they work
straight from a raw GitHub URL:

```bash
helm install planton oci://ghcr.io/plantonhq/charts/planton \
  --namespace planton \
  --set platform.spec.version=<release> \
  --values https://raw.githubusercontent.com/plantonhq/planton/main/helm/planton/values.rke2.yaml
```

| File | For |
|---|---|
| [`values.eks.yaml`](planton/values.eks.yaml) | Amazon EKS |
| [`values.gke.yaml`](planton/values.gke.yaml) | Google GKE |
| [`values.aks.yaml`](planton/values.aks.yaml) | Azure AKS |
| [`values.doks.yaml`](planton/values.doks.yaml) | DigitalOcean Kubernetes |
| [`values.rke2.yaml`](planton/values.rke2.yaml) | RKE2 / Rancher |

Only facts true of every such cluster are active in these files (for example,
RKE2's bundled nginx ingress class); everything specific to your account —
hostnames, IAM roles, certificates — is a precisely annotated comment, one
uncomment away.

## Rules worth knowing

- **One Planton operator per cluster.** The operator watches every namespace,
  so a second installation would fight the first over the same platforms —
  and it refuses to start, with the remedy in its log. The `planton` chart
  can be installed many times beside it, one platform per namespace.
- **The operator chart owns its definitions.** `helm upgrade planton-operator`
  upgrades the `PlantonPlatform` and `PlantonIdentityProvider` schemas with
  the operator; `helm uninstall` keeps them (and every platform) unless you
  set `crds.keep=false`. An install whose definition predates this ownership
  is adopted with two `kubectl` commands the chart prints for you (see the
  operator chart's README).
- **Your data outlives an uninstall.** `helm uninstall` removes the platform
  workloads but deliberately leaves the data services' volumes; delete the
  namespace to reclaim them — and always do so before reinstalling into the
  same namespace (generated credentials die with the release, so a reinstall
  against surviving volumes would not authenticate).

## Versioning and publishing

Every release in this repository is a git tag named by the artifact's
directory and its version, and a `make release-*` front door computes the
next version and pushes the tag; CI derives every artifact's version from the
tag and changes nothing in git.

| Tag | Releases | Front door |
|---|---|---|
| `vX.Y.Z` | the catalog (Terraform and Pulumi modules, InfraCharts) | `make release` |
| `operator/vX.Y.Z` | the operator image `ghcr.io/plantonhq/planton/operator:vX.Y.Z` **and** the `planton-operator` chart as `X.Y.Z` with `appVersion` `vX.Y.Z` — one version line, the chart publishes only after the image is verified | `make release-operator` |
| `helm/<chart>/vX.Y.Z` | the `<chart>` chart as `X.Y.Z` (`planton`, `planton-runner`) | `make release-helm chart=<chart>` |

Each front door bumps the patch version of its own namespace by default
(`bump=minor` or `bump=major` to bump more; `version=vX.Y.Z` to name it). The
chart `version` is stamped at package time
([`release.helm.yaml`](../.github/workflows/release.helm.yaml), `helm package
--version`), so `Chart.yaml` in git carries the development placeholder
`0.0.0-dev` and a checkout is a dev build: install one with `--set
image.tag=<published tag>` where the chart deploys an image. Published
versions are immutable (a re-run of a tag skips what is already published) and
new chart packages are made publicly pullable in the same workflow. Pin what
you install: `helm install ... --version X.Y.Z`.

Two version lines meet here. The operator's own line is the tag above. The
platform images the operator runs (control plane, runner, console) are built
in the platform repository at the version a `PlantonPlatform`'s
`spec.version` names; the `planton` chart's `appVersion` records the platform
version a zero-values install deploys, and the `planton-runner` chart's
`appVersion` the runner image it deploys. Those two `appVersion`s are pins
kept in `Chart.yaml`, moved deliberately when a new platform line publishes,
never stamped from a tag.
