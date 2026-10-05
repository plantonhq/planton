# Default

This preset installs the Strimzi cluster operator in its standard
posture: the pinned `strimzi-kafka-operator` chart, own-namespace
watch scope, and the chart's own defaults for everything else
(reconciliation timing, resources, leader election). Kafka clusters
are declared afterwards as KubernetesKafka resources in the SAME
namespace — the operator watches its own namespace by default, and
that is a deliberate posture (Kafka clusters live beside their
operator), not a limitation to work around.

## When to Use

- Any cluster that will run KubernetesKafka clusters
- The 30-second choice: this is the standard first installation; widen
  the watch only when one operator must manage Kafka clusters across
  namespaces

## Key Configuration Choices

- **`namespace: kafka` + `create_namespace: true`** — the operator's
  namespace is where the Kafka clusters it reconciles live (the
  default watch scope); this resource creates and owns it
- **Own-namespace watch (no `watch` block)** — the chart default.
  `watch.any_namespace: true` makes one operator manage every
  namespace; `watch.namespaces` fences an explicit set. The two are
  mutually exclusive
- **`chart_version: "1.1.0"`** (the spec default, stated explicitly)
  — chart and operator versions move together for this chart, and the
  version must exist as a SERVED chart in the repository index
  (https://strimzi.io/charts/); upgrades re-run the release with the
  new chart, deliberately
- **Chart defaults everywhere else** — one operator replica,
  reconciliation every 2 minutes, operation timeout 5 minutes,
  requests 200m/384Mi and limits 1000m/384Mi. The modules render
  values only on divergence, so this preset installs the chart exactly
  as upstream ships it
- **CRDs outlive the release** — the chart ships the Strimzi CRDs in
  its Helm-native `crds/` directory: installed on first install, never
  upgraded or deleted by Helm. Uninstalling the operator therefore
  never cascade-deletes the Kafka clusters, and a `chart_version`
  upgrade runs new operator code against the EXISTING CRDs — apply the
  new release's CRDs yourself when its release notes call for it

## Placeholders to Replace

None — this preset deploys as-is.

## Related Kinds

- **KubernetesKafka** — the Kafka clusters this operator reconciles,
  one resource per cluster, declared in the watched namespace
- **KubernetesKafkaTopic / KubernetesKafkaUser** — topics and users,
  reconciled by each cluster's entity operators
