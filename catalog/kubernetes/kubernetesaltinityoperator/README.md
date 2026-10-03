# Kubernetes Altinity Operator

## When NOT to Use This

**This component installs the ENGINE, not a database cluster.** The
Altinity ClickHouse operator reconciles `ClickHouseInstallation` and
`ClickHouseKeeperInstallation` custom resources into running clusters;
those clusters are declared with KubernetesClickHouse — one resource
per cluster. Install the operator once per Kubernetes cluster (or once
per watched namespace set), then declare ClickHouse clusters against
it.

Also not the right kind when:

- **You want a ClickHouse cluster** — that is KubernetesClickHouse;
  this kind is the controller that reconciles it, including the
  managed-Keeper arm (`ClickHouseKeeperInstallation` resources are
  reconciled by this same operator).
- **You want a managed cloud analytics service** — use ClickHouse
  Cloud or the host cloud provider's managed offerings; this component
  is for running ClickHouse ON the Kubernetes cluster itself.

## Overview

**KubernetesAltinityOperator** installs the Altinity ClickHouse
operator — the Apache-2.0 operator for running ClickHouse (the
columnar OLAP database) on Kubernetes — from the official
`altinity-clickhouse-operator` Helm chart
(https://docs.altinity.com/clickhouse-operator/). The operator
reconciles `ClickHouseInstallation` (CHI) and
`ClickHouseKeeperInstallation` (CHK) custom resources into running
clusters with generated server configuration, rolling restarts, and
per-host StatefulSets.

**Key design points:**

- **The chart owns the CRD lifecycle — safely.** The chart ships its
  four CRDs in its `crds/` directory: Helm installs them on first
  install and NEVER deletes them on uninstall, so removing the
  operator never cascade-deletes ClickHouse clusters or their data.
  CRD UPGRADES ride the chart's pre-install/pre-upgrade hook job
  (`crd_hook`, enabled by default), which server-side-applies the CRDs
  on every install and upgrade — disabling it means chart upgrades
  silently run new operators against old schemas.
- **Chart and operator versions move in lockstep.** Chart 0.27.2 runs
  operator image 0.27.2 (the pinned default); pick `chart_version`
  from the served repository index.
- **Watch scope defaults to the operator's OWN namespace.** Unlike
  many operators, the chart default watches only the install
  namespace. `watch_namespaces` entries are regular expressions —
  every namespace that will hold KubernetesClickHouse resources must
  be covered, or use `[".*"]` for cluster-wide; a fenced operator
  silently ignores clusters elsewhere. `namespace_scoped_rbac` pairs a
  single-namespace watch with Role/RoleBinding-only RBAC — the tenancy
  posture for shared clusters.
- **The operator logs into every cluster it manages.**
  `operator_credentials` is the ClickHouse user the operator uses for
  host management, schema propagation, and metrics scraping —
  provisioned as a chart-managed Secret and auto-injected (restricted
  to the operator pod's address) into every managed cluster. The chart
  defaults are publicly documented — unset is unsafe outside throwaway
  environments.
- **Metrics for every managed cluster.** The metrics-exporter sidecar
  serves Prometheus metrics for all managed clusters on port 8888;
  `service_monitor_enabled` adds a ServiceMonitor (requires the
  Prometheus Operator CRDs — enabling it without them fails the
  install).
- **Keep the resource name at 39 characters or fewer.** The modules
  pin the chart fullname to the resource name, and the longest
  generated child name (`<fullname>-keeper-templatesd-files`) adds 24
  characters to it against the Kubernetes 63-character cap.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines). One key is off limits:
  `fullnameOverride` is re-pinned by the modules AFTER the merge —
  overriding it would break the naming budget and every
  fullname-derived output.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install the operator into —
  literal or a KubernetesNamespace reference (`create_namespace` to
  own it)

### Common

- **`spec.chart_version`**: chart pin (default `0.27.2`; the chart
  version and the operator image tag move in lockstep)
- **`spec.watch_namespaces`**: regexp list of namespaces the operator
  watches; empty = the operator's own namespace only (the chart
  default), `[".*"]` = the whole cluster
