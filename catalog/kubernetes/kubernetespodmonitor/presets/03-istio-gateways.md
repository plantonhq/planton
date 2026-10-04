# Istio Gateways

Scrapes every Istio-managed Gateway API gateway on the cluster -- each one istiod deploys for a `Gateway` of class `istio`, in whatever namespace the Gateway lives -- with one monitor declared beside the cluster's Prometheus agent. Each series carries the gateway's name as `job`. Only the Istio standard request metrics and Envoy's liveness are kept: "which gateway is failing or slow, for which backend?" is `istio_requests_total` by `response_code` and `destination_service`, and the request duration histogram; TLS passthrough (a tunnel on a TLSRoute) shows as TCP connections.

## When to Use

- The cluster's front doors are Gateway API `Gateway` objects served by Istio (`KubernetesGateway` with the `istio` class), and you want their request rate, errors and latency in Prometheus.
- The gateways are installed by the cluster's own composition, before the agent stack exists, so they cannot carry a monitor of their own: this one, in the agent's composition, reads them all.

## How It Works

istiod creates a Deployment and a Service for each `Gateway`; the Service publishes only the traffic ports, so the gateway's Envoy statistics port (`http-envoy-prom`, 15090, path `/stats/prometheus`) is reached on the pods. The selector matches the label istiod stamps on every gateway pod it deploys, and `namespace_selector: {any: true}` finds them in every namespace, so a gateway added later is scraped with no change here. `job_label` turns the gateway's name label into `job`. Envoy's own statistics run to thousands of series per gateway; the keep relabeling stores only the series that answer an operator's question.

## Key Configuration Choices

- **The keep list.** `istio_requests_total` and `istio_request_duration_milliseconds_*` answer error rate and latency per backend; `istio_tcp_connections_*` cover TLS passthrough, which the HTTP series never see; `envoy_server_live` says the proxy is serving. The two byte-size histograms are left out: they double the series count and answer no common question.
- **`sample_limit: 20000`.** A gateway in front of a few dozen services stays well under it; a limit hit turns the target down loudly instead of flooding storage.
- **No credentials.** The statistics port serves plain HTTP inside the cluster.

## Prerequisites

- Gateways served by Istio (`gateway.networking.k8s.io/gateway-class-name: istio` on their pods).
- A Prometheus that selects this object (`KubernetesKubePrometheusStack`; under its default discovery no label is needed).
- Where a gateway's namespace has a NetworkPolicy, it admits the Prometheus pods on 15090.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<agent_namespace>` | Namespace of the agent stack, where the monitor lives. |
