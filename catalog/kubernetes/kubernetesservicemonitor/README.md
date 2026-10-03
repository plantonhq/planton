# KubernetesServiceMonitor

Declares a prometheus-operator `ServiceMonitor`: a namespaced object telling every Prometheus that selects it which Services' endpoints to scrape for metrics, and how. The spec mirrors the upstream `monitoring.coreos.com/v1` ServiceMonitor field for field (pinned to prometheus-operator v0.94.1, the operator `KubernetesKubePrometheusStack` installs), after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One ServiceMonitor** in the target namespace, carrying the selector and endpoints exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

The prometheus-operator renders the monitor into the scrape configuration of every Prometheus whose ServiceMonitor selector matches it: each Service matching `selector` contributes one target per ready endpoint address, scraped on each entry of `endpoints`. Nothing else is created.

## Which Prometheus Scrapes Through It

A Prometheus picks monitors up by the labels on the object and the namespace it lives in:

- **KubernetesKubePrometheusStack with its default discovery (`all_monitors`)** loads every ServiceMonitor in every namespace. No label is needed.
- **KubernetesKubePrometheusStack with `release_managed_only` discovery** loads only objects labelled `release: <the stack's release_name output>`. Put that label in `spec.labels`.
- **A Prometheus installed any other way** selects by its own `serviceMonitorSelector` and `serviceMonitorNamespaceSelector`; set the labels it names.

## ServiceMonitor or PodMonitor

Scrape through a Service (this kind) when the workload already has a Service that names its metrics port: targets carry the Service's name and labels, and `target_labels` copies Service labels onto the series. Scrape pods directly (**KubernetesPodMonitor**) when no Service names the port (a database operator's instances, a DaemonSet's exporters), or when every replica must be scraped even while it is not ready, since a Service drops unready pods from its endpoints.

## Credentials Are References

Every Secret and ConfigMap an endpoint reads (a bearer token, basic-auth halves, OAuth2 client credentials, a CA, a client certificate and key, proxy headers) is a selector whose `name` defaults to a reference to a **KubernetesSecret** or **KubernetesConfigMap**. The resource graph shows the dependency and creates the Secret first; the object receives the resolved name. The Secret or ConfigMap must live in the monitor's namespace.

## What the Spec Refuses

The operator skips a whole ServiceMonitor, in silence, when an endpoint breaks one of its rules: the apply succeeds and nothing is scraped. The spec refuses those shapes before the apply:

- two authentication methods on one endpoint (`authorization`, `basic_auth`, `oauth2`, `bearer_token_secret`), or an authorization of type `Basic`;
- a client certificate without its key, or a CA, certificate or key taken from both a Secret and a file;
- proxy settings upstream rejects (CONNECT headers without a proxy, the environment's proxy beside an explicit one, a bypass list without a proxy);
- a relabeling step that breaks its action's rules (a `replace` without `target_label`, a `hashmod` without `modulus`, extra fields on `keepequal`, `dropequal`, `labeldrop` and `labelkeep`);
- a `target_port` that is neither a port number nor a port name, and a monitor with no endpoint.

## Prerequisites

- The prometheus-operator CRDs on the cluster. **KubernetesKubePrometheusStack** installs them, and it is this kind's registry prerequisite.
- A Prometheus that selects the object (see above).
- The target namespace (**KubernetesNamespace**), and any Secrets or ConfigMaps the endpoints reference, in that namespace.

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesServiceMonitor
metadata:
  name: api
spec:
  namespace:
    value: api
  selector:
    match_labels:
      app.kubernetes.io/name: api
  endpoints:
    - port: http-metrics
      interval: 30s
```

```shell
planton apply -f service-monitor.yaml
```

## Configuration Reference

### Envelope

| Field | Description |
|-------|-------------|
| `namespace` | Namespace the object is created in, and where Services are searched by default. A literal or a reference to a KubernetesNamespace's `spec.name`. Required. |
| `labels` | The object's own labels. How a Prometheus with a ServiceMonitor selector decides the monitor is its own. |
| `annotations` | The object's own annotations, for people and tools reading it. They do not reach the series. |

### What is scraped

| Field | Description |
|-------|-------------|
| `selector` | Which Services, by label. Required; `{}` selects every Service in the searched namespaces. |
| `namespace_selector` | Which namespaces are searched: `any`, or `match_names` (references to KubernetesNamespace). Unset, the monitor's own. |
| `endpoints` | One entry per scraped port. At least one. |
| `selector_mechanism` | `RelabelConfig` (default) or `RoleSelector` (selector passed to the API server). |
| `service_discovery_role` | `Endpoints` or `EndpointSlice`. Unset, the Prometheus's own setting. |
| `attach_metadata.node` | Attach the node's metadata as discovery labels. |
| `scrape_class` | The Prometheus scrape class whose defaults the monitor inherits. |

### Series labels and limits

| Field | Description |
|-------|-------------|
| `job_label` | The Service label whose value becomes `job`. Unset, the Service's name. |
| `target_labels` / `pod_target_labels` | Service and pod labels copied onto every series. |
| `sample_limit`, `target_limit`, `label_limit`, `label_name_length_limit`, `label_value_length_limit`, `keep_dropped_targets` | Per-scrape guards against a target that floods the Prometheus. |
| `body_size_limit` | Largest uncompressed response accepted ("10MiB"). |
| `scrape_protocols`, `fallback_scrape_protocol` | Exposition formats offered, and the one assumed for a bad Content-Type. |
| `scrape_native_histograms`, `scrape_classic_histograms`, `native_histogram_bucket_limit`, `native_histogram_min_bucket_factor`, `convert_classic_histograms_to_nhcb` | Native-histogram handling. |

### Endpoints (`endpoints[]`)

| Field | Description |
|-------|-------------|
| `port` | The Service port's name. Takes precedence over `target_port`. |
| `target_port` | The container port, by number ("9090") or name; a number reaches the object as a number. |
| `path`, `scheme`, `params` | The URL scraped. `params` values are lists: `{module: {values: [http_2xx]}}`. |
| `interval`, `scrape_timeout` | Scrape cadence and timeout, as Prometheus durations; the timeout may not exceed the interval. |
| `tls_config` | TLS to the target: CA, client certificate and key from Secrets and ConfigMaps (or files in the Prometheus container), server name, versions. |
| `authorization`, `basic_auth`, `oauth2`, `bearer_token_secret`, `bearer_token_file` | At most one authentication method; every credential is a Secret reference. The last two are deprecated upstream. |
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
      name: api-namespace
      fieldPath: spec.name
  endpoints:
    - port: http-metrics
      authorization:
        credentials:
          name:
            valueFrom:
              kind: KubernetesSecret
              name: api-scrape-token
          key: token
```

The Services a monitor scrapes are matched by label at discovery time, not referenced, so the diagram draws no edge to them. The registry prerequisite orders the monitor after the KubernetesKubePrometheusStack that installs its CRDs.

## Outputs

| Output | Description |
|--------|-------------|
| `service_monitor_name` | Name of the created ServiceMonitor (equals `metadata.name`). |
| `namespace` | Namespace the ServiceMonitor was created in. |

## Related Kinds

- **KubernetesKubePrometheusStack**: installs the CRDs and the Prometheus that scrapes through the monitor.
- **KubernetesPodMonitor**: scrapes pods directly, without a Service.
- **KubernetesPrometheusRule**: the alerting and recording rules over what is scraped.
- **KubernetesNamespace**, **KubernetesSecret**, **KubernetesConfigMap**: where the monitor lives and what it reads.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
