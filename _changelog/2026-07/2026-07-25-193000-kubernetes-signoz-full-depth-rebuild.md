# Kubernetes observability: SigNoz — the all-in-one platform rebuilt at full depth

## What changed

- **KubernetesSignoz (872, rebuilt)** — deploys SigNoz, the
  OpenTelemetry-native observability platform (traces, metrics and logs
  in one application with one UI, stored in ClickHouse), from the
  official `signoz` Helm chart pinned 0.133.0 (the chart version tracks
  the application in lockstep). The previous surface — extension-based
  container blocks, a boolean external-database toggle and an embedded
  ingress block — is replaced end to end.

- **The database seam is a oneof.** Empty or `managed_clickhouse` = the
  bundled appliance: the chart's own ClickHouse stack (a
  namespace-fenced Altinity operator, the installation, ZooKeeper) with
  capacity/topology knobs (shards, replicas per shard, volume size and
  class, resources, the network allow-list, an odd-count ZooKeeper
  quorum) plus S3/GCS cold-storage tiering with a keyless IRSA arm.
  `external_clickhouse` = composition: nothing database-related
  installs, and every connection field default-references a
  KubernetesClickHouse resource's outputs (client Service, logical
  cluster name, auth Secret) — the user there needs `access_management`
  and cluster-wide DDL grants, since SigNoz owns its schema migrations.
  Deep ClickHouse control is deliberately not re-modeled inside SigNoz.

- **The bundled database credential is module-generated.** The chart's
  publicly-documented default password never ships: both engines
  generate a random credential per install, deliver it outside the
  values documents (Terraform `set_sensitive`; a Pulumi secret Output
  injected after the values merge), and export it through the
  module-owned `<name>-clickhouse-auth` Secret — the composition handle
  the outputs promise. Imported installs never silently rotate it (the
  generation-shape arguments are ignored after creation).

- **Single-instance truth taught, not papered over:** the community
  SigNoz server keeps users/dashboards/alert rules in SQLite on a
  PersistentVolume (the Postgres-backed HA store is enterprise-only
  upstream), so the server deliberately has no replica knob; the
  ingestion collector scales instead (replicas or a typed HPA arm).

- **Ingestion receivers move as one contract:** OTLP gRPC/HTTP always
  on; Jaeger (default on) and Zipkin (default off) toggles drive the
  Service ports AND the collector pipeline receiver lists from one
  derivation, so a receiver is never exposed without being wired or
  wired without being exposed. Plain-HTTP log endpoints are
  toggleable; exception-grouping cardinality is a typed trade-off.

- **Alerting wired secret-safe:** SMTP for alert emails and invitations
  (address, From, auth with the password as a secretKeyRef env entry —
  never rendered material) and the external UI URL that alert links
  point at; an advanced env map covers the rest of the server's
  documented configuration surface, with typed fields winning.

- The `postgresql`, `signoz-otel-gateway` (enterprise/licensed) and
  `redpanda` (excluded license family) subcharts are never modeled and
  never enabled; the embedded ingress block is gone — exposure composes
  from first-class kinds over the exported handles (the server Service,
  UI endpoint, OTLP gRPC/HTTP endpoints, ClickHouse endpoint and
  credential Secret).

## Engines

Both engines render byte-identical chart values from the same typed
spec, with the `helm_values` escape hatch merged last (Helm `-f`
semantics) and BOTH fullnames — the release and the bundled
`<name>-clickhouse` — re-pinned after it so every exported name stays
deterministic. Every chart image is the split registry+repository form
deferring to `global.imageRegistry`, so one `image_registry` value
re-points the server, collector, ClickHouse, operator, metrics-exporter
and ZooKeeper images at a private mirror. Resource names are capped at
30 characters, failing loudly on both engines — the bundled ClickHouse
wraps the name in ~27 characters of operator scaffolding inside
Kubernetes' 63-character cap.

## Validation

Spec tests (32 cases); both engines' `plan`/`preview` proofs across
full, minimal and external-clickhouse shapes (real chart pulls, rendered
values spot-checked); a dedicated E2E verifier authored and compiled —
the product-grade proof registers the first admin through SigNoz's own
API, opens a session, pushes a span over OTLP and retrieves the trace by
ID through the authenticated query API, with a state proof that
re-authenticates and re-queries after a server pod replacement; an
import map covering the release, the anchor namespace, the
module-materialized credentials Secret and its random_password
companion (imported by value); presets and docs. Secret-coverage,
outputs conformance, importmap conformance, catalogkindreflect, the containment
golden, the structural guards and `make build-go` all pass.
