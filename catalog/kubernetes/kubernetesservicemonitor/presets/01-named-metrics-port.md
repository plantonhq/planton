# Named Metrics Port

Scrapes one Service's named metrics port every 30 seconds. The series get a `job` label that stays the same in every environment, and one scrape is capped at 50,000 samples. This is the shape almost every service needs: a selector on the Service's own name label, a port referenced by name, and a ceiling that turns a cardinality explosion into one loud failed scrape instead of a Prometheus out of memory.

## When to Use

- A workload already has a Service that names its metrics port (`http-metrics`, `metrics`), and it serves plain HTTP inside the cluster.
- You want the series to carry `job="<service>"`, whatever the Service is called in each environment.

## How It Works

The selector matches the Service by its `app.kubernetes.io/name` label, and `port` names the Service port. A port name survives a renumbering, where a number would silently point at the wrong port. `job_label` copies that same label's value into `job`, so dashboards and alerts written against `job="<service>"` hold across environments whose Service names carry a prefix. `sample_limit` makes a scrape that suddenly returns more than 50,000 samples fail as a whole, which sets `up` to 0 and gets noticed.

## Key Configuration Choices

- **`interval: 30s`.** It is enough for nearly every service; halving it doubles the stored samples.
- **`scrape_timeout: 10s`.** It must not exceed the interval, or the operator skips the monitor.
- **`sample_limit: 50000`.** Size it a few times above the service's normal series count (`scrape_samples_scraped` after the first scrape).

## Prerequisites

- A Service labelled `app.kubernetes.io/name: <service>` with a named metrics port, in the monitor's namespace.
- A Prometheus that selects this object (`KubernetesKubePrometheusStack`; under its default discovery no label is needed).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<service>` | The service's `app.kubernetes.io/name`, also used as the monitor's name. |
| `<namespace>` | Namespace of the monitor and the Service. |
| `<metrics_port_name>` | The Service port's name (`spec.ports[].name`). |
