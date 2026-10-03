# Standard

This preset installs the Percona Operator for MySQL (XtraDB Cluster)
in its standard posture: the pinned `pxc-operator` chart, own-namespace
watch scope, telemetry off, structured logs, and explicit control-plane
resources. Databases are declared afterwards as KubernetesMysql
resources in the SAME namespace — the operator watches its own
namespace by default, and that is a deliberate posture (databases live
beside their operator), not a limitation to work around.

## When to Use

- Any cluster that will run KubernetesMysql databases
- The 30-second choice: this is the standard first installation; widen
  the watch only when one operator must manage databases across
  namespaces

## Key Configuration Choices

- **`namespace: percona-mysql` + `create_namespace: true`** — the
  operator's namespace is where the databases it reconciles live (the
  default watch scope); this resource creates and owns it
- **Own-namespace watch (no `watch` block)** — the upstream default.
  `watch.cluster_wide: true` makes one operator manage every
  namespace; `watch.namespaces` fences an explicit set. The two are
  mutually exclusive
- **`chart_version: "1.20.0"`** (the spec default, stated explicitly)
  — chart and operator versions move together for this chart; upgrades
  re-run the release with the new chart, deliberately
- **`disable_telemetry: true`** — no anonymous version/feature pings
  to check.percona.com
- **`log.structured: true`** — JSON logs instead of the console
  encoder, for whatever pipeline collects control-plane logs
- **Explicit resources** — the chart's defaults (requests 100m/20Mi,
  limits 200m/500Mi) are tight for a control plane reconciling real
  databases; a modest bump costs nothing and removes a failure mode
- **CRDs outlive the release** — the chart ships the
  PerconaXtraDBCluster CRDs in its Helm-native `crds/` directory:
  installed on first install, never upgraded or deleted by Helm.
  Uninstalling the operator therefore never cascade-deletes the
  database clusters; CRD upgrades ride chart upgrades through the
  module's server-side apply

## Placeholders to Replace

None — this preset deploys as-is.

## Related Kinds

- **KubernetesMysql** — the databases this operator reconciles, one
  resource per XtraDB Cluster, declared in the watched namespace
