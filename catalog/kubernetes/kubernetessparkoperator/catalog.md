# Spark Operator

Deploys the Apache Spark Kubernetes Operator -- the official ASF controller that turns `SparkApplication` (one batch/streaming job, run to completion) and `SparkCluster` (a long-lived standalone cluster) declarations into running Spark workloads -- from the official `spark-kubernetes-operator` Helm chart. This component installs the ENGINE only: Spark jobs are submitted separately as `SparkApplication` objects, per pipeline run, typically from an orchestrator such as KubernetesAirflow or via KubernetesManifest. One operator per cluster is the normal posture: empty workload namespaces means it watches cluster-wide. The two spark.apache.org CRDs ride the chart's `crds/` directory -- installed once, never upgraded by chart bumps, and KEPT on uninstall along with every Spark workload declaration. There is NO admission webhook and NO cert-manager dependency.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The spark.apache.org CRDs** -- `sparkapplications` and `sparkclusters`, installed by Helm from the chart's `crds/` directory: installed once, never upgraded on chart bumps (apply the new release's CRD files manually when a bump changes them), and left on the cluster on uninstall -- the keep-on-uninstall posture this catalog wants for workload-bearing CRDs
- **Helm Release** -- the `spark-kubernetes-operator` chart, creating:
  - Deployment for the operator with the configured replica count, JVM arguments, and resource limits; with more than one replica the modules set the leader-election properties for you (the chart refuses multi-replica installs without them, by design) and ONE active reconciler leads while the rest stand by warm
  - Release-derived RBAC -- the chart hardcodes its cluster-scoped RBAC names, so the modules derive every RBAC name from the release identity instead; instances coexist on one cluster
  - Under the fenced posture (non-empty `workload.namespaces`): each listed namespace CREATED by the chart, with the workload service account and a namespace-scoped Role planted in each, all marked `helm.sh/resource-policy: keep` so running jobs never lose their identity on uninstall
  - Optionally a ConfigMap for hot property reloading when `dynamicConfig` is enabled
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster-admin-grade permissions on the first install** -- applying the cluster-scoped spark.apache.org CRDs and (under the cluster-wide posture) the workload ClusterRole requires them.
- **A name within budget** -- keep `metadata.name` at 40 characters or fewer: the modules pin the chart's fullname AND every RBAC name to the resource name, and the longest derived suffix (`-config-monitor-binding`, 23 characters) must fit the Kubernetes 63-character cap. Both engines fail loudly over the budget.
- **No cert-manager needed** -- this operator has no admission webhook and no certificate machinery. It validates in its reconcile loop: a bad SparkApplication surfaces on the CR's status, not as an admission rejection.

## Deploy

### Console

Open the deployment store, find **Spark Operator**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default preset** for the standard cluster-wide install, or the **Fenced team-namespaces preset** for the multi-tenant posture (the operator watches only the namespaces you list) in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSparkOperator
metadata:
  name: spark-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: spark-operator
  createNamespace: true
  replicas: 2
  workload:
    namespaces:
      - data-pipelines
      - ml-jobs
    serviceAccount: spark
```

```shell
planton apply -f spark-operator.yaml
```

This deploys the operator with a warm standby behind leader election in the `spark-operator` namespace, fenced to the `data-pipelines` and `ml-jobs` namespaces: the chart creates both, plants the `spark` service account and a namespace-scoped Role in each, and the operator watches ONLY that list. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the operator to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: spark-operator-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the operator into it.

## Key Configuration

These are the most important decisions when configuring the Spark Operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The workload namespaces fence IS the watch scope** -- Where Spark is allowed to run and what the operator watches are ONE chart surface, decided together in `workload.namespaces`. Empty = cluster-wide: the operator watches every namespace and the chart creates a workload ClusterRole. Non-empty = the fenced multi-tenant posture: the chart CREATES each listed namespace, plants the workload service account and a namespace-scoped Role in each, and the operator watches ONLY that list. SparkApplications anywhere else are ignored WITHOUT an error -- a missing namespace in the list looks like a job that never starts. The workload resources are marked `helm.sh/resource-policy: keep`, so the fenced namespaces and their RBAC survive uninstall.

**The CRD lifecycle is upstream's keep-forever posture** -- The chart ships its two CRDs from its `crds/` directory: Helm installs them once, NEVER upgrades them on chart bumps, and LEAVES them (and every Spark workload declaration) on uninstall. The modules neither re-own nor template them -- when a chart bump changes the CRDs, apply the new release's CRD files manually.

**Chart version lockstep** -- `chartVersion` (default `"1.8.0"`) pins the chart, and chart 1.8.0 pairs with operator 1.0.0. The version must exist as a served chart in the upstream repository index; bumps never touch the `crds/`-directory CRDs.

**Replicas are warm standbys** -- A single replica suits most clusters. With `replicas: 2` the modules set the leader-election properties for you (the chart refuses multi-replica installs without them) and ONE active reconciler leads -- reconcile throughput does not change with replicas.

**The workload service account stays the upstream convention** -- `workload.serviceAccount` (default `spark`) is the identity Spark driver/executor pods run as in every workload namespace. SparkApplications reference it by that name -- change it only if your job declarations reference something else.

**Properties, not values, configure the operator** -- `operatorProperties` appends `spark.kubernetes.operator.*` keys over the chart's defaults (the full key catalog ships with the operator's docs at the pinned version). `dynamicConfig` (off by default, the upstream default) adds hot property reloading from a ConfigMap: changes apply WITHOUT an operator restart.

**Sizing is deliberate** -- `resources` empty means the chart defaults (1 CPU / 2Gi, requests = limits); lower for lab clusters consciously -- an OOM-killed operator strands every reconciling Spark job. `jvmArgs` empty means the chart's tuned default (parallel GC, 80% RAM percentage, crash on OOM).

**The image dial covers the operator ONLY** -- `imageRegistry` rewrites the registry part of the operator's own image (`apache/spark-kubernetes-operator`, Docker Hub implied) -- the air-gap path for the operator. It does NOT rewrite the images Spark workloads run: those ride each SparkApplication's own image field.

**The Helm-values escape hatch is unguarded here** -- `helmValues` merges LAST over the typed fields (Helm `-f` semantics, identical on both engines) for the chart surface beyond them: sentinel health canaries, the operator NetworkPolicy, extra RBAC shapes. This chart needs no post-merge re-pins: it has no release-owned CRDs and no webhook machinery whose keys an escape-hatch value could weaponize.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in | Locating the control plane for diagnostics |
| `release_name` | Helm release name (equals metadata.name; the chart's fullname and every RBAC name are pinned to it) | Helm management and debugging |
| `workload_service_account` | Service account Spark driver/executor pods run as in every workload namespace | SparkApplication declarations reference it |
| `watched_namespaces` | Namespaces the operator watches for Spark workloads (empty = cluster-wide) | Verifying a job's namespace is inside the fence |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard operator** -- One operator in its own namespace watching cluster-wide, chart defaults for sizing and images. Start from the **Default preset**.

**Fenced multi-tenant platform** -- The operator watches an explicit namespace list; the chart creates the namespaces and plants workload RBAC in each; a warm standby behind leader election; reconcile cadence tuned via operator properties. Start from the **Fenced team-namespaces preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the operator install
- [**Apache Airflow**](/infra-catalog/kubernetes-airflow) -- the typical orchestrator submitting `SparkApplication` objects per pipeline run against this operator
- [**Kubernetes Manifest**](/infra-catalog/kubernetes-manifest) -- declares standing `SparkApplication` or `SparkCluster` objects outside an orchestrator
