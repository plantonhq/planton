# KubernetesPodMonitor

Declares a prometheus-operator `PodMonitor`: a namespaced object telling every Prometheus that selects it which pods to scrape for metrics directly, and how. The spec mirrors the upstream `monitoring.coreos.com/v1` PodMonitor field for field (pinned to prometheus-operator v0.94.1, the operator `KubernetesKubePrometheusStack` installs), after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One PodMonitor** in the target namespace, carrying the selector and endpoints exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

The prometheus-operator renders the monitor into the scrape configuration of every Prometheus whose PodMonitor selector matches it: each running pod matching `selector` contributes one target per declared container port each entry of `pod_metrics_endpoints` selects. Nothing else is created.

## Which Prometheus Scrapes Through It

A Prometheus picks monitors up by the labels on the object and the namespace it lives in:

- **KubernetesKubePrometheusStack with its default discovery (`all_monitors`)** loads every PodMonitor in every namespace. No label is needed.
- **KubernetesKubePrometheusStack with `release_managed_only` discovery** loads only objects labelled `release: <the stack's release_name output>`. Put that label in `spec.labels`.
- **A Prometheus installed any other way** selects by its own `podMonitorSelector` and `podMonitorNamespaceSelector`; set the labels it names.

## PodMonitor or ServiceMonitor

