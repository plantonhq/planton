# Per-Node Exporter

Scrapes an exporter that runs as a DaemonSet, one target per node, and labels every series with the node it came from. Use it for node-level agents (a GPU exporter, a log shipper's own metrics, a network or storage driver) that expose metrics on a container port without a Service.

## When to Use

- An exporter runs on every node as a DaemonSet and declares a named metrics port on its container.
- Queries need to tell nodes apart ("which node's GPU is hot", "which node's driver is erroring").

## How It Works

The selector matches the DaemonSet's pods by `app.kubernetes.io/name`. Each pod is a target, so each node is one target. The first relabeling copies the pod's node name (a discovery label every pod target carries) into `node`. The second copies the pods' `app.kubernetes.io/component` label onto the series, so one exporter shipping several components stays separable. Without the node label, the series from different nodes differ only by pod name, which changes on every rollout.

## Key Configuration Choices

- **A port name, not a number.** The DaemonSet's container must declare the port; a pod that does not yields no target.
- **`node` from discovery.** It survives rollouts, unlike the pod name, so dashboards and alerts can key on it.
- **No `filter_running` change.** A pod stuck in a failed phase is dropped from targets by default, which keeps a crashed node agent from reading as a permanently down target; its absence is what an alert on `up` per node should catch.

## Prerequisites

- A DaemonSet whose pods are labelled `app.kubernetes.io/name: <exporter>` and declare `<metrics_port_name>`.
- A Prometheus that selects this object (`KubernetesKubePrometheusStack`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<exporter>` | The exporter's `app.kubernetes.io/name`, also used as the monitor's name. |
| `<namespace>` | Namespace of the DaemonSet and the monitor. |
| `<metrics_port_name>` | The container port's name the exporter declares. |