- **`spec.namespace_scoped_rbac`**: Roles/RoleBindings instead of
  cluster-wide RBAC — only sound when the operator watches its own
  namespace alone
- **`spec.operator_credentials`**: the operator's own ClickHouse
  login (username default `clickhouse_operator`; password literal or
  reference) — always set a real password outside throwaway
  environments
- **`spec.metrics`**: the metrics-exporter sidecar (enabled by
  default; per-cluster Prometheus metrics on port 8888)
- **`spec.crd_hook`**: the CRD install/upgrade hook job (enabled by
  default; leave it on). Its default image is
  `bitnami/kubectl:latest` — pullable today but frozen since Bitnami's
  2025 public-catalog retirement, so it will silently age; override
  `crd_hook.image` for air-gapped mirrors or a maintained kubectl
  build
- **`spec.service_monitor_enabled`**: ServiceMonitor for Prometheus
  Operator scraping (chart default false; requires the Prometheus
  Operator CRDs)
- **`spec.resources`**: operator container resources — empty = the
  chart defaults (no requests/limits)
- **`spec.node_selector` / `spec.tolerations`**: operator pod
  scheduling
- **`spec.image` / `spec.image_pull_secrets`**: air-gap and
  private-mirror path (empty = `altinity/clickhouse-operator` at the
  chart's version)
- **`spec.helm_values`**: the escape hatch (see above for the one
  off-limits key)

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the operator is installed into |
| `release_name` | Helm release name of the operator install (= `metadata.name`) |
| `deployment_name` | The operator Deployment — the chart fullname, which the modules pin to the resource name |
| `credentials_secret_name` | Chart-managed Secret holding the operator's ClickHouse credentials (keys `username`/`password`) |
| `metrics_endpoint` | In-cluster Prometheus metrics URL for every managed cluster (`http://<name>-metrics.<namespace>.svc.cluster.local:8888/metrics`); empty when metrics are disabled |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **`spec.operator_credentials.password`** accepts a reference to
  another resource's output, so the operator password can flow from a
  secrets store instead of living in the manifest.
- **KubernetesClickHouse resources depend on this component**: the
  operator must be running and watching their namespace before their
  CHI/CHK resources reconcile. With the chart-default watch scope,
  only clusters in the operator's OWN namespace reconcile — cover
  every ClickHouse namespace in `watch_namespaces`, or use `[".*"]`
  so one install serves the whole cluster.
- **The install is deliberately blocking**: the Helm release waits for
  the operator Deployment to become Available (atomic, 600s timeout),
  so an unpullable image fails THIS apply instead of surfacing later
  as clusters that mysteriously never reconcile.

## Examples

### Standard cluster-wide install

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAltinityOperator
metadata:
  name: clickhouse-operator
spec:
  namespace:
    value: clickhouse-operator
  create_namespace: true
  watch_namespaces:
    - ".*"
  operator_credentials:
    username: clickhouse_operator
    password:
      value: change-me-operator-password
```

### Namespace-scoped install on a shared cluster

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAltinityOperator
metadata:
  name: team-clickhouse-operator
spec:
  namespace:
    value: team-data
  create_namespace: true
  watch_namespaces:
    - team-data
  namespace_scoped_rbac: true
  operator_credentials:
    username: clickhouse_operator
    password:
      value: change-me-operator-password
```

### Private-mirror images

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAltinityOperator
metadata:
  name: clickhouse-operator
spec:
  namespace:
    value: clickhouse-operator
  create_namespace: true
  watch_namespaces:
    - ".*"
  operator_credentials:
    username: clickhouse_operator
    password:
      value: change-me-operator-password
  image:
    repo: mirror.example.com/altinity/clickhouse-operator
    tag: 0.27.2
  crd_hook:
    image:
      repo: mirror.example.com/kubectl
      tag: "1.31"
  image_pull_secrets:
    - mirror-pull-secret
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
