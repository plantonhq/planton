# GitHub Actions Runner Scale Set Controller

Deploys the GitHub Actions Runner Scale Set Controller -- GitHub's official actions-runner-controller (ARC) manager -- from the official `gha-runner-scale-set-controller` chart (OCI, ghcr.io/actions/actions-runner-controller-charts). This component installs the ENGINE only: runner fleets are declared separately as KubernetesGhaRunnerScaleSet resources (one per repository/organization/enterprise registration), and the controller reconciles them into listener pods and ephemeral runner pods -- each runner executes exactly one job and is replaced. One cluster-wide controller is the sane default: it watches all namespaces, so every runner scale set on the cluster is served by it. The controller holds no GitHub credential -- each fleet brings its own.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Helm Release** -- the `gha-runner-scale-set-controller` chart, creating:
  - Deployment for the controller manager with the configured replica count and resource limits; with more than one replica the chart enables leader election automatically (extra replicas are hot standbys, not workload shards)
  - ServiceAccount with RBAC permissions to manage the runner custom resources -- its name is exported as `service_account_name`
  - The `actions.github.com` CRDs (AutoscalingRunnerSet, EphemeralRunner, ...) from the chart's `crds/` directory -- Helm installs them on FIRST install and never removes them afterwards (see the destroy contract under Key Configuration)
  - Metrics wiring on the controller args and every listener pod, only when `metrics` is declared
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster-admin-grade permissions on the first install** -- the chart's first install creates the cluster-scoped `actions.github.com` CRDs. Later installs adopt CRDs already on the cluster cleanly (they carry no release ownership metadata).
- **Outbound network access to GitHub** from the namespaces where listener pods will run -- listeners LONG-POLL GitHub for queued jobs (no inbound webhooks are required).

## Deploy

### Console

Open the deployment store, find **GitHub Actions Runner Scale Set Controller**, and click **Deploy**. The creation wizard walks you through placement, the chart version, replicas and resources, behavior flags, metrics, and scheduling. Start from the **Standard controller preset** for the one-cluster-wide-controller shape, or the **Production controller preset** for a hardened control plane (hot standby, `eventual` upgrades, probes, metrics, eviction priority) in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "arc-system"
  createNamespace: true
  replicas: 2
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 512Mi
  flags:
    logLevel: info
    updateStrategy: eventual
    priorityClassName: system-cluster-critical
```

```shell
planton apply -f arc-controller.yaml
```

This deploys the controller with a hot standby behind automatic leader election in the `arc-system` namespace, structured production logging, upgrades that wait for running jobs, and eviction priority so node pressure evicts workloads before the thing that schedules your CI. The controller watches all namespaces. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the controller to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: arc-system-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the controller into it.

## Key Configuration

These are the most important decisions when configuring the GitHub Actions Runner Scale Set Controller. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The CRD lifecycle and the destroy contract** -- The chart ships the `actions.github.com` CRDs in its `crds/` directory: Helm installs them once and NEVER removes them. Destroying the controller KEEPS the CRDs -- and any declared runner scale set objects -- on the cluster, and a later controller install adopts the kept CRDs cleanly (they carry no release ownership metadata). Teardown is a clean recovery path, not a data-loss event. Still destroy KubernetesGhaRunnerScaleSet resources FIRST: without a controller they stop reconciling (queued jobs wait in GitHub while nothing scales).

**Chart version lockstep** -- `chartVersion` (default `"0.14.2"`) pins the chart, and chart and controller image move in lockstep. GitHub supports controller and scale-set charts only at MATCHING versions -- keep every KubernetesGhaRunnerScaleSet's `chartVersion` equal to this one, and bump them together.

**Watch scope and multi-tenancy** -- By default the controller watches ALL namespaces: one install serves every runner scale set on the cluster, which is what almost everyone wants. `flags.watchSingleNamespace` fences the controller to one namespace for hard multi-tenancy -- runner scale sets outside the fence are silently ignored, and every scale set the fenced controller serves must then name its ServiceAccount explicitly (the scale set kind's `controller_service_account`, wired from this kind's `service_account_name` output) because auto-discovery cannot see a fenced controller.

**Replicas are hot standbys** -- A single replica suits most clusters. With more than one, the chart enables leader election automatically: extra replicas take over on failure, they never share the reconcile workload. For runner throughput, raise `flags.runnerMaxConcurrentReconciles` (default 2) instead -- it trades API-server and GitHub-API load for runner startup parallelism.

**Logging ships at debug** -- `flags.logLevel` left empty defers to the chart, and the chart default is `debug`: full reconcile detail on every scaling decision -- the right dial while bringing the install up, and a noisy one once fleets are busy. Production clusters usually set `info`. `flags.log_format: json` feeds structured pipelines.

**Metrics are declare-to-enable** -- An absent `metrics` block IS the disabled chart default. Declaring it enables metrics and then all three addresses are required (`controllerManagerAddr`, `listenerAddr`, `listenerEndpoint`) -- the chart wires them into the controller args and every listener pod. The listener metrics (queued vs started jobs) are the fleet's operational truth.

**Upgrades while jobs run** -- `flags.updateStrategy` empty defers to the chart default `immediate` (recreate listeners at once; may briefly overprovision runners). `eventual` tears down, then waits for running jobs to finish before recreating -- no overprovisioning, slower rollout. Setting `flags.healthProbeBindAddress` ADDS liveness/readiness probes (the chart default is probes off), and `flags.priority_class_name: system-cluster-critical` keeps the runner control plane alive when node pressure evicts workloads.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the controller runs in | Locating the control plane for diagnostics |
| `release_name` | Helm release name (equals metadata.name) | Helm management and debugging |
| `service_account_name` | The controller's ServiceAccount name | A fenced scale set's `controller_service_account` reference |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard controller** -- One cluster-wide controller in its own namespace: it watches all namespaces and serves every runner scale set on the cluster. Start from the **Standard controller preset**.

**Production control plane** -- A hot standby behind leader election, `eventual` upgrades that wait for running jobs, structured logs at info, health probes, metrics on the controller and every listener, and `system-cluster-critical` eviction priority. Start from the **Production controller preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the controller install
- [**GitHub Actions Runner Scale Set**](/infra-catalog/kubernetes-gha-runner-scale-set) -- the runner fleets this controller reconciles, one per repository/organization/enterprise registration
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- scrapes the controller and listener metrics endpoints when `metrics` is declared
