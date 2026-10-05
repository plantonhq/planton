# Kubernetes SigNoz

## When NOT to Use This

**One resource is ONE SigNoz install** — the all-in-one open-source
observability platform: traces, metrics and logs in one UI, stored in
ClickHouse, ingested over OpenTelemetry.

Not the right kind when:

- **You want the composed best-of-breed stack** — that is
  `KubernetesKubePrometheusStack` + `KubernetesGrafana` +
  `KubernetesLoki` + `KubernetesTempo`. SigNoz is the "one product, one
  UI" alternative; both paths are first-class.
- **You need something to SHIP cluster telemetry** — SigNoz receives
  OTLP; it does not scrape nodes or tail container logs by itself.
  Deploy a `KubernetesOtelCollector` pointed at this component's
  exported `otlp_grpc_endpoint` / `otlp_http_endpoint`.
- **You expect a public endpoint out of the box** — everything is
  ClusterIP; exposure composes from first-class kinds
  (`KubernetesIngress`, the Gateway API kinds) over the exported
  service handles.

## The database is composed, never bundled

SigNoz stores every trace, metric and log in ClickHouse — a required,
separate component. Run a `KubernetesClickHouse` (with its
`KubernetesAltinityOperator`) and wire it through `clickhouse`; the
fields default-reference that kind's outputs (client Service, cluster
name, auth Secret), so the composition is one `valueFrom` per field.
Any reachable ClickHouse also works with literal values.

Why composed: the database — and your telemetry — outlives SigNoz
reinstalls; upgrades roll independently; deep ClickHouse control
(users, profiles, quotas, keeper topology) lives on the component that
owns it. And verified live: a chart-bundled database cannot uninstall
cleanly — the operator and its installation die in the same Helm
release, and the installation's deletion finalizer deadlocks with
nobody left to process it. Composition removes that failure by
construction.

What the composed ClickHouse needs:

- **SigNoz's tested version** — run the ClickHouse at the version
  SigNoz's chart pairs with (25.12.5 at chart 0.133.0). Verified
  live: the schema migrations use settings newer ClickHouse ships,
  and an older server (e.g. 25.3) fails them with
  `Unknown setting 'object_serialization_version'`.
- **Coordination** — SigNoz runs its schema migrations `ON CLUSTER`,
  and distributed DDL requires a keeper: on a single-replica
  `KubernetesClickHouse`, declare `coordination.type: managed_keeper`
  explicitly (bare 1x1 topologies default to none).
- **A user** — the simplest honest posture is a user declared with NO
  grants (verified live: unrestricted config-user access, which covers
  the migrator). A constrained user needs `GRANT CLUSTER ON *.*` plus
  CREATE/DROP/INSERT/SELECT on the `signoz_*` databases.
- **Explicit `networks` on that user** — verified live: a user
  declared without networks is fenced by the operator to the
  ClickHouse pods and localhost, and SigNoz's pods are rejected with
  what reads as a password failure ("Authentication failed"). Declare
  `networks` (e.g. `0.0.0.0/0` + `::/0` when the password is the
  gate).
- **Co-location or a replicated Secret** — SigNoz reads the password
  through a `secretKeyRef`, which cannot cross namespaces: deploy
  SigNoz into the ClickHouse namespace (the default composition) or
  replicate the Secret.

## Single-instance truth

The community SigNoz server keeps users, dashboards and alert rules in
SQLite on a PersistentVolume — a single-writer store, so the server runs
exactly one replica (the Postgres-backed HA store is enterprise-only
upstream). Telemetry lives in ClickHouse and scales independently; the
ingestion collector scales horizontally (replicas or the HPA arm).

## Retention

Telemetry retention is managed inside SigNoz (UI → Settings → General),
not in this spec — size the composed ClickHouse's `disk_size` for the
retention you set there.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
