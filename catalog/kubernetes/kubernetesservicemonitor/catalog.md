# Service Monitor

Declares a prometheus-operator ServiceMonitor: a namespaced object that tells every Prometheus selecting it which Services to scrape for metrics, on which port and path, how often, and with which credentials. Each Service matching its label selector becomes a set of scrape targets, one per ready endpoint. The spec is faithful to the upstream `monitoring.coreos.com/v1` ServiceMonitor (pinned to prometheus-operator v0.94.1), after a Planton envelope that places the object and sets its own labels -- this is how you declare that a workload is watched, next to the workload itself.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **A ServiceMonitor** -- one namespaced object carrying your selector and scrape endpoints, which the prometheus-operator renders into the scrape configuration of every Prometheus whose ServiceMonitor selector matches it.
- **Kubernetes Labels** -- your own labels on the object (the ones a Prometheus selects by), with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **prometheus-operator CRDs installed** -- **Kube Prometheus Stack** installs them, along with the Prometheus that scrapes through the monitor.
- **A Prometheus that selects the object** -- the stack's default discovery loads every ServiceMonitor; a fenced stack loads only objects carrying its `release` label.
- **A Service that names the metrics port** -- the monitor scrapes through Services; for pods without one, use **Pod Monitor**.
- **Target namespace exists**, together with any Secret or ConfigMap the monitor reads for credentials.

## Deploy

### Console

Open the deployment store, find **Service Monitor**, and click **Deploy**. The creation wizard walks you through the namespace, the Services to select, and each scraped endpoint with its port, interval, TLS and credentials.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesServiceMonitor
metadata:
  name: api
  org: acme-corp
  env: prod
spec:
  namespace:
    value: api
  job_label: app.kubernetes.io/name
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

The stack's Prometheus picks the monitor up on its next reload and starts scraping every endpoint of the `api` Service, with `job="api"` on the series. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace and the scrape token to resources managed by other Cloud Resources:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: api-namespace
      fieldPath: spec.name
  selector:
    match_labels:
      app.kubernetes.io/name: api
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

The InfraPipeline creates the namespace and the Secret first, then the monitor. The Services it scrapes and the Prometheus that reads it are chosen by labels, not by references.

## Key Configuration

These are the most important decisions when configuring a service monitor. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Labels decide who reads it; labels decide what it scrapes.** The object's `labels` are what a Prometheus's ServiceMonitor selector matches (no label is needed under the stack's default discovery). The `selector` matches labels on Services, not on pods; copy it from the Service, not from the Deployment.

**Name the port.** `port` is the Service port's name, which survives a renumbering. `target_port` takes the container port by number or name for a Service that does not name it.

**One mistake silences the whole object.** The operator skips a monitor whose endpoint has two authentication methods, a client certificate without its key, or a malformed relabeling step, and nothing else says so. The spec refuses those before the apply; a missing Secret is caught only by the operator.

**Credentials are references.** Every token, password, CA and client certificate is read from a Secret or ConfigMap in the monitor's namespace, referenced so the graph creates it first. Prefer `authorization` or `oauth2` over `basic_auth`.

**Bound the cost.** `sample_limit` fails a scrape that suddenly returns too many samples, and `metric_relabelings` drop series you never read before they are stored.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace`, `namespace_selector.match_names` | `spec.name` |
| **KubernetesSecret** | every Secret selector's `name` (`authorization.credentials`, `basic_auth`, `oauth2.client_secret`, `tls_config.key_secret`, ...) | `status.outputs.secret_name` |
| **KubernetesConfigMap** | every ConfigMap selector's `name` (`tls_config.ca.config_map`, `oauth2.client_id.config_map`, ...) | `status.outputs.configmap_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `service_monitor_name` | Name of the created ServiceMonitor (equals `metadata.name`) | Ordering resources that depend on the scrape being in place |
| `namespace` | The namespace the monitor was created in | Confirming where the monitor lives |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Scrape a service's metrics port** -- one named port, a stable `job` label, and a cap on samples per scrape. Start from the **Named Metrics Port** preset.

**Scrape an endpoint that wants a token** -- a bearer token read from a Secret, over HTTPS with the issuing CA. Start from the **Token-Protected Endpoint** preset.

## Works With

- [**Kube Prometheus Stack**](/cloud-catalog/kubernetes-kube-prometheus-stack) -- installs the CRDs and the Prometheus that scrapes through the monitor.
- [**Pod Monitor**](/cloud-catalog/kubernetes-pod-monitor) -- scrapes pods directly, without a Service.
- [**Prometheus Rule**](/cloud-catalog/kubernetes-prometheus-rule) -- alerts and recording rules over the scraped series.
- [**Kubernetes Secret**](/cloud-catalog/kubernetes-secret) -- holds the scrape credentials the monitor references.
- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) -- the placement target the monitor lives in.
