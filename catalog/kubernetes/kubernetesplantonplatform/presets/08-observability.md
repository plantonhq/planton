# Observability

The zero-config platform you can watch from the same Grafana as everything
else on the cluster. Metrics need no setting: the control plane and the
runner always serve Prometheus text on their Services' port named `metrics`,
inside the cluster only. Traces are the one signal with a setting, because
they need somewhere to go: this preset names the cluster's trace collector
by reference, so every API request is traced and the console's browser
spans join the same traces. Logs are one JSON object per line on stdout,
each carrying its `trace_id`, so a log line opens its trace.

## When to Use

- Any platform a team depends on: you want to know an API call failed, a
  deployment waited too long for its runner, or a job attempt failed, before
  someone reports it
- A cluster that already runs a monitoring stack (a Prometheus and a trace
  store such as an OpenTelemetry collector in front of Tempo), or one you are
  about to give one

## Prerequisites

- A planton-operator chart that knows `observability` (0.27.0 or newer, the
  catalog default); an older definition refuses the declaration
  (`.spec.observability: field not declared in schema`), so upgrade the
  operator first
- A `KubernetesOtelCollector` named `cluster-traces` whose configuration
  declares the standard `otlp` receiver (so it exports
  `otlp_http_endpoint`), and whose network policy, if any, admits the
  platform's namespace on port 4318
- For the metrics: a Prometheus that reads ServiceMonitors, such as the one
  a `KubernetesKubePrometheusStack` installs

## Key Configuration Choices

- **The address is a reference** — `otlp_http_endpoint` follows the
  collector's output, so the address can never drift from the collector's
  Service. A `KubernetesTempo` or `KubernetesSignoz` exports the same output;
  a trace store outside the catalog takes a literal base address
  (`http://…:4318`, no trailing slash and no `/v1/traces`: the platform adds
  the path)
- **Setting it is the switch** — there is no `enabled` flag to disagree with
  the address. Remove the block and tracing is off
- **One monitor per component** — the two components serve different paths.
  Declare a `KubernetesServiceMonitor` for each, selecting
  `app.kubernetes.io/managed-by: planton-operator` and
  `app.kubernetes.io/name: control-plane` (path `/actuator/prometheus`) or
  `app.kubernetes.io/name: runner` (path `/metrics`), port `metrics`, with
  `job_label: app.kubernetes.io/name` so the series read
  `job="control-plane"` and `job="runner"` whatever the platform is named

## What You Will See

- `planton_api_requests_total` by outcome (`ok`, `caller_error`,
  `server_fault`), including the failures the console's gRPC-Web calls carry
  inside an HTTP 200
- `planton_deployment_start_latency_seconds`, a deployment's wait for its
  runner, and how deployments end
- `planton_runner_job_attempts_total`, and the JVM's and Go runtime's own
  series
- In the trace store, every request as a trace whose id its log lines carry,
  with each pipeline step as a span
