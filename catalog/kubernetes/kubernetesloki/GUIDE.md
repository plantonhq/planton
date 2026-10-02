# KubernetesLoki Guide

The judgment this guide carries: Loki STORES logs — it ships none. An
architecture with Loki but no shipper collects nothing, silently; and the
mode choice is really a storage commitment.

## Nothing ships logs by itself — compose the shipper

Deploy a [KubernetesOtelCollector](../kubernetesotelcollector/GUIDE.md)
in daemonset mode with the cluster-logs pipeline (its presets carry it),
pointed at this component's `gateway_endpoint` output — that pair is the
cluster's log pipeline. Grafana reads the logs back through a `loki`
datasource on the same endpoint. The full wired composition:
[observability-stack pattern](../../_patterns/observability-stack.md).
Loki alone looks healthy and receives nothing — the gap appears only when
someone searches for logs that were never shipped.

OTLP-ingested lines are indexed by Loki's default resource labels, so a
cluster's pod logs are searched as `{k8s_namespace_name="<ns>"}` (also
`k8s_pod_name`, `k8s_container_name`, `service_name`), and a record's
`trace_id` is kept as structured metadata, not a label. When several
clusters ship to one Loki, stamp each line with OpenTelemetry's own
`k8s.cluster.name` and `deployment.environment.name` (a collector
`resource` processor with `action: insert`, so a pod that states its own
environment keeps it): both are in Loki's default index-label list, so
`{k8s_cluster_name="..."}` and `{deployment_environment_name="..."}` work
with no Loki setting. A custom attribute such as `cluster` would land as
structured metadata instead, filterable but not a stream label. The
canary is worth keeping wherever a person relies on the logs: it proves
the write and read paths end to end. Switching it off also switches off
the chart's Helm test, which reads the canary's metrics. The
`scheduling` block reaches Loki itself and its gateway, not the caches
or the canary.

One Loki taking several clusters' logs needs `limits` sized for the
moment they all catch up at once, after a Loki outage, not for the
steady rate. At the defaults (4 MB/s, a 6 MB burst) that catch-up is
refused in bursts and retried, which delays every line and fills the
senders' queues. Two rules: set `ingestion_rate_mb` above the summed
catch-up rate, and keep `ingestion_burst_size_mb` above every sender's
largest batch, because Loki refuses a push larger than its burst every
time, and a collector retrying without a deadline then stalls for good.
A hub for a few clusters at 12 and 24 took 8 MB/s for three minutes
through its door with nothing refused and Loki under 450 MiB. A single
hot stream does not hit the 3 MB/s per-stream limit: Loki shards hot
streams by default (`shard_streams`).

## The mode choice is a storage commitment

`monolithic` (default) runs everything in one StatefulSet — right for
dev and small production volumes. `simpleScalable` splits write/read/
backend tiers AND REQUIRES object storage — on-cluster, that means
composing a KubernetesSeaweedFs (the same S3 move the Flink and Tempo
guides make). Choosing the scalable mode without the storage is a
deployment that cannot come up; the microservices mode is deliberately
not modeled (the reference page says why).

## Logs outside the cluster: R2

When the logs must outlive the cluster that wrote them (a rebuilt
cluster, a moved region, an incident review a month later), store them in
Cloudflare R2 through the `r2` arm. It names the bucket, its account and
its jurisdiction by reference to a `CloudflareR2Bucket`, and the key pair
by reference (`$secret/` for a token minted in the dashboard, or a
`CloudflareAccountApiToken`'s outputs). The modules compose the S3 host,
region `auto` and path-style addressing, and write the pair into their own
`<name>-r2-credentials` Secret; a rotated key rolls the pods on the next
apply. Three things to get right:

- **Scope the token to the one bucket** with Object Read & Write. Nothing
  else needs it.
- **The bucket's expiry runs later than `retention_period`**, or objects vanish
  under the index. Declare the bucket's lifecycle rule with the retention,
  never shorter.
- **Pick the location hint where the cluster runs.** R2 honours it only
  at creation; a bucket can't move later. `jurisdiction` is a different,
  legal setting that changes the bucket's host; leave it unset unless
  data residency requires it.

## Alerts route through the one Alertmanager

`ruler.alertmanagerUrl` is a typed reference to the
kube-prometheus-stack's Alertmanager endpoint — wire it so log-driven
alerts join the same routing, silencing, and paging as everything else,
instead of growing a second alert system.

## Namespace ownership

Observability components conventionally share one namespace (the
pattern's example uses `observability`) — wire `spec.namespace` through
that KubernetesNamespace
([namespace-ownership pattern](../../_patterns/namespace-ownership.md)).

## Pairs well with

- KubernetesOtelCollector (daemonset) — the shipper; without it, nothing
  arrives.
- KubernetesGrafana — the reader (`loki` datasource on the gateway).
- KubernetesKubePrometheusStack — Alertmanager for log-driven alerts.
- KubernetesSeaweedFs — object storage when the mode demands it.
- CloudflareR2Bucket — object storage outside every cluster (the `r2` arm).
