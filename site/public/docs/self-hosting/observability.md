---
title: "Watch Your Planton"
description: "See how your self-hosted Planton is doing from your own Grafana: metrics are always served, one setting sends every request's trace to your collector, and every log line opens its trace"
icon: history
order: 30
tags:
  - Self-Hosting
  - Observability
  - Prometheus
  - OpenTelemetry
---

# Watch Your Planton

A platform your team depends on should tell you it is failing before someone reports it. A self-hosted Planton gives your monitoring stack three signals, and two of them need nothing from you:

- **Metrics, always.** The control plane and the runner serve Prometheus metrics on their Services' port named `metrics` (9464), inside the cluster only. No front door routes the port, and no series names a customer.
- **Logs, always.** Every component writes one JSON object per line on stdout, and each control-plane line carries the `trace_id` of the request it belongs to.
- **Traces, with one setting.** Name your trace collector (or Tempo) and every API request is traced there, with the console's browser spans joined to the same traces.

## What the metrics answer

| Series | Question it answers |
|---|---|
| `planton_api_requests_total` | Is the API failing? Every call counted by outcome (`ok`, `caller_error`, `server_fault`), including the failures the console's gRPC-Web calls carry inside an HTTP 200 that no gateway counts |
| `planton_deployment_start_latency_seconds` | Are deployments waiting too long for a runner? |
| `planton_runner_job_attempts_total` | Are the runner's jobs failing? |

The control plane serves them at `/actuator/prometheus`, the runner at `/metrics`, beside the JVM's and the Go runtime's own series.

## Scrape the metrics

Point one ServiceMonitor at each component. The two components serve different paths, so they need a monitor each. `jobLabel` makes the series read `job="control-plane"` and `job="runner"`, whatever you named the platform:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: planton-control-plane
  namespace: planton
spec:
  selector:
    matchLabels:
      app.kubernetes.io/managed-by: planton-operator
      app.kubernetes.io/name: control-plane
  jobLabel: app.kubernetes.io/name
  endpoints:
    - port: metrics
      path: /actuator/prometheus
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: planton-runner
  namespace: planton
spec:
  selector:
    matchLabels:
      app.kubernetes.io/managed-by: planton-operator
      app.kubernetes.io/name: runner
  jobLabel: app.kubernetes.io/name
  endpoints:
    - port: metrics
      path: /metrics
```

If your namespaces run a network policy, admit your Prometheus on port 9464 in the platform's namespace.

## Send the traces

Set the collector's OTLP/HTTP base address on the platform:

```bash
helm upgrade planton oci://ghcr.io/plantonhq/charts/planton --namespace planton \
  --reuse-values \
  --set platform.spec.observability.otlpHttpEndpoint=http://cluster-traces-collector.observability.svc.cluster.local:4318
```

Give the base address only, with no trailing slash and no `/v1/traces`: the platform adds the path itself, and the resource refuses an address that already carries it. Setting the address is the switch, and removing it turns tracing off. If your collector runs behind a network policy, admit the platform's namespace on port 4318, or the spans are dropped without an error.

Through infrastructure as code, the platform kind takes the collector by reference, so the address follows the collector and never drifts:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  createNamespace: true
  version: <release>
  observability:
    otlpHttpEndpoint:
      valueFrom:
        kind: KubernetesOtelCollector
        name: cluster-traces
        fieldPath: status.outputs.otlp_http_endpoint
```

A `KubernetesTempo` or `KubernetesSignoz` exports the same `otlp_http_endpoint` output and works the same way.

## From a log line to its trace

Collect the platform's pod logs into Loki (or any store that keeps a log line's fields), parse each line as JSON, and keep its `trace_id`. A derived field on `trace_id` in your Grafana's Loki datasource then opens the request's trace in Tempo. A failed call's log line leads to the step that failed and its error code.

## Requirements

- The operator chart 0.27.0 or newer. Its definition knows `spec.observability`, and an older one drops the field without a word.
- Platform release v0.0.140 or newer, the operator's floor.
