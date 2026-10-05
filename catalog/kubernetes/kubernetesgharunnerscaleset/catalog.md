# GitHub Actions Runner Scale Set

Deploys an autoscaling fleet of self-hosted GitHub Actions runners for ONE GitHub repository, organization or enterprise -- from the official `gha-runner-scale-set` chart (OCI, ghcr.io/actions/actions-runner-controller-charts). The fleet renders an AutoscalingRunnerSet; the controller runs a listener that long-polls GitHub for queued jobs and creates one EPHEMERAL runner pod per job -- each runner executes exactly one job and is replaced. Workflows target the fleet BY NAME (`runs-on: <name>`), authentication is secret-native (an existing Secret reference, or an inline PAT / GitHub App the module materializes into a Secret), and container modes cover Docker builds (dind) and unprivileged container jobs (the Kubernetes hook). Requires a KubernetesGhaRunnerScaleSetController on the cluster first. Uses a Kubernetes Provider Connection for cluster access.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Helm Release** -- the `gha-runner-scale-set` chart, creating:
  - An AutoscalingRunnerSet custom resource the controller reconciles
  - A long-lived LISTENER pod that long-polls GitHub for queued jobs
  - EPHEMERAL runner pods scaling between `minRunners` and `maxRunners` -- one per job, replaced after each job
  - In `kubernetes` container mode, one ephemeral PersistentVolumeClaim per runner from the declared StorageClass
- **GitHub Credential Secret** -- created only when the credential is declared INLINE (`auth.pat` / `auth.githubApp`); the values are materialized into a Kubernetes Secret and never rendered into chart values. With `auth.existingSecretName`, nothing is created -- the chart reads your Secret by name
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **The runner scale set controller** -- a KubernetesGhaRunnerScaleSetController must already run on the cluster; without it, this fleet installs but nothing reconciles it and no runner ever starts. Keep this fleet's `chartVersion` EQUAL to the controller's (GitHub supports only matching pairs).
- **A GitHub credential** -- a classic PAT with `repo` scope (repository registration) or `admin:org` (organization), a fine-grained PAT with the Self-hosted runners read/write permission, or a GitHub App with that permission installed on the target. The recommended posture is a pre-created Secret: key `github_token` for a PAT, or keys `github_app_id` / `github_app_installation_id` / `github_app_private_key` for an App.
- **For `dind` container mode** -- the namespace must allow privileged pods.
- **For `kubernetes` container mode** -- a StorageClass that provisions dynamically (each runner claims an ephemeral work volume per job).

## Deploy

### Console

Open the deployment store, find **GitHub Actions Runner Scale Set**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Repository runners preset** (scale-to-zero for one repo), the **Organization Docker-builds preset** (a dind fleet behind a runner group), or the **Unprivileged Kubernetes-mode preset** (container jobs without privileged pods) in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: build-runners
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "ci-runners"
  createNamespace: true
  githubConfigUrl: https://github.com/acme-corp/api-server
  auth:
    existingSecretName: github-credential
  minRunners: 0
  maxRunners: 10
```

```shell
planton apply -f build-runners.yaml
```

This registers a fleet named `build-runners` for one repository: workflows say `runs-on: build-runners`, runners exist only while jobs run (scale-to-zero), and at most 10 run at once. The credential lives in a Secret you created -- it never rides the manifest. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the fleet to a namespace managed by another Infra Component -- and, in `kubernetes` container mode, to a managed StorageClass:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: ci-runners-namespace
      fieldPath: spec.name
  createNamespace: false
  containerMode:
    mode: kubernetes
    kubernetesWorkVolume:
      storageClass:
        valueFrom:
          kind: KubernetesStorageClass
          name: ci-fast-ssd
          fieldPath: metadata.name
      size: 2Gi
```

The InfraPipeline deploys the namespace and StorageClass first, then provisions the fleet.

## Key Configuration

These are the most important decisions when configuring a GitHub Actions Runner Scale Set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Workflows target the fleet BY NAME** -- The fleet registers in GitHub under `runnerScaleSetName` (empty = this resource's metadata.name; at most 45 characters, GitHub's own limit), and workflows select it with `runs-on: <that name>`. Labels are not how scale sets route. Renaming re-registers the fleet: workflows still saying the old name queue forever.

