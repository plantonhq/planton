# Kubernetes observability: Grafana Loki + Grafana Tempo — the logs and traces tier at full depth

## What changed

- **KubernetesLoki (873, new)** — deploys Grafana Loki, the
  label-indexed log-aggregation backend, from the official `loki` Helm
  chart pinned 18.5.4 (Loki 3.7.4). Typed surface: two deployment
  modes (a single-StatefulSet monolithic default and independently-
  scaling write/read/backend simple-scalable tiers); object-storage
  backends (S3 including S3-compatible endpoints, GCS, Azure) with
  keyless ambient-identity or referenced-Secret credentials that ride
  environment-variable expansion so they never enter the rendered
  config; a derived TSDB v13 index schema so no `schema_config` need be
  hand-authored; compactor-driven retention; ingestion/query limits;
  the nginx gateway front door; the memcached chunks/results caches;
  the write-read canary; the LogQL ruler firing at an Alertmanager;
  and multi-tenant gateway basic auth via bcrypt password hashes.
  Filesystem storage is honest only for a single monolithic replica
  (mirrored from the chart's own validation in CEL). The bundled MinIO
  dev subchart is never enabled. Exports the gateway endpoint, the
  OTLP push endpoint and the Loki HTTP Service for composition.

- **KubernetesTempo (874, new)** — deploys Grafana Tempo, the
  object-storage-backed distributed-tracing backend, from the official
  monolithic `tempo` Helm chart pinned 2.2.3 (Tempo 2.10.7). Typed
  surface: replicas; `local`/S3/GCS/Azure trace storage with the same
  secret-safe credential model; persistent-by-default trace storage
  (the chart's own emptyDir default would lose traces on restart, with
  an explicit `ephemeral` arm); Go-duration retention (`h`/`m` only —
  Tempo rejects a day unit, enforced in CEL); OTLP receivers always on
  with the four legacy Jaeger protocols opt-in; multi-tenancy; the
  metrics generator deriving service-graph and span metrics from the
  trace stream and remote-writing them to a Prometheus (with the
  standard `/api/v1/write` path appended when the URL carries none);
  and the tempo-query Jaeger-UI sidecar. Exports the query HTTP
  endpoint and both OTLP ingest endpoints.

- Both kinds compose with the rest of the observability tier by
  reference: a KubernetesOtelCollector ships logs/traces in, a
  KubernetesGrafana datasource reads them back, and Tempo's metrics
  generator and Loki's ruler wire into KubernetesKubePrometheusStack.

## Engines

Both kinds ship Pulumi and Terraform modules that render byte-identical
chart values from the same typed spec, with a `helm_values` escape hatch
merged last (Helm `-f` semantics) and `fullnameOverride` re-pinned after
it so every exported name stays deterministic. The charts mix split
`registry`+`repository` image values with combined docker-library images
(Loki's memcached, Tempo's tempo-query); both engines translate an
air-gap registry override correctly for both forms.

## Validation

Spec tests; both engines' `plan`/`preview` proofs across minimal and
full-surface shapes (real chart pulls); dedicated E2E verifiers (Loki
push→LogQL round-trip + volume-loss durability; Tempo OTLP→trace-by-ID
round-trip + persistence-through-pod-loss) authored and compiled; import
maps; presets; docs. Secret-coverage, outputs conformance, importmap
conformance, catalogkindreflect, the structural guards and `make build-go` all
green.
