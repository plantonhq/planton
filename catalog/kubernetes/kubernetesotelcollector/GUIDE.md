# KubernetesOtelCollector Guide

The judgment this guide carries: the collector's value is its pipeline
config and the MODE you run it in — and two composition needs (credentials
and cluster-read RBAC) are easy to get wrong in ways that leak secrets or
silently collect nothing.

## Pick the mode by what you are collecting

`deployment` (default) is the scalable gateway/fan-in collector;
`daemonset` is per-node collection (log files, host and kubelet metrics —
this is how cluster logs reach a KubernetesLoki, paired with hostPath
`volumes`); `statefulset` is for stable identities (target allocator,
persistent queues); `sidecar` injects into annotated pods and creates no
standalone workload (the field doc on [reference.md](v1alpha1/reference.md)). The
mode is an architecture decision — a gateway where you needed a daemonset
collects none of the per-node telemetry you wanted.

## Never inline credentials; grant RBAC for cluster receivers

Two traps the spec's own docs call out: load exporter credentials as env
from existing Secrets and reference them as `${env:VAR}` in the config, so
tokens never land in the rendered ConfigMap. And receivers that read
cluster state (k8s_events, kubeletstats, k8s_cluster, enriched filelog)
need RBAC beyond the default ServiceAccount — compose a
KubernetesServiceAccount + KubernetesRbac and set `serviceAccount`, or the
receiver silently collects nothing.

## A log collector that loses nothing and reads only others

These settings decide whether a daemonset log pipeline works, and none
is the default (the `01-cluster-logs-to-loki` preset carries all of
them):
- `include_file_path: true` on the `file_log` receiver — the `container`
  operator takes pod, namespace and container from the file path;
  without it every line is dropped with "log.file.path is missing".
- `file_storage` on a hostPath, used by the receiver's `storage` (offsets
  survive a restart) and the exporter's `sending_queue` with
  `retry_on_failure.max_elapsed_time: 0` (lines wait on the node's disk
  while the destination is down, then arrive). A mounted checkpoint
  volume that nothing points at keeps nothing.
- batching in that queue (`sending_queue.batch`), never in a `batch`
  processor: the processor holds lines in memory after the reader has
  moved its checkpoint past them, so a restart loses them.
- the batch capped in bytes (`sizer: bytes`, `max_size`) below the
  backend's largest accepted push. Loki refuses a push larger than its
  `ingestion_burst_size_mb` every time; retried without a deadline, one
  such batch stops the node's logs for good.
- `block_on_overflow: true`, or a full queue drops new lines instead of
  leaving them in the pod log files until there is room.
- an `exclude` for the collector's own pods (`<name>-collector-<hash>`),
  or every export error echoes back into the stream it failed to send.

Proven on a live hub: with these settings, a log writer's 1,440
numbered lines all arrived across a ten-minute Loki outage with the
node's collector deleted halfway, none missing.

## Watch the collector itself

`service_monitor_enabled: true` has the operator create a ServiceMonitor
for the collector's own metrics on port 8888 (a PodMonitor in sidecar
mode): `otelcol_receiver_accepted_log_records`,
`otelcol_exporter_sent_log_records`,
`otelcol_exporter_send_failed_log_records` and
`otelcol_exporter_queue_size` against `otelcol_exporter_queue_capacity`.
The queue's fill is the earliest sign that a backend is refusing or
away, before any line is late enough to notice. The operator detects
the Prometheus operator's CRDs once, when it starts, so a monitoring
stack installed after it needs the operator restarted before monitors
appear; declare the KubernetesOtelOperator `depends_on` the
KubernetesKubePrometheusStack.

## Operator prerequisite

KubernetesOtelOperator is the registry prerequisite, watching cluster-wide
([operator-prerequisite pattern](../../_patterns/operator-prerequisite.md);
its [guide](../kubernetesoteloperator/GUIDE.md) has the watch
judgment). Wire `spec.namespace` to a dedicated KubernetesNamespace, not
`createNamespace: true`
([namespace-ownership pattern](../../_patterns/namespace-ownership.md)).

## Pairs well with

- KubernetesOtelOperator — required.
- KubernetesLoki / KubernetesTempo — the log and trace backends collector
  pipelines export to.
- KubernetesServiceAccount + KubernetesRbac — the permissions cluster-read
  receivers need.
