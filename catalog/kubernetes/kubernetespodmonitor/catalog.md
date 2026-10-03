# Pod Monitor

Declares a prometheus-operator PodMonitor: a namespaced object that tells every Prometheus selecting it which pods to scrape for metrics directly -- on which container port and path, how often, and with which credentials -- without a Service in between. Each running pod matching its label selector becomes a scrape target, one per replica. The spec is faithful to the upstream `monitoring.coreos.com/v1` PodMonitor (pinned to prometheus-operator v0.94.1), after a Planton envelope that places the object and sets its own labels -- this is how you watch database instances, node-level exporters and sidecars that no Service exposes.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **A PodMonitor** -- one namespaced object carrying your selector and scrape endpoints, which the prometheus-operator renders into the scrape configuration of every Prometheus whose PodMonitor selector matches it.
- **Kubernetes Labels** -- your own labels on the object (the ones a Prometheus selects by), with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **prometheus-operator CRDs installed** -- **Kube Prometheus Stack** installs them, along with the Prometheus that scrapes through the monitor.
- **A Prometheus that selects the object** -- the stack's default discovery loads every PodMonitor; a fenced stack loads only objects carrying its `release` label.
- **Pods that declare the metrics port** -- a container port the pod does not declare yields no target.
- **Target namespace exists**, together with any Secret or ConfigMap the monitor reads for credentials.

## Deploy

### Console

Open the deployment store, find **Pod Monitor**, and click **Deploy**. The creation wizard walks you through the namespace, the pods to select, and each scraped endpoint with its port, interval, TLS and credentials.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPodMonitor
metadata:
  name: orders-db
  org: acme-corp
  env: prod
spec:
  namespace:
    value: orders
  job_label: cnpg.io/cluster
  pod_target_labels:
    - cnpg.io/instanceRole
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

The stack's Prometheus picks the monitor up on its next reload and scrapes every instance of the `orders-db` cluster, with `job="orders-db"` and each instance's role on the series. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace and the credentials to resources managed by other Cloud Resources:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: orders-namespace
      fieldPath: spec.name
  selector:
    match_labels:
      app.kubernetes.io/name: exporter
  pod_metrics_endpoints:
    - port: metrics
      authorization:
        credentials:
          name:
            valueFrom:
              kind: KubernetesSecret
              name: exporter-token
          key: token
```

The InfraPipeline creates the namespace and the Secret first, then the monitor. The pods it scrapes and the Prometheus that reads it are chosen by labels, not by references.

## Key Configuration

These are the most important decisions when configuring a pod monitor. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pods, not a Service.** A pod monitor scrapes every matching replica, ready or not, which is what you want for database instances, DaemonSet exporters and sidecars. When a Service already exposes the named metrics port, a **Service Monitor** is the simpler choice.

**The port must be declared.** `port` names a container port, and `port_number` gives its number; a pod that does not declare the port yields no target.

**Labels decide who reads it; labels decide what it scrapes.** The object's `labels` are what a Prometheus's PodMonitor selector matches. The `selector` matches labels on pods; choose labels the workload keeps stable across rollouts.

**One mistake silences the whole object.** The operator skips a monitor whose endpoint has two authentication methods, a client certificate without its key, or a malformed relabeling step, and nothing else says so. The spec refuses those before the apply.

**Credentials are references.** Every token, password, CA and client certificate is read from a Secret or ConfigMap in the monitor's namespace, referenced so the graph creates it first.

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
| `pod_monitor_name` | Name of the created PodMonitor (equals `metadata.name`) | Ordering resources that depend on the scrape being in place |
| `namespace` | The namespace the monitor was created in | Confirming where the monitor lives |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Watch a PostgreSQL cluster's instances** -- every CloudNativePG instance scraped on its `metrics` port, with its role on the series. Start from the **CloudNativePG Instances** preset.

**Scrape a DaemonSet's exporter on every node** -- one target per node, labelled with the node it runs on. Start from the **Per-Node Exporter** preset.

## Works With

- [**Kube Prometheus Stack**](/cloud-catalog/kubernetes-kube-prometheus-stack) -- installs the CRDs and the Prometheus that scrapes through the monitor.
- [**Service Monitor**](/cloud-catalog/kubernetes-service-monitor) -- scrapes through a Service instead.
- [**Prometheus Rule**](/cloud-catalog/kubernetes-prometheus-rule) -- alerts and recording rules over the scraped series.
- [**Kubernetes Secret**](/cloud-catalog/kubernetes-secret) -- holds the scrape credentials the monitor references.
- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) -- the placement target the monitor lives in.
