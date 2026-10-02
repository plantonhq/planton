# Error-Ratio Recording Rules

Precomputes a service's error ratio over the four windows multi-window burn-rate alerting reads: 5 minutes, 30 minutes, 1 hour and 6 hours. Dashboards and alerts then read four cheap series instead of each re-running a rate over hours of raw samples.

## When to Use

- You alert on an error budget (see the **Error-Budget Burn Alerts** preset) and want the alerts to read recorded series.
- A dashboard panel shows the same error ratio as an alert, and the two should never disagree.

## How It Works

One group evaluates the four recording rules every 30 seconds. Each divides the rate of failed requests by the rate of all requests for one job over its window, keeping the `job` label, and writes the result under a `level:metric:operations` name (`job:slo_errors_per_request:ratio_rate1h`), so a reader can tell a recorded series from a scraped one.

## Key Configuration Choices

- **One group, recording rules only.** Keep the alerts that read these series in a later group or a separate object, so an edit to an alert can never break the recording.
- **`interval: 30s`.** The windows are minutes to hours long; evaluating faster only repeats the same answer.
- **Error selector.** Decide what counts as failure for the service (`code=~"5.."` for an HTTP server, `grpc_code!="OK"` for gRPC) and keep it identical in all four rules.

## Prerequisites

- A Prometheus that scrapes the service's request metric and selects this object (`KubernetesKubePrometheusStack`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<service>` | Short name of the service, used in the object and group names. |
| `<namespace>` | Namespace the object lives in. |
| `<requests_metric>` | The service's request counter (for example `http_requests_total`). |
| `<job>` | The `job` label of the service's scrape target. |
| `<error_selector>` | The label matcher that selects failed requests (for example `code=~"5.."`). |
