# Cluster-logs-to-Loki preset

The per-node log pipeline: daemonset mode puts one collector on every
node, the `file_log` receiver tails every container's log files under
`/var/log/pods` (the standard `container` operator parses the runtime
format and extracts the Kubernetes metadata from the file path), the
`k8s_attributes` processor enriches records with pod/namespace/workload
attributes, and the `otlp_http` exporter ships everything to a Loki
gateway's `/otlp` route. The standard `otlp` receiver rides along so
node-local applications can push their own telemetry too — and so the
exported OTLP endpoints stay valid. The component names are the
collector's current ones (`file_log`, `k8s_attributes`, `otlp_http`); the
older `filelog`, `k8sattributes` and `otlphttp` still load as deprecated
aliases. Renaming a receiver changes the key its checkpoint is stored
under, so the first restart after a rename reads each file from its end.

PREREQUISITE: a `KubernetesOtelOperator` on the cluster. The
`k8s_attributes` processor reads cluster state, which needs RBAC beyond
the operator's default ServiceAccount — compose a
`KubernetesServiceAccount` + `KubernetesRbac` for the
`otel-logs-collector` account this preset names in `serviceAccount`.

The volumes are the daemonset log-collection pattern: `/var/log/pods`
mounted read-only from the host (the receiver only reads), plus a
writable hostPath at `/var/lib/otelcol-checkpoints`. The mount alone
keeps nothing: the `file_storage` extension points at it, the receiver's
`storage` keeps its offsets there (a restarted collector resumes where it
left off), and the exporter's `sending_queue` keeps unsent lines there
with retries that never give up. The control-plane toleration covers
every node — remove it if control-plane logs should stay uncollected.

The sending queue is what makes the pipeline lose nothing, and three of
its settings are load-bearing:
- **Batching lives in the queue, not in a `batch` processor.** A batch
  processor holds lines in memory after the reader has moved its
  checkpoint past them, so a restart loses them. Batching from the
  queue happens after the disk, and an item leaves the disk only once
  its batch is delivered (at least once: a batch in flight at a restart
  is sent again).
- **The batch is capped in bytes (`sizer: bytes`, `max_size`) below
  Loki's ingestion burst** (`ingestion_burst_size_mb`, 6 MB by
  default). Loki refuses a push larger than its burst every time, and
  with retries that never give up, one oversized batch would stop that
  node's logs for good. A line count is no cap: a few stack traces
  push 8,192 lines past 6 MB.
- **`block_on_overflow: true`.** A full queue otherwise drops new lines;
  with it the reader waits, and the lines stay in the pods' own log
  files until there is room. `queue_size` (in bytes here) bounds the
  node disk the queue may use; past it, the kubelet's log rotation
  bounds how long a line can wait.

Two receiver lines that look optional are not. `include_file_path: true`
is what the `container` operator reads pod, namespace and container
from; without it the receiver drops every line with "log.file.path is
missing". The `exclude` keeps the collector off its own log files (pods
`<name>-collector-<hash>`; change the namespace and name with yours),
which would otherwise echo each export error back into the stream it is
failing to send.

`serviceMonitorEnabled: true` has the operator create a ServiceMonitor
for the collector's own metrics (port 8888): lines accepted and sent,
failed sends, and the queue's fill, which rises first when Loki refuses
or goes away. The operator looks for the Prometheus operator's CRDs once,
when it starts, so install the monitoring stack before it.

The `podSecurityContext.runAsUser: 0` is load-bearing, not a
convenience: container runtimes write pod log files readable only by
root, and the default collector image runs as a non-root user that
cannot open them — without it the receiver reports permission errors and
ships nothing.

Sizing discipline: the `memory_limiter` (400 MiB limit, 100 MiB spike)
sheds load visibly instead of OOMing — if you add container resources,
keep the memory limit and `limit_mib` in agreement.

Change first: the `otlp_http` endpoint's host — point it at your
`KubernetesLoki`'s `otlp_push_endpoint` output (Loki ingests OTLP at the
gateway's `/otlp` route) — the `exclude` path's namespace and name, and
`max_size` if your Loki's burst is not the default.

See [01-cluster-logs-to-loki.yaml](./01-cluster-logs-to-loki.yaml) for
the manifest.
