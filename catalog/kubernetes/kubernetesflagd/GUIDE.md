# KubernetesFlagd Guide

flagd is easy to deploy and easy to wedge: its pods turn ready only after every source has synced, its ConfigMap sources must share its namespace, and a flag flip arrives on the kubelet's schedule, not instantly. This guide is the judgment for choosing it, wiring its sources, and operating it on Planton.

Substitutes for: flagd sidecar injection by the OpenFeature Operator (this kind runs flagd as one central Service instead of a sidecar per pod; the same flag definition format and providers apply).

Pinned: flagd v0.17.0 (`ghcr.io/open-feature/flagd:v0.17.0`). Coverage: every `flagd start` setting and source option at that version is a typed field except these, left out by design -- the unix socket listeners (`--socket-path`, `--sync-socket-path`), which serve only a sidecar in the same pod, never a Service; `--keep-alive-min-time` and `--keep-alive-permit-without-stream`, no-ops kept for compatibility; the HTTP source's OAuth `folder` and `reloadDelayS`, which re-read client credentials from mounted files where the module passes them directly; the `file`, `fsnotify` and `fileinfo` providers, reached through `configMap` sources (whose `watcher` picks `fsnotify` or `fileinfo`); and exposure, which composes as a Gateway API route against the `service` output.

## When to use it (and when not)

Choose **KubernetesFlagd** when JSONLogic targeting fits your rules, when you want the CNCF-governed reference daemon, or when services should evaluate in-process from the sync stream (`sync_endpoint`) instead of paying a network hop per evaluation. Choose **KubernetesGoFeatureFlag** instead when you need change notifications (Slack, Teams, webhooks), evaluation export (to buckets, BigQuery, Kafka and more), API-key-isolated flag sets, readable YAML flag files, or a flip that lands within one polling interval -- GO Feature Flag reads its ConfigMap through the Kubernetes API, while flagd waits for the kubelet to sync a mounted file, typically one to two minutes.

Keep flags OUT of this kind. Declare them as **KubernetesFlagdFlagFile** resources: a flip then edits only the flag file, never re-applies flagd, and can be granted to a team that has no rights over the daemon.

## Conventions and gotchas

- **Readiness waits for every source.** flagd answers `/readyz` only after each source has synced once, and the module's readiness probe is exactly that endpoint. One unreachable URL or missing FeatureFlag keeps every pod unready and the apply waiting until its timeout. A missing ConfigMap is worse: the volume never mounts and the pod never starts. Deploy flag files in the same InfraChart (the reference orders them first) or before the daemon.
- **ConfigMaps mount as directories, never `subPath`.** Each `configMap` source mounts at `/etc/flagd/sources/<index>` and flagd reads `<mount>/<key>`. A `subPath` mount would never see an edit; the directory mount follows the kubelet's atomic volume update, which flagd's file watcher handles. The flip latency is the kubelet sync period, not flagd's.
- **Same namespace only.** ConfigMap sources, `server.tlsSecretName`, the gRPC and OpenTelemetry CA and client certificate Secrets, and `extraEnvFromSecret` all resolve in flagd's own namespace: pod volumes and secretKeyRefs cannot cross namespaces. A KubernetesFlagdFlagFile in another namespace is unreachable from this daemon.
- **The source list is a Secret, and changing it rolls the pods.** The whole SourceConfig array -- including HTTP authorization and credential headers -- renders into `<name>-sources` and reaches flagd as `FLAGD_SOURCES`. A variable read from a Secret is fixed at container start, so the module stamps the document's SHA-256 on the pod template; adding, removing or reordering a source is a rolling restart, while editing a mounted flag file is not.
- **Later sources win.** When two sources define one flag key, the later entry in `sources` wins. A `grpc` source's `selector` (`flagSetId=<id>` or `source=<name>`) narrows what that sync server contributes; the other source types take no selector.
- **FeatureFlag sources need the OpenFeature Operator's CRDs.** The module creates the `<name>-flag-reader` Role (get/list/watch on `featureflags.core.openfeature.dev`) in each namespace those sources read. Kubernetes escalation prevention means the IaC runner must itself hold those verbs to grant them. Two flagds with the same name in different namespaces that read FeatureFlags from one shared namespace would both claim `<name>-flag-reader` there: give them distinct names.
- **An empty CORS list allows every origin.** flagd always installs its CORS handler, and with no `evaluation.corsOrigins` it accepts any origin. List the origins whenever browsers reach flagd.
- **TLS covers evaluation and sync only.** `server.tlsSecretName` serves the evaluation and sync listeners over TLS; the OFREP and management listeners stay plain HTTP, so terminate TLS for OFREP at a gateway.
- **An HTTP source authenticates one way.** `authHeader` or `oauth` client credentials (flagd exchanges them at `tokenUrl` for a token), never both -- validation refuses the pair.
- **The name budget is 63 characters.** The Service is named after the resource, and a Service name is a 63-character DNS label; both engines refuse a longer `metadata.name` before creating anything.

## On the diagram

flagd renders as its own node, and each KubernetesFlagdFlagFile it mounts renders as a separate node wired to it -- the diagram shows which team's flags feed which daemon. Flags embedded in an HTTP or bucket source render as nothing: the diagram cannot show where they come from, which is one more reason to prefer flag files for flags you own.

## Pairs well with

- **KubernetesFlagdFlagFile** -- the typed, plan-time-validated flags this daemon serves.
- **KubernetesNamespace** -- one namespace for the daemon and every flag file it mounts.
- **Gateway API routes** (KubernetesHttpRoute or KubernetesGrpcRoute) -- for OFREP or gRPC evaluation from outside the cluster; set `ofrepSse.publicUrl` to the public origin so clients find the change stream.
- **KubernetesKubePrometheusStack** -- the Prometheus that scrapes the ServiceMonitor `metrics.serviceMonitorEnabled` renders; set `metrics.serviceMonitorLabels` to its serviceMonitorSelector. flagd serves Prometheus metrics on the management port unless `telemetry.metricsExporter` is `otel`.