**The registration scope is the URL's shape** -- `githubConfigUrl` (required) is a repository (`https://github.com/my-org/my-repo`), an organization (`https://github.com/my-org`), or an enterprise (`https://github.com/enterprises/my-enterprise`); GitHub Enterprise Server URLs work the same way. On org/enterprise registrations, `runnerGroup` governs WHICH repositories may use the fleet -- the group must already exist in GitHub.

**Authentication is secret-native, exactly one method** -- `auth` is required with exactly one arm: `existingSecretName` (RECOMMENDED -- the credential never rides a manifest; the Secret's key contract is `github_token` for a PAT or the three `github_app_*` keys for an App), or an inline `pat` / `githubApp` whose sensitive values are materialized into a Secret and never rendered into chart values. A GitHub App is the production posture (fine-grained permissions, expiring installation tokens); a PAT is the quick start.

**Scaling is scale-to-zero by default** -- `minRunners` empty = 0 (runners exist only while jobs run; one pod schedule of cold-start latency per job); `maxRunners` empty = unbounded (queued jobs above the ceiling WAIT in GitHub, they never fail). When both are set, max must be >= min.

**Docker builds need a container mode** -- No `containerMode` = the plain runner: shell/tool jobs only. `dind` runs a privileged Docker-in-Docker sidecar per runner (docker build/run work; the cluster must allow privileged pods). `kubernetes` runs container jobs as separate unprivileged pods via the container hook -- it REQUIRES `kubernetesWorkVolume` (a dynamically-provisioning StorageClass + a per-runner size), and jobs must declare a container. `kubernetes-novolume` is the hook without a shared work volume.

**The runner container is the jobs' budget** -- `runner.image` empty = `ghcr.io/actions/actions-runner:latest` (pin a tag on production fleets; latest changes under you). `runner.resources` is sized for the JOBS, not the agent -- every build the fleet runs inherits these limits.

**Network seams for locked-down clusters** -- `proxy` routes listener and runner egress through corporate proxies (per-scheme URL + an existing credential Secret NAME with `username`/`password` keys; put in-cluster hosts in `noProxy`). `githubServerTls` trusts a private CA towards a self-signed GHES (a CA ConfigMap reference; the runner mount path also sets NODE_EXTRA_CA_CERTS).

**The controller reference is for fenced controllers only** -- Leave `controllerServiceAccount` EMPTY with a cluster-wide controller (the chart auto-discovers it). It is required when the controller was fenced with `watchSingleNamespace` -- wire the name from the controller's `service_account_name` output.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `containerMode.kubernetesWorkVolume.storageClass` | `metadata.name` |
| **KubernetesConfigMap** | `githubServerTls.configMapName` | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the fleet (listener + runner pods) runs in | Operational verification |
| `release_name` | Helm release name (equals metadata.name) | Helm management and debugging |
| `runner_scale_set_name` | The name registered in GitHub -- the exact `runs-on:` value | Workflow YAML configuration |
| `github_config_url` | The GitHub URL the fleet serves | Operational verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Repository runners** -- Scale-to-zero runners for one repository with a pre-created credential Secret. The quick start. Start from the **Repository runners preset**.

**Organization Docker builds** -- An org-wide dind fleet behind a runner group, warm runners for the compounding cold start, and resources sized for builds. Start from the **Organization Docker-builds preset**.

**Unprivileged Kubernetes mode** -- Container jobs without privileged pods via the container hook and a per-runner work volume -- for Pod Security `restricted` and regulated clusters. Start from the **Unprivileged Kubernetes-mode preset**.

## Works With

- [**GitHub Actions Runner Scale Set Controller**](/infra-catalog/kubernetes-gha-runner-scale-set-controller) -- the prerequisite engine that reconciles this fleet into listener and runner pods
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the fleet
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) -- backs the per-runner work volumes in `kubernetes` container mode
- [**Kubernetes ConfigMap**](/infra-catalog/kubernetes-config-map) -- holds the private CA certificate for a self-signed GitHub Enterprise Server
