# KubeRay Operator

Deploys the KubeRay operator -- the controller that turns `RayCluster` custom resources (declared with KubernetesRayCluster) into running Ray clusters -- from the official `kuberay-operator` Helm chart. This component installs the ENGINE only: Ray clusters are declared separately as KubernetesRayCluster resources, each carrying its own head/worker topology, autoscaling, and fault tolerance. One operator per cluster is the grain: it watches every namespace by default and runs leader election out of the box. The three ray.io CRDs (`rayclusters`, `rayjobs`, `rayservices`) ride the chart's `crds/` directory -- installed once, never upgraded by chart bumps, and KEPT on uninstall along with every Ray declaration. There is NO admission webhook and NO cert-manager dependency.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The ray.io CRDs** -- `rayclusters`, `rayjobs`, and `rayservices`, installed by Helm from the chart's `crds/` directory: installed once, never upgraded on chart bumps (apply the new release's CRD files manually when a bump changes them), and left on the cluster on uninstall. They are large (~1MB each) and install SERVER-SIDE.
- **Helm Release** -- the `kuberay-operator` chart, creating:
  - Deployment for the operator with the configured resources, feature gates, and (optionally) a batch-scheduler integration; leader election on by default
  - Re-pinned names -- the chart hardcodes its fullname, name, and service account to `kuberay-operator`; the modules re-pin all three to this resource's name so instances stay distinguishable
  - Under the fenced posture (non-empty `watchNamespaces`): the per-namespace reconcile RBAC scoped to exactly that list
  - A ServiceMonitor for the operator's own metrics, only when `serviceMonitorEnabled` is `true`
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster-admin-grade permissions on the first install** -- applying the cluster-scoped ray.io CRDs requires them; the CRDs install server-side because of their ~1MB size.
- **A name within budget** -- keep `metadata.name` at 47 characters or fewer: the modules re-pin the chart's fullname, name, and service account to the resource name, and the longest derived child-name suffix (`-leader-election`, 16 characters) must fit the Kubernetes 63-character cap. Both engines fail loudly over the budget.
- **A batch scheduler, if you name one** -- `volcano`, `yunikorn`, or `scheduler-plugins` must ALREADY run on the cluster; the operator only emits its scheduling directives.
- **The monitoring.coreos.com CRDs, if `serviceMonitorEnabled`** -- deploy KubernetesKubePrometheusStack first; the install FAILS without them, by upstream design rather than a silent skip.
- **No cert-manager needed** -- this operator has no admission webhook and no certificate machinery. It validates in its reconcile loop: a bad RayCluster surfaces on the CR's status conditions, not as an admission rejection.

## Deploy

### Console

Open the deployment store, find **KubeRay Operator**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default preset** for the standard cluster-wide install, or the **Private-mirror preset** for the air-gapped posture (the operator image mirrored, metrics and the ServiceMonitor on) in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubeRayOperator
metadata:
  name: kuberay-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "ray-system"
  createNamespace: true
  watchNamespaces:
    - ml-team
    - data-team
```

```shell
planton apply -f kuberay-operator.yaml
```

This deploys the operator into the `ray-system` namespace, fenced to the `ml-team` and `data-team` namespaces: the operator watches ONLY that list and its per-namespace reconcile RBAC scopes to the same list. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the operator to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: ray-system-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the operator into it.

## Key Configuration

These are the most important decisions when configuring the KubeRay Operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The watch scope is a fence that does NOT create namespaces** -- `watchNamespaces` empty = every namespace, the normal one-operator-per-cluster posture. Non-empty = the fenced multi-tenant posture: Ray declarations OUTSIDE the list are ignored WITHOUT an error (a missing namespace looks like a cluster that never starts), and the chart scopes the per-namespace reconcile RBAC to the same list. Unlike the Spark operator's workload fence, this chart does not create the listed namespaces -- they must already exist.

**The CRD lifecycle is upstream's keep-forever posture** -- The chart ships its three CRDs from its `crds/` directory: Helm installs them once, NEVER upgrades them on chart bumps, and LEAVES them (and every Ray declaration) on uninstall. The modules neither re-own nor template them -- when a chart bump changes the CRDs, apply the new release's CRD files manually, server-side (~1MB each).

**Chart version lockstep** -- `chartVersion` (default `"1.6.2"`) pins the chart, and chart 1.6.2 pairs with operator v1.6.2. The version must exist as a served chart in the upstream repository index; bumps never touch the `crds/`-directory CRDs.

**Feature gates MERGE, because Helm lists REPLACE** -- The chart's `featureGates` value is a list, and Helm lists replace rather than merge: rendering only your entries would silently drop every chart-default gate. The modules render the FULL list whenever you flip any gate -- the chart's five defaults at 1.6.2 (`RayClusterStatusConditions`, `RayJobDeletionPolicy`, `RayMultiHostIndexing` ON; `RayServiceIncrementalUpgrade`, `RayCronJob` OFF) overridden by name from your spec, then unknown gates appended. Only list gates you are deliberately flipping.

**Gang scheduling is opt-in and prerequisite-bound** -- `batchScheduler` integrates `volcano`, `yunikorn`, or `scheduler-plugins` for all-or-nothing placement of Ray pods -- the cure for half-scheduled clusters holding GPUs while waiting for peers that never fit. The named scheduler must already run on the cluster; the operator only emits its scheduling directives.

**Leader election defaults on** -- safe for single replicas and required for standbys. Disable only in constrained RBAC environments that cannot grant lease permissions -- and then never run more than one replica.

**Sizing scales with the FLEET** -- `resources` empty means the chart defaults (100m CPU / 512Mi limits). Upstream sizes ~500MB per 500 managed Ray pods: memory scales with how many Ray pods the operator manages, not request rate. An OOM-killed operator stops reconciling every Ray cluster on its watch.

**The image dial covers the operator ONLY** -- `imageRegistry` rewrites the registry part of the operator's own image (`quay.io/kuberay/operator`) -- the air-gap path for the one image this component's pods pull. Ray CLUSTER images ride each KubernetesRayCluster's own image field; mirroring the operator does nothing for them.

**The Helm-values escape hatch is unguarded here** -- `helmValues` merges LAST over the typed fields (Helm `-f` semantics, identical on both engines) for the chart surface beyond them: operator env-var feature flags, logging encoders, single-namespace RBAC shapes, the Kubernetes-proxy dialing mode. This chart needs no post-merge re-pins: it has no release-owned CRDs and no webhook machinery.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in | Locating the control plane for diagnostics |
| `release_name` | Helm release name (equals metadata.name; the chart's fullname is pinned to it, so child names derive from it) | Helm management and debugging |
| `watched_namespaces` | Namespaces the operator watches for Ray CRs (empty = cluster-wide) | Verifying a Ray cluster's namespace is inside the fence |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard operator** -- One operator in its own `ray-system` namespace watching every namespace, leader election on, control-plane metrics on, chart defaults for sizing. Start from the **Default preset**.

**Air-gapped operator** -- The operator image mirrored to a private registry, metrics explicitly on, and the ServiceMonitor rendered for Prometheus discovery. Start from the **Private-mirror preset** -- and mirror the Ray images on each KubernetesRayCluster declaration too; the operator mirror does nothing for them.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the operator install
- [**Ray Cluster**](/infra-catalog/kubernetes-ray-cluster) -- the Ray clusters this operator reconciles, declared one per cluster with their own topology and lifecycle
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- scrapes the operator's metrics when `serviceMonitorEnabled` is set (and provides the CRDs the ServiceMonitor requires)
