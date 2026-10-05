# Kubernetes Spark Operator

## When NOT to Use This

**This component installs the ENGINE, not a Spark job.** The Apache
Spark Kubernetes Operator reconciles `SparkApplication` (one
batch/streaming job, run to completion) and `SparkCluster` (a
long-lived standalone cluster) custom resources into running Spark
workloads. Install the operator once per cluster, then submit jobs
against it.

Also not the right kind when:

- **You want to declare a Spark job** — `SparkApplication` objects are
  per-JOB, run-to-completion declarations: submit them per pipeline
  run, typically from an orchestrator such as KubernetesAirflow, or
  declare them via KubernetesManifest. They are deliberately not a
  catalog kind.
- **You expect admission-time validation** — this operator has NO
  admission webhook, no certificate machinery, and no cert-manager
  dependency. It validates in its reconcile loop: a bad
  SparkApplication surfaces on the CR's status, not as an admission
  rejection.
- **You expect chart upgrades to upgrade the CRDs** — the two
  spark.apache.org CRDs ride the chart's `crds/` directory; Helm
  installs them once and never upgrades them (see below).

## Overview

**KubernetesSparkOperator** installs the Apache Spark Kubernetes
Operator — the official ASF controller for Spark workloads on
Kubernetes — from the official `spark-kubernetes-operator` Helm chart
(https://apache.github.io/spark-kubernetes-operator). Chart 1.8.0
(the pinned default) pairs with operator 1.0.0. One operator per
cluster is the normal posture: it watches cluster-wide by default.

**Key design points:**

- **The CRDs are kept on uninstall — by upstream design.** The chart
  ships its two spark.apache.org CRDs (`sparkapplications`,
  `sparkclusters`) from its `crds/` directory: Helm installs them
  once, NEVER upgrades them on chart upgrades, and LEAVES them (and
  every Spark workload declaration) on uninstall. That posture is
  exactly the keep-on-uninstall this catalog wants for
  workload-bearing CRDs, so the modules neither re-own nor template
  them — apply the new release's CRD files manually when a chart bump
  changes them.
- **No webhook, no cert-manager.** The operator validates in its
  reconcile loop — there is no admission webhook and no certificate
  machinery, so there is one less lifecycle to manage and nothing that
  can fail-close the cluster's write path.
- **The workload namespaces fence IS the watch scope.** Where Spark is
  allowed to run and what the operator watches are ONE chart surface,
  decided together. Empty `workload.namespaces` = cluster-wide: the
  operator watches every namespace and the chart creates a workload
  ClusterRole. Non-empty = the fenced, multi-tenant posture: the chart
  CREATES each listed namespace, plants the workload service account
  and a namespace-scoped Role in each, and the operator watches ONLY
  that list — SparkApplications anywhere else are ignored without an
  error. The chart marks workload resources
  `helm.sh/resource-policy: keep`, so the fenced namespaces and their
  RBAC survive uninstall (running jobs must not lose their identity
  mid-flight).
- **RBAC names are release-derived because the chart hardcodes them.**
  The chart hardcodes all its cluster-scoped RBAC names as plain
  values (`spark-operator-clusterrole`, …) — a second install anywhere
  on the cluster would collide by construction. The modules derive
  every RBAC name from the release identity instead, so instances
  coexist. The workload SERVICE ACCOUNT deliberately stays the
  upstream convention (`spark` unless overridden): SparkApplications
  reference it by that name.
- **Keep `metadata.name` at 40 characters or fewer.** The modules pin
  the chart's fullname AND every RBAC name to the resource name; the
  longest derived suffix is `-config-monitor-binding` (23 chars)
  against the Kubernetes 63-character cap. Both engines fail loudly
  over the budget.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines): sentinel health canaries, the
  operator NetworkPolicy, extra RBAC shapes — a safety valve, never
  the primary interface. This chart needs no post-merge re-pins: it
  has no release-owned CRDs and no webhook machinery whose keys an
  escape-hatch value could weaponize.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install the operator into —
  literal or a KubernetesNamespace reference (`create_namespace` to
  own it)

### Common

- **`spec.chart_version`**: chart pin (default `1.8.0` — pairs with
  operator 1.0.0; bumps never touch the `crds/`-directory CRDs)
- **`spec.replicas`**: operator replicas (default 1); more than 1
  turns on leader election — the modules set the leader-election
  properties for you (the chart REFUSES multi-replica installs without
  them, by design)
- **`spec.workload.namespaces`**: the fence — watch scope and workload
  RBAC in one value; empty = cluster-wide
- **`spec.workload.service_account`**: service account Spark
  driver/executor pods run as in every workload namespace (default
  `spark` — the upstream convention SparkApplications reference)
- **`spec.operator_properties`**: `spark.kubernetes.operator.*` keys
  appended over the chart's defaults (the full key catalog ships with
  the operator's docs at the pinned version)
- **`spec.dynamic_config`**: hot property reloading from a ConfigMap —
  changes apply WITHOUT an operator restart; off by default (upstream
  default)
- **`spec.jvm_args`**: operator JVM arguments — empty = the chart's
  tuned default (parallel GC, 80% RAM percentage, crash on OOM)
- **`spec.resources`**: operator container resources — empty = the
  chart defaults (1 CPU / 2Gi, requests = limits); lower for lab
  clusters consciously — an OOM-killed operator strands every
  reconciling Spark job
- **`spec.image_registry`**: registry replacing the registry part of
  the operator image (`apache/spark-kubernetes-operator`, Docker Hub
  implied) — the air-gap path for the operator's own image; does NOT
  rewrite the images Spark workloads run (those ride each
  SparkApplication's own image field)
- **`spec.image_pull_secrets`**: names of existing image-pull Secrets
  in the namespace
- **`spec.scheduling`**: node selector, tolerations, priority class
  for the operator pods
- **`spec.helm_values`**: the escape hatch (see above)

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the operator runs in |
| `release_name` | Helm release name (= `metadata.name`; the chart's fullname and every RBAC name are pinned to it) |
| `workload_service_account` | Service account Spark driver/executor pods run as in every workload namespace — SparkApplication declarations reference it |
| `watched_namespaces` | Namespaces the operator watches for Spark workloads (empty = cluster-wide) |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **No prerequisites**: no webhook means no cert-manager dependency —
  the operator installs standalone.
- **Spark jobs depend on this component**: submit `SparkApplication`
  resources per pipeline run (from an orchestrator such as
  KubernetesAirflow, or via KubernetesManifest), referencing the
  exported `workload_service_account`. Under the fenced posture, jobs
  outside `watched_namespaces` are ignored without an error — a
  missing namespace in the list looks like a job that never starts.
- **The install is deliberately blocking**: the Helm release waits for
  the operator to become Available (atomic, 600s timeout), so an
  unpullable image or a broken config fails THIS apply with a
  readiness timeout instead of surfacing later as SparkApplications
  that mysteriously never reconcile.

## Examples

### Standard install

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSparkOperator
metadata:
  name: spark-operator
spec:
  namespace:
    value: spark-operator
  createNamespace: true
```

### Fenced team namespaces

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSparkOperator
metadata:
  name: spark-operator
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
  operatorProperties:
    spark.kubernetes.operator.reconciler.intervalSeconds: "15"
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
