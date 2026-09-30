# Kubernetes Tempo

## When NOT to Use This

**One resource is ONE Grafana Tempo install** — a distributed-tracing
backend that stores whole traces in object storage and retrieves them by
ID or TraceQL.

Not the right component when:

- **You need something to SEND the traces** — Tempo stores traces, it does
  not instrument or collect them. Applications (or a
  `KubernetesOtelCollector` in between) send OTLP to the exported
  `otlp_grpc_endpoint` / `otlp_http_endpoint`.
- **You want to READ traces in a UI** — that is Grafana. Point a
  `KubernetesGrafana` datasource of type `tempo` at the exported
  `http_endpoint`.
- **You expect a public endpoint out of the box** — everything is
  ClusterIP; exposure composes from first-class kinds over the exported
  handles.

## Grain

This kind models the **monolithic** Tempo chart — one StatefulSet,
production-capable with an object-storage backend. The separate
`tempo-distributed` microservices chart is deliberately not modeled.

## Storage

`local` (a PersistentVolume, single replica) or an object store — `s3`
(including S3-compatible endpoints like an in-cluster `KubernetesSeaweedFs`),
`r2` (a Cloudflare R2 bucket by reference, with the key pair in the module's
own `<name>-r2-credentials` Secret), `gcs`, or `azure`. More than one replica requires object storage.
Credentials are references to existing Secrets that ride env expansion —
never the rendered config; empty credentials use ambient cloud identity.
The chart's own default is an emptyDir (traces vanish on restart), so this
component provisions a PersistentVolume by default; `ephemeral` restores
the chart posture for throwaway clusters.

## Ingest

OTLP is always on (gRPC 4317, HTTP 4318 — the 2026 wire standard). The
four legacy Jaeger protocols are opt-in (`jaeger_receivers_enabled`) — every
closed receiver is one less ingest surface. Retention is a Go duration
(`h`/`m` only — Tempo does not accept a day unit).

## Metrics from traces

The metrics generator derives service-graph and span metrics from the
trace stream and remote-writes them to a Prometheus — the seam that lights
up Grafana's service map. Point `remote_write_url` at a
`KubernetesKubePrometheusStack` (whose Prometheus must enable its
remote-write receiver).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
