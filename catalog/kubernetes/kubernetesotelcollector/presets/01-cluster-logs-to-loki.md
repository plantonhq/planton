# Cluster-logs-to-Loki preset

The per-node log pipeline: daemonset mode puts one collector on every
node, the filelog receiver tails every container's log files under
`/var/log/pods` (the standard `container` operator parses the runtime
format and extracts the Kubernetes metadata from the file path), the
`k8sattributes` processor enriches records with pod/namespace/workload
attributes, and the `otlphttp` exporter ships everything to a Loki
gateway's `/otlp` route. The standard `otlp` receiver rides along so
node-local applications can push their own telemetry too — and so the
exported OTLP endpoints stay valid.

PREREQUISITE: a `KubernetesOtelOperator` on the cluster. The
`k8sattributes` processor reads cluster state, which needs RBAC beyond
the operator's default ServiceAccount — compose a
`KubernetesServiceAccount` + `KubernetesRbac` for the
`otel-logs-collector` account this preset names in `serviceAccount`.

The volumes are the daemonset log-collection pattern: `/var/log/pods`
mounted read-only from the host (the receiver only reads), plus a
writable hostPath at `/var/lib/otelcol-checkpoints`. The mount alone
keeps nothing: the `file_storage` extension points at it, the filelog
receiver's `storage` keeps its offsets there (a restarted collector
resumes where it left off), and the exporter's `sending_queue` keeps
unsent lines there with retries that never give up, so logs written
while Loki is down arrive when it returns. The control-plane toleration
covers every node — remove it if control-plane logs should stay
uncollected.

Two lines that look optional are not. `include_file_path: true` is what
the `container` operator reads pod, namespace and container from;
without it the receiver drops every line with "log.file.path is
missing". The `exclude` keeps the collector off its own log files
(pods `<name>-collector-<hash>`; change the namespace and name with
yours), which would otherwise echo each export error back into the
stream it is failing to send.

The `podSecurityContext.runAsUser: 0` is load-bearing, not a
convenience: container runtimes write pod log files readable only by
root, and the default collector image runs as a non-root user that
cannot open them — without it the filelog receiver reports permission
errors and ships nothing.

Sizing discipline: the `memory_limiter` (400 MiB limit, 100 MiB spike)
sheds load visibly instead of OOMing — if you add container resources,
keep the memory limit and `limit_mib` in agreement.

Change first: the `otlphttp` endpoint's host — point it at your
`KubernetesLoki`'s `otlp_push_endpoint` output (Loki ingests OTLP at the
gateway's `/otlp` route) — and the `exclude` path's namespace and name.

See [01-cluster-logs-to-loki.yaml](./01-cluster-logs-to-loki.yaml) for
the manifest.
