# Helm Release

Installs an upstream Helm chart as a REAL Helm release — the catalog's sole intentional passthrough. Both engines perform an actual `helm install`: hooks run, the release secret is written, and `helm list` shows the release exactly as if installed by the Helm CLI. Reach for this only when the catalog has no first-class kind for the chart you need: typed kinds validate their configuration before deploy, export composable outputs, and teach their trade-offs field by field — a generic chart install does none of that.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm Release** -- the chart fetched from `repo` (an HTTPS repository index or an `oci://` registry pull), installed at the PINNED `version` into `namespace`, named `releaseName` when set (otherwise `metadata.name`)
- **Everything the chart renders** -- workloads, Services, ConfigMaps, and any CRDs the chart templates as release resources (Helm-owned; a chart that templates them without `helm.sh/resource-policy: keep` is refused unless `crds.allowHelmManaged` accepts it)
- **The CRDs in the chart's `crds/` directory** -- derived from the pinned chart and applied by the module itself outside the release, keyed by CRD name: a `version` bump moves them with it, destroy keeps them (unless `crds.keepOnUninstall` is false), a reinstall re-adopts them, and a version below what the cluster's CRDs carry is refused before anything changes
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true, and deleted with the resource

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- For a **private chart source**: an organization secret holding the repository password or registry token — `repositoryPassword` stores a `$secret/<name>` reference, resolved just-in-time at deploy.

### Kubernetes Cluster

- The chart repository or OCI registry must be reachable from the execution environment.
- If `createNamespace` is false, the target namespace must already exist.
- Any operators or CRDs the CHART's resources depend on (but the chart does not itself ship) must be in place first.

## Deploy

### Console

Open the deployment store, find **Helm Release**, and click **Deploy**. The creation wizard walks you through placement, the pinned chart coordinates (with OCI-vs-HTTPS teaching), private-source access, the values file, the three `--set` override layers (with reference-only pickers for sensitive values), install behavior (atomic vs skip-await), the CRD lifecycle and dependencies, upgrade behavior (reuse vs reset, rollback history), and the advanced escape dials. Start from the **HTTPS Repo Chart** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHelmRelease
metadata:
  name: podinfo
  org: acme-corp
  env: prod
spec:
  namespace:
    value: apps
  createNamespace: true
  repo: https://stefanprodan.github.io/podinfo
  chart: podinfo
  version: 6.9.2
  valuesYaml: |
    replicaCount: 2
    resources:
      requests:
        cpu: 100m
  setString:
    image.tag: "6.9.2"
  atomic: true
```

```shell
planton apply -f helm-release.yaml
```

This installs the podinfo chart at the pinned 6.9.2 into the `apps` namespace as a real Helm release, with two replicas and an atomic rollback guarantee. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the release to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: apps-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then installs the release into it.

## Key Configuration

These are the most important decisions when configuring a Helm release. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The version is pinned on purpose** -- an unpinned "latest" install is not reproducible, and reproducibility is the point of declaring the release here. Editing the version on a deployed release IS the `helm upgrade`.

**The values model is Helm's own** -- `valuesYaml` is the values file (applied first), and `set`, `setString`, `setSensitive` are the `--set` layers applied on top, in that order. Both engines merge the layers with identical precedence and hand Helm ONE final values map — the same manifest installs byte-identical releases on either engine.

**The coercion boundary** -- `set` values get Helm's coercion ("true"/"false" become booleans, digits become numbers, null removes the key); a version-like value that must STAY a string ("1.30") belongs in `setString`.

**Secrets never enter the spec** -- `setSensitive` values and `repositoryPassword` are `$secret/<name>` references to organization secrets, resolved at deploy time and kept out of rendered plans and state where each engine supports it.

**Atomic vs skip-await** -- atomic rolls a failed install/upgrade back whole (it must WAIT for readiness to know whether to roll back), skip-await returns as soon as Helm records the release; the spec rejects the pair. `waitForJobs` extends the readiness wait to chart Jobs and is ignored while the wait is skipped.

**Upgrades are declarative by default** -- the values declared here are the whole story on every upgrade. `reuseValues` merges over the previous release's values instead, `resetValues` starts from chart defaults — mutually exclusive. `maxHistory` bounds rollback history (0 keeps everything).

**Adoption is engine-asymmetric** -- `takeOwnership` (adopt matching resources an earlier tool created) is honored by the Terraform provisioner only; the Pulumi engine's pinned SDK rejects the flag loudly rather than silently ignoring it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | The namespace the release is installed in | Debugging and composition |
| `release_name` | The Helm release name (as shown by `helm list`) — `spec.release_name` when set, otherwise `metadata.name` | Addressing the release's derived objects |
| `version` | The installed chart version | Auditing which chart bytes run |
| `app_version` | The chart's appVersion — the upstream application version the chart packages | Auditing which software version runs |
| `status` | The release status as Helm records it (e.g. `deployed`) | Health checks and composition gates |
| `revision` | The release revision (1 on first install, incremented by each upgrade or rollback) | Change tracking |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public HTTPS chart** -- a pinned chart from a public repository with a values document — the everyday shape. Start from the **HTTPS Repo Chart** preset.

**OCI registry chart** -- an `oci://` reference pulled like a container image, with targeted `set`/`setString` overrides. Start from the **OCI Registry Chart** preset.

**Private repository with secrets** -- repository credentials and sensitive chart values, all as organization-secret references; atomic install for rollback safety. Start from the **Private Repo With Secrets** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the placement target; reference one by name or compose it as a first-class resource.
- First-class chart kinds -- when the catalog grows a typed kind for a chart you run through this kind ([**Cert Manager**](/infra-catalog/kubernetes-cert-manager), [**Ingress NGINX**](/infra-catalog/kubernetes-ingress-nginx), [**Istio**](/infra-catalog/kubernetes-istio), and the rest already exist), prefer it: validation, outputs, and field-level teaching come with it.
