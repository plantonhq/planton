# Kubernetes KubeRay Operator

## When NOT to Use This

**This component installs the ENGINE, not a Ray cluster.** The KubeRay
operator reconciles `RayCluster` custom resources — declared with
KubernetesRayCluster — into running Ray clusters (and, unmodeled in
this catalog, RayJob/RayService CRs authored directly). Install the
operator once per Kubernetes cluster, then declare Ray clusters
against it.

Also not the right kind when:

- **You want a Ray cluster** — that is KubernetesRayCluster; this
  component is the controller that reconciles it.
- **You expect admission-time validation** — the operator has NO
  admission webhook, no certificate machinery, and no cert-manager
  dependency. It validates in its reconcile loop: a bad RayCluster
  surfaces on the CR's status conditions, not as an admission
  rejection.
- **You expect chart upgrades to upgrade the CRDs** — the three ray.io
  CRDs ride the chart's `crds/` directory; Helm installs them once and
  never upgrades them (see below).

## Overview

**KubernetesKubeRayOperator** installs the KubeRay operator from the
official `kuberay-operator` Helm chart
(https://ray-project.github.io/kuberay-helm). Chart 1.6.2 (the pinned
default) pairs with operator v1.6.2. One operator per cluster is the
normal posture: it watches every namespace by default and runs leader
election out of the box.

**Key design points:**

- **The CRDs are kept on uninstall — by upstream design.** The chart
  ships its three ray.io CRDs (`rayclusters`, `rayjobs`,
  `rayservices`) from its `crds/` directory: Helm installs them once,
  NEVER upgrades them on chart upgrades, and LEAVES them (and every
  Ray declaration) on uninstall. The modules neither re-own nor
  template them — apply the new release's CRD files manually when a
  chart bump changes them. The CRDs are large (~1MB each): they
  install server-side.
- **No webhook, no cert-manager.** The operator validates in its
  reconcile loop — no admission webhook, no certificate machinery, and
  no cert-manager dependency.
- **Feature gates merge over the chart's defaults — because Helm
  lists REPLACE.** The chart's `featureGates` value is a LIST, and
  Helm lists replace rather than merge: rendering only your entries
  would silently DROP every chart-default gate. The modules therefore
  render the FULL list whenever you flip any gate — the chart's five
  defaults at 1.6.2 (`RayClusterStatusConditions`,
  `RayJobDeletionPolicy`, `RayMultiHostIndexing` ON;
  `RayServiceIncrementalUpgrade`, `RayCronJob` OFF) overridden by name
  from the spec, then your unknown gates appended. Only list gates you
  are deliberately flipping.
- **The operator image mirror and Ray cluster images are DIFFERENT
  seams.** `image_registry` replaces only the registry part of the
  operator's own image (`quay.io/kuberay/operator`) — the air-gap path
  for the one image THIS component's pods pull. Ray CLUSTER images
  ride each KubernetesRayCluster's own image field; mirroring the
  operator does nothing for them.
- **Name pins keep instances distinguishable.** The chart hardcodes
  its fullname, name, AND service account to `kuberay-operator` by
  default — every install would collapse onto the same child names.
  The modules re-pin all three to this resource's name. Keep
  `metadata.name` at 47 characters or fewer: the longest derived child
  name suffixes `-leader-election` (16 chars) and Kubernetes caps
  names at 63; both engines fail loudly over the budget.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines): operator env-var feature
  flags, logging encoders, single-namespace RBAC shapes, the
  Kubernetes-proxy dialing mode — a safety valve, never the primary
  interface. This chart needs no post-merge re-pins.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install the operator into —
  literal or a KubernetesNamespace reference (`create_namespace` to
  own it)

### Common

- **`spec.chart_version`**: chart pin (default `1.6.2` — pairs with
  operator v1.6.2; bumps never touch the `crds/`-directory CRDs)
- **`spec.watch_namespaces`**: namespaces the operator watches for Ray
  CRs; empty = every namespace (the normal posture). With a list set,
  Ray declarations OUTSIDE these namespaces are ignored without an
  error — the fenced multi-tenant posture; the chart scopes the
  per-namespace reconcile RBAC to the same list
- **`spec.leader_election_enabled`**: default true (the chart
  default — safe for single replicas, required for standbys); disable
  only in constrained RBAC environments that cannot grant lease
  permissions
- **`spec.batch_scheduler`**: gang-scheduling integration —
  `volcano`, `yunikorn`, or `scheduler-plugins`; the named scheduler
  must already run on the cluster (the operator only emits its
  scheduling directives)
- **`spec.feature_gates`**: gates you are deliberately flipping (see
  the list-replacement truth above)
- **`spec.metrics_enabled`**: operator control-plane metrics on port
  8080 (default true)
- **`spec.service_monitor_enabled`**: ServiceMonitor for Prometheus
  discovery (requires the monitoring.coreos.com CRDs — deploy
  KubernetesKubePrometheusStack first; the install FAILS without them)
- **`spec.resources`**: operator container resources — empty = the
  chart defaults (100m CPU / 512Mi limits; upstream sizes ~500MB per
  500 managed Ray pods — scale memory with fleet size)
- **`spec.image_registry`**: the operator-image mirror seam (see
  above)
- **`spec.helm_values`**: the escape hatch (see above)

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the operator runs in |
| `release_name` | Helm release name (= `metadata.name`; the chart's fullname is pinned to it, so child names derive from it) |
| `watched_namespaces` | Namespaces the operator watches for Ray CRs (empty = cluster-wide — RayCluster declarations reconcile anywhere) |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **No prerequisites**: no webhook means no cert-manager dependency —
  the operator installs standalone.
- **KubernetesRayCluster resources depend on this component**: the
  operator must be running before their `RayCluster` resources
  reconcile. Under the fenced posture, declarations outside
  `watched_namespaces` are ignored without an error.
- **Watch namespaces must already exist** — they are a watch scope,
  deliberately not this module's resources.
- **The install is deliberately blocking**: the Helm release waits for
  the operator to become Available (atomic, 600s timeout), so an
  unpullable image, a missing ServiceMonitor CRD, or a broken config
  fails THIS apply with a readiness timeout instead of surfacing later
  as RayClusters that mysteriously never reconcile.

## Examples

### Standard install

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubeRayOperator
metadata:
  name: kuberay-operator
spec:
  namespace:
    value: ray-system
  createNamespace: true
```

### Fenced watch with a feature-gate flip

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubeRayOperator
metadata:
  name: kuberay-operator
spec:
  namespace:
    value: ray-system
  createNamespace: true
  watchNamespaces:
    - ml-team
    - data-team
  featureGates:
    - name: RayServiceIncrementalUpgrade
      enabled: true
```

### Private-mirror operator image

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubeRayOperator
metadata:
  name: kuberay-operator
spec:
  namespace:
    value: ray-system
  createNamespace: true
  imageRegistry: mirror.example.com
  metricsEnabled: true
  serviceMonitorEnabled: true
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
