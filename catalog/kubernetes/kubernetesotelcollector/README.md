# Kubernetes OTel Collector

## When NOT to Use This

**One resource is ONE OpenTelemetry Collector** — the declaration of
the `OpenTelemetryCollector` CR (`opentelemetry.io/v1beta1`) that the
OpenTelemetry Operator reconciles into the collector workload, its
Services, and the rendered config ConfigMap.

Not the right kind when:

- **The operator is missing** — a `KubernetesOtelOperator` on the
  cluster is the PREREQUISITE (it watches every namespace; one install
  serves the whole cluster). Nothing reconciles this declaration
  without it.
- **You want to STORE telemetry** — the collector receives, processes
  and exports; it keeps nothing. Backends are their own kinds: a
  `KubernetesLoki` for logs, Tempo for traces — this kind is how
  telemetry reaches them.
- **You want to READ telemetry in a UI** — that is Grafana, pointed at
  the backends this collector ships to.

## The pipeline is the product

`config_yaml` carries the collector's own configuration document —
receivers, processors, exporters, connectors, extensions and the
service pipelines wiring them together — on the collector's OWN open
contract. The OpenTelemetry component registry is unbounded by design,
so this is the upstream grain, not an escape hatch. The operator's
admission webhook validates the document's shape at apply time, and
both modules parse it up front — an unparseable document fails before
anything applies. The presets carry the high-demand pipelines ready to
remix: cluster logs → Loki (daemonset mode), traces → Tempo, and OTLP
fan-in → both.

## Four modes, one field

- **`deployment`** (the default) — the scalable gateway/fan-in shape:
  applications push to it, it fans out to backends. Takes `replicas`
  or the operator-managed `autoscaler`.
- **`daemonset`** — one collector per node, for log files and
  host/kubelet metrics that only exist node-locally. Pair with
  `volumes` hostPath mounts to reach /var/log.
- **`statefulset`** — stable identities, for the target allocator and
  persistent sending queues.
- **`sidecar`** — no standalone workload: the operator injects this
  collector into pods annotated `sidecar.opentelemetry.io/inject`.

Scaling fields (`replicas`, `autoscaler`) apply only to
deployment/statefulset — a daemonset runs one per node and a sidecar
runs inside the target pods. Sidecar mode also rejects tolerations and
`priority_class_name`: the collector runs inside the TARGET pods,
whose scheduling this CR does not control. Both rules are validation
errors at apply time, mirrored from the CRD's own admission rules.

## Credentials never ride the config

Never inline tokens in `config_yaml`. Load existing Secrets whole as
environment variables (`env_from_secrets` — each key becomes a
variable) and reference them in the config as `${env:VAR_NAME}` — the
collector expands them at start, so nothing secret-bearing ever lands
in the rendered ConfigMap.

## The operator fills in what you leave unsaid

Leave `image` empty and the operator injects its default collector
image (the opentelemetry-collector-k8s distribution at the operator's
paired version — override fleet-wide via the operator kind's
`default_collector_image`). The operator also derives the collector
Service's ports from the declared receivers — it knows the standard
components' ports (OTLP, jaeger, zipkin, prometheus), so
`additional_ports` exists only for receivers it cannot infer.

## Receivers that read cluster state need RBAC

Receivers like k8s_events, kubeletstats, k8s_cluster and filelog with
Kubernetes enrichment need permissions beyond the operator's default
ServiceAccount — compose a `KubernetesServiceAccount` +
`KubernetesRbac` and set `service_account` to the composed account.

## The 42-character name budget

Keep this resource's name at 42 characters or fewer: the operator
derives child names by suffixing (`-collector-monitoring` is the
longest at 21) and Kubernetes caps names at 63. Both modules fail
loudly past the budget.

## The outputs contract

The exported `otlp_grpc_endpoint` (`<service>:4317`) and
`otlp_http_endpoint` (`http://<service>:4318`) are valid when the
config declares the standard `otlp` receiver — every preset does.
Sidecar mode creates no standalone workload, so every Service-derived
handle is empty there.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