Scrape pods directly (this kind) when no Service names the metrics port (a database operator's instances, a DaemonSet's exporters, a sidecar's metrics port), or when every replica must be scraped even while it is not ready, since a Service drops unready pods from its endpoints. Scrape through a Service (**KubernetesServiceMonitor**) when one already exists and its labels are what the series should carry.

## Credentials Are References

Every Secret and ConfigMap an endpoint reads (a bearer token, basic-auth halves, OAuth2 client credentials, a CA, a client certificate and key, proxy headers) is a selector whose `name` defaults to a reference to a **KubernetesSecret** or **KubernetesConfigMap**. The resource graph shows the dependency and creates the Secret first; the object receives the resolved name. The Secret or ConfigMap must live in the monitor's namespace.

## What the Spec Refuses

The operator skips a whole PodMonitor, in silence, when an endpoint breaks one of its rules: the apply succeeds and nothing is scraped. The spec refuses those shapes before the apply:

- two authentication methods on one endpoint (`authorization`, `basic_auth`, `oauth2`, `bearer_token_secret`), or an authorization of type `Basic`;
- a client certificate without its key, or a value taken from both a Secret and a ConfigMap;
- proxy settings upstream rejects (CONNECT headers without a proxy, the environment's proxy beside an explicit one, a bypass list without a proxy);
- a relabeling step that breaks its action's rules (a `replace` without `target_label`, a `hashmod` without `modulus`, extra fields on `keepequal`, `dropequal`, `labeldrop` and `labelkeep`);
- a `port_number` outside 1-65535, and a `target_port` that is neither a port number nor a port name.

## Prerequisites

- The prometheus-operator CRDs on the cluster. **KubernetesKubePrometheusStack** installs them, and it is this kind's registry prerequisite.
- A Prometheus that selects the object (see above).
- The target namespace (**KubernetesNamespace**), and any Secrets or ConfigMaps the endpoints reference, in that namespace.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPodMonitor
metadata:
  name: orders-db
spec:
  namespace:
    value: orders
  selector:
    match_labels:
      cnpg.io/cluster: orders-db
  pod_metrics_endpoints:
    - port: metrics
      interval: 30s
```

```shell
planton apply -f pod-monitor.yaml
```

## Configuration Reference

### Envelope

| Field | Description |
|-------|-------------|
| `namespace` | Namespace the object is created in, and where pods are searched by default. A literal or a reference to a KubernetesNamespace's `spec.name`. Required. |
| `labels` | The object's own labels. How a Prometheus with a PodMonitor selector decides the monitor is its own. |
| `annotations` | The object's own annotations, for people and tools reading it. They do not reach the series. |

### What is scraped

| Field | Description |
|-------|-------------|
| `selector` | Which pods, by label. Required; `{}` selects every pod in the searched namespaces. |
| `namespace_selector` | Which namespaces are searched: `any`, or `match_names` (references to KubernetesNamespace). Unset, the monitor's own. |
| `pod_metrics_endpoints` | One entry per scraped container port. |
| `selector_mechanism` | `RelabelConfig` (default) or `RoleSelector` (selector passed to the API server). |
| `attach_metadata.node` | Attach the node's metadata as discovery labels. |
| `scrape_class` | The Prometheus scrape class whose defaults the monitor inherits. |

### Series labels and limits

| Field | Description |
|-------|-------------|
| `job_label` | The pod label whose value becomes `job`. Unset, `<monitor namespace>/<monitor name>`. |
| `pod_target_labels` | Pod labels copied onto every series. |
| `sample_limit`, `target_limit`, `label_limit`, `label_name_length_limit`, `label_value_length_limit`, `keep_dropped_targets` | Per-scrape guards against a target that floods the Prometheus. |
| `body_size_limit` | Largest uncompressed response accepted ("10MiB"). |
| `scrape_protocols`, `fallback_scrape_protocol` | Exposition formats offered, and the one assumed for a bad Content-Type. |
| `scrape_native_histograms`, `scrape_classic_histograms`, `native_histogram_bucket_limit`, `native_histogram_min_bucket_factor`, `convert_classic_histograms_to_nhcb` | Native-histogram handling. |

### Endpoints (`pod_metrics_endpoints[]`)

| Field | Description |
|-------|-------------|
| `port` | The container port's name. Takes precedence over `port_number` and `target_port`. |
| `port_number` | The container port's number; the pod must declare it. |
| `target_port` | Deprecated upstream: the container port by number ("9187") or name; a number reaches the object as a number. |
| `path`, `scheme`, `params` | The URL scraped. `params` values are lists: `{module: {values: [http_2xx]}}`. |
| `interval`, `scrape_timeout` | Scrape cadence and timeout, as Prometheus durations; the timeout may not exceed the interval. |
| `tls_config` | TLS to the target: CA, client certificate and key from Secrets and ConfigMaps, server name, versions. |
| `authorization`, `basic_auth`, `oauth2`, `bearer_token_secret` | At most one authentication method; every credential is a Secret reference. `bearer_token_secret` is deprecated upstream. |
| `relabelings`, `metric_relabelings` | Relabeling before the scrape (targets) and before storage (samples). |
| `honor_labels`, `honor_timestamps`, `track_timestamps_staleness` | Label and timestamp conflict handling. |
| `proxy_url`, `no_proxy`, `proxy_from_environment`, `proxy_connect_header` | The proxy scrapes go through. |
| `follow_redirects`, `enable_http2`, `filter_running` | HTTP client and discovery switches. |

## Composing in Infra Charts

Reference the namespace from a KubernetesNamespace and the token from a KubernetesSecret in the same chart; the pipeline creates both first:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: orders-namespace
      fieldPath: spec.name
  pod_metrics_endpoints:
    - port: metrics
      basic_auth:
        username:
          name:
            valueFrom:
              kind: KubernetesSecret
              name: exporter-auth
          key: username
        password:
          name:
            valueFrom:
              kind: KubernetesSecret
              name: exporter-auth
          key: password
```

The pods a monitor scrapes are matched by label at discovery time, not referenced, so the diagram draws no edge to them. The registry prerequisite orders the monitor after the KubernetesKubePrometheusStack that installs its CRDs.

## Outputs

| Output | Description |
|--------|-------------|
| `pod_monitor_name` | Name of the created PodMonitor (equals `metadata.name`). |
| `namespace` | Namespace the PodMonitor was created in. |

## Related Kinds

- **KubernetesKubePrometheusStack**: installs the CRDs and the Prometheus that scrapes through the monitor.
- **KubernetesServiceMonitor**: scrapes through a Service's endpoints.
- **KubernetesPrometheusRule**: the alerting and recording rules over what is scraped.
- **KubernetesNamespace**, **KubernetesSecret**, **KubernetesConfigMap**: where the monitor lives and what it reads.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
